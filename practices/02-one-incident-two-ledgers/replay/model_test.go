package replay

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const (
	contractPath = "../contract.json"
	evidencePath = "../evidence.jsonl"
	starterPath  = "../starter/analysis.json"
	solutionPath = "../solution/analysis.json"
)

func TestEvidenceMatchesDeclaredReplay(t *testing.T) {
	contract := mustLoadContract(t)
	events := mustLoadEvidence(t)
	want := GenerateStartEvidence(contract)
	if !reflect.DeepEqual(events, want) {
		t.Fatal("evidence.jsonl differs from the declared deterministic replay")
	}
}

func TestStartLedgersUseDifferentPopulations(t *testing.T) {
	contract := mustLoadContract(t)
	derived, violations := Derive(contract, mustLoadEvidence(t))
	if len(violations) != 0 {
		t.Fatalf("derive violations = %v", violations)
	}
	if derived.LoadLedger.EligibleLogicalOperations != 100 ||
		derived.LoadLedger.CheckoutAttemptsForEligibleOperations != 120 ||
		derived.LoadLedger.PreWindowQueuedAttempts != 5 {
		t.Fatalf("load ledger = %+v", derived.LoadLedger)
	}
	if derived.ProductLedger.GoodLogicalOperations != 95 ||
		derived.ProductLedger.BadLogicalOperations != 5 ||
		derived.ProductLedger.MaxFirstResultLatencyMS != 240 {
		t.Fatalf("product ledger = %+v", derived.ProductLedger)
	}
	late := SortedLateOperations(contract, mustLoadEvidence(t))
	wantLate := []string{"C-096", "C-097", "C-098", "C-099", "C-100"}
	if !reflect.DeepEqual(late, wantLate) {
		t.Fatalf("late operations = %v, want %v", late, wantLate)
	}
}

func TestStarterAnalysisFails(t *testing.T) {
	evaluation := mustEvaluate(t, starterPath)
	if len(evaluation.Violations) == 0 {
		t.Fatal("starter analysis unexpectedly passed")
	}
	if !containsViolation(evaluation.Violations, "wrong population") {
		t.Fatalf("violations = %v, want operation-population rejection", evaluation.Violations)
	}
}

func TestSolutionAnalysisPasses(t *testing.T) {
	evaluation := mustEvaluate(t, solutionPath)
	if len(evaluation.Violations) != 0 {
		t.Fatalf("solution violations = %v", evaluation.Violations)
	}
	if evaluation.CorrectedPrediction.CheckoutAttempts != 100 ||
		evaluation.CorrectedPrediction.GoodLogicalOperations != 100 ||
		evaluation.CorrectedPrediction.MaxFirstResultLatencyMS != 30 ||
		evaluation.CorrectedPrediction.BrowsePermitsRetained != 10 ||
		evaluation.CorrectedPrediction.BrowseAttemptsWaiting != 1 ||
		!evaluation.CorrectedPrediction.ProducerClaimPaused {
		t.Fatalf("corrected prediction = %+v", evaluation.CorrectedPrediction)
	}
}

func TestOracleRejectsAlteredEvidence(t *testing.T) {
	events := mustLoadEvidence(t)
	events = events[:len(events)-1]
	evaluation := Evaluate(mustLoadAnalysis(t, solutionPath), mustLoadContract(t), events)
	if !containsViolation(evaluation.Violations, "raw evidence differs") {
		t.Fatalf("violations = %v, want altered-evidence rejection", evaluation.Violations)
	}
}

func TestLoadAnalysisRejectsTrailingJSON(t *testing.T) {
	data, err := os.ReadFile(solutionPath)
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, []byte("\n{\"future_slo_guaranteed\":true}\n")...)
	path := filepath.Join(t.TempDir(), "analysis.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadAnalysis(path); err == nil {
		t.Fatal("analysis with a trailing JSON value unexpectedly passed")
	}
}

func TestLoadContractRejectsTrailingJSON(t *testing.T) {
	data, err := os.ReadFile(contractPath)
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, []byte("\n{\"good_deadline_ms\":999}\n")...)
	path := filepath.Join(t.TempDir(), "contract.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadContract(path); err == nil {
		t.Fatal("contract with a trailing JSON value unexpectedly passed")
	}
}

func TestLoadEvidenceRejectsUnknownField(t *testing.T) {
	line := "{\"sequence\":1,\"at_ms\":0,\"kind\":\"result_emitted\",\"contract_valid\":true,\"future_slo_guaranteed\":true}\n"
	path := filepath.Join(t.TempDir(), "evidence.jsonl")
	if err := os.WriteFile(path, []byte(line), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadEvidence(path); err == nil {
		t.Fatal("evidence with an unknown field unexpectedly passed")
	}
}

func TestLoadEvidenceRejectsTrailingJSONOnOneLine(t *testing.T) {
	line := "{\"sequence\":1,\"at_ms\":0,\"kind\":\"result_emitted\",\"contract_valid\":true} {\"sequence\":2}\n"
	path := filepath.Join(t.TempDir(), "evidence.jsonl")
	if err := os.WriteFile(path, []byte(line), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadEvidence(path); err == nil {
		t.Fatal("evidence line with a trailing JSON value unexpectedly passed")
	}
}

func TestInvalidResultDoesNotQualifyOperation(t *testing.T) {
	contract := mustLoadContract(t)
	events := mustLoadEvidence(t)
	changed := false
	for index := range events {
		if events[index].Kind == KindResultEmitted && events[index].OperationID == "C-001" {
			events[index].ContractValid = false
			changed = true
		}
	}
	if !changed {
		t.Fatal("C-001 has no emitted-result event")
	}
	derived, violations := Derive(contract, events)
	if !containsViolation(violations, "raw evidence differs") {
		t.Fatalf("violations = %v, want altered-evidence rejection", violations)
	}
	if derived.ProductLedger.GoodLogicalOperations != 94 || derived.ProductLedger.BadLogicalOperations != 6 {
		t.Fatalf("product ledger = %+v, want 94 good and 6 bad", derived.ProductLedger)
	}
	if !containsString(SortedLateOperations(contract, events), "C-001") {
		t.Fatalf("late operations omit C-001")
	}
}

func TestDeriveScenarioMarksOperationWithoutValidResultBad(t *testing.T) {
	contract := mustLoadContract(t)
	events, _ := simulateCheckout(contract, 2)
	changed := false
	for index := range events {
		if events[index].Kind == KindResultEmitted && events[index].OperationID == "C-001" {
			events[index].ContractValid = false
			changed = true
		}
	}
	if !changed {
		t.Fatal("C-001 has no emitted-result event")
	}
	derived, violations := DeriveScenario(contract, events)
	if len(violations) != 0 {
		t.Fatalf("derive scenario violations = %v", violations)
	}
	if derived.ProductLedger.GoodLogicalOperations != 99 || derived.ProductLedger.BadLogicalOperations != 1 {
		t.Fatalf("product ledger = %+v, want 99 good and 1 bad", derived.ProductLedger)
	}
}

func TestVerifierAcceptsAnotherPassingAllocation(t *testing.T) {
	analysis := mustLoadAnalysis(t, solutionPath)
	analysis.Allocation = Allocation{BrowsePermits: 9, CheckoutPermits: 3}
	contract := mustLoadContract(t)
	analysis.CorrectedPrediction = PredictScenario(contract, analysis.Allocation)
	analysis.BoundaryPrediction = PredictBoundary(contract, analysis.Allocation)
	evaluation := Evaluate(analysis, contract, mustLoadEvidence(t))
	if len(evaluation.Violations) != 0 {
		t.Fatalf("9/3 allocation violations = %v", evaluation.Violations)
	}
}

func TestOneCheckoutPermitDoesNotSatisfyTheObjective(t *testing.T) {
	analysis := mustLoadAnalysis(t, solutionPath)
	analysis.Allocation = Allocation{BrowsePermits: 11, CheckoutPermits: 1}
	contract := mustLoadContract(t)
	analysis.CorrectedPrediction = PredictScenario(contract, analysis.Allocation)
	analysis.BoundaryPrediction = PredictBoundary(contract, analysis.Allocation)
	evaluation := Evaluate(analysis, contract, mustLoadEvidence(t))
	if !containsViolation(evaluation.Violations, "does not satisfy the declared W-2 objective") {
		t.Fatalf("violations = %v, want insufficient-allocation rejection", evaluation.Violations)
	}
}

func TestBoundaryIsolationLosesFungibility(t *testing.T) {
	contract := mustLoadContract(t)
	prediction := PredictBoundary(contract, Allocation{BrowsePermits: 10, CheckoutPermits: 2})
	if !prediction.DependencyXCapable || prediction.BrowseLatencyMS != 50 ||
		prediction.SharedPoolBrowseAdmitted != 12 || prediction.IsolatedBrowseAdmitted != 10 ||
		prediction.IsolatedBrowseWaiting != 2 || prediction.IdleCheckoutPermits != 2 || !prediction.FungibilityLost {
		t.Fatalf("boundary prediction = %+v", prediction)
	}
}

func TestBoundaryPassingWindowDoesNotProveFutureSLO(t *testing.T) {
	analysis := mustLoadAnalysis(t, solutionPath)
	analysis.Claims.FutureSLOGuaranteed = true
	evaluation := Evaluate(analysis, mustLoadContract(t), mustLoadEvidence(t))
	if !containsViolation(evaluation.Violations, "does not guarantee future SLO") {
		t.Fatalf("violations = %v, want stronger-claim rejection", evaluation.Violations)
	}
}

func TestBoundaryProductDoesNotRequireLongTransaction(t *testing.T) {
	analysis := mustLoadAnalysis(t, solutionPath)
	analysis.ResourceRetention.ProductRequiresLongTransaction = true
	evaluation := Evaluate(analysis, mustLoadContract(t), mustLoadEvidence(t))
	if !containsViolation(evaluation.Violations, "not one long transaction") {
		t.Fatalf("violations = %v, want implementation/product boundary rejection", evaluation.Violations)
	}
}

func TestBoundaryFullQueueDoesNotEstablishBackpressure(t *testing.T) {
	analysis := mustLoadAnalysis(t, solutionPath)
	analysis.Backpressure.WaitingEvidenceKind = "queue_full"
	analysis.Backpressure.ProducerEvidenceKind = "queue_depth"
	analysis.Backpressure.FullQueueAloneIsBackpressure = true
	evaluation := Evaluate(analysis, mustLoadContract(t), mustLoadEvidence(t))
	if !containsViolation(evaluation.Violations, "wrong raw evidence kinds") ||
		!containsViolation(evaluation.Violations, "full queue alone") {
		t.Fatalf("violations = %v, want backpressure evidence rejection", evaluation.Violations)
	}
}

func mustEvaluate(t *testing.T, analysisPath string) Evaluation {
	t.Helper()
	return Evaluate(mustLoadAnalysis(t, analysisPath), mustLoadContract(t), mustLoadEvidence(t))
}

func mustLoadContract(t *testing.T) Contract {
	t.Helper()
	contract, err := LoadContract(contractPath)
	if err != nil {
		t.Fatal(err)
	}
	return contract
}

func mustLoadEvidence(t *testing.T) []Event {
	t.Helper()
	events, err := LoadEvidence(evidencePath)
	if err != nil {
		t.Fatal(err)
	}
	return events
}

func mustLoadAnalysis(t *testing.T, path string) Analysis {
	t.Helper()
	analysis, err := LoadAnalysis(path)
	if err != nil {
		t.Fatal(err)
	}
	return analysis
}

func containsViolation(violations []string, fragment string) bool {
	for _, violation := range violations {
		if strings.Contains(violation, fragment) {
			return true
		}
	}
	return false
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
