package replay

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	contractPath  = "../contract.json"
	historiesPath = "../histories.json"
	starterPath   = "../starter/review.json"
	solutionPath  = "../solution/review.json"
)

func TestStarterReviewFails(t *testing.T) {
	evaluation := mustEvaluate(t, starterPath)
	if len(evaluation.Violations) == 0 {
		t.Fatal("starter review unexpectedly passed")
	}
	if !containsViolation(evaluation.Violations, "required fact W-1 is missing") {
		t.Fatalf("violations = %v, want same-session requirement rejection", evaluation.Violations)
	}
}

func TestSolutionReviewPasses(t *testing.T) {
	evaluation := mustEvaluate(t, solutionPath)
	if len(evaluation.Violations) != 0 {
		t.Fatalf("solution violations = %v", evaluation.Violations)
	}
	if evaluation.VisibilitySatisfies != 1 || evaluation.VisibilityViolates != 3 ||
		evaluation.ObjectSatisfies != 2 || evaluation.ObjectViolates != 2 {
		t.Fatalf("evaluation counts = %+v", evaluation)
	}
}

func TestV3ShorterResultUsesAdvancedState(t *testing.T) {
	contract := visibilityContractByID(t, mustLoadContract(t), "C-V3")
	history := visibilityHistoryByID(t, mustLoadCorpus(t), "V-3")
	truth, violations := deriveVisibility(history, contract)
	if len(violations) != 0 {
		t.Fatalf("derive violations = %v", violations)
	}
	if truth.Verdict != VerdictSatisfies || len(truth.MissingFacts) != 0 {
		t.Fatalf("truth = %+v, want satisfying shorter result", truth)
	}
	if len(truth.Changes) != 1 || truth.Changes[0].FactID != "N-12" || truth.Changes[0].AffectsFactID != "N-11" {
		t.Fatalf("changes = %+v", truth.Changes)
	}
}

func TestV4RequiresSemanticContribution(t *testing.T) {
	contract := visibilityContractByID(t, mustLoadContract(t), "C-V4")
	history := visibilityHistoryByID(t, mustLoadCorpus(t), "V-4")
	truth, violations := deriveVisibility(history, contract)
	if len(violations) != 0 {
		t.Fatalf("derive violations = %v", violations)
	}
	if len(truth.RequiredFacts) != 1 || truth.RequiredFacts[0].FactID != "X" ||
		truth.RequiredFacts[0].Basis != "consumed_dependency_input" {
		t.Fatalf("required facts = %+v", truth.RequiredFacts)
	}
	if len(truth.NonDependencies) != 1 || truth.NonDependencies[0].FactID != "U" {
		t.Fatalf("non-dependencies = %+v", truth.NonDependencies)
	}
}

func TestConnectedTelemetryDoesNotProveDependency(t *testing.T) {
	contract := visibilityContractByID(t, mustLoadContract(t), "C-V4")
	history := visibilityHistoryByID(t, mustLoadCorpus(t), "V-4")
	for index := range history.Events {
		if history.Events[index].Kind == "predicate_evaluated" {
			history.Events[index].PredicateInputRole = "telemetry_version"
		}
	}
	_, violations := deriveVisibility(history, contract)
	if !containsViolation(violations, "does not show that the X-derived fact participates") {
		t.Fatalf("violations = %v, want semantic relevance rejection", violations)
	}
}

func TestTimestampDoesNotTurnUIntoDependency(t *testing.T) {
	evaluation := mustEvaluate(t, solutionPath)
	if len(evaluation.Violations) != 0 {
		t.Fatalf("solution violations = %v", evaluation.Violations)
	}
	review := visibilityReviewByID(t, evaluation.Review, "V-4")
	for _, fact := range review.RequiredFacts {
		if fact.FactID == "U" {
			t.Fatal("chronological U appears in the required fact set")
		}
	}
}

func TestObjectHistoriesDeriveExpectedWitnessSpace(t *testing.T) {
	contract := objectContractByID(t, mustLoadContract(t), "C-O1")
	corpus := mustLoadCorpus(t)
	tests := []struct {
		id      string
		verdict string
		witness []string
		reason  string
	}{
		{id: "O-1", verdict: VerdictViolates, reason: "forced_order_result_conflict"},
		{id: "O-2", verdict: VerdictSatisfies, witness: []string{"R-43", "W-43"}},
		{id: "O-3", verdict: VerdictSatisfies, witness: []string{"W-44", "R-44"}},
		{id: "O-4", verdict: VerdictViolates, reason: "unexplained_result"},
	}
	for _, test := range tests {
		t.Run(test.id, func(t *testing.T) {
			history := objectHistoryByID(t, corpus, test.id)
			truth, violations := deriveObject(history, contract)
			if len(violations) != 0 {
				t.Fatalf("derive violations = %v", violations)
			}
			if truth.Verdict != test.verdict {
				t.Fatalf("verdict = %s, want %s", truth.Verdict, test.verdict)
			}
			if len(test.witness) != 0 && !containsWitness(truth.LegalWitnesses, test.witness) {
				t.Fatalf("witnesses = %v, want %v", truth.LegalWitnesses, test.witness)
			}
			if test.reason != "" && truth.Contradiction.Kind != test.reason {
				t.Fatalf("contradiction = %+v, want %s", truth.Contradiction, test.reason)
			}
		})
	}
}

func TestVerifierAcceptsAnyLegalWitness(t *testing.T) {
	contract := objectContractByID(t, mustLoadContract(t), "C-O1")
	history := ObjectHistory{
		ID: "O-ALT", ContractID: "C-O1",
		Events: []ObjectEvent{
			{ID: "A-INVOKE", Order: 1, Kind: "invoke", OperationID: "R-A", OperationKind: "get_display_name"},
			{ID: "B-INVOKE", Order: 2, Kind: "invoke", OperationID: "R-B", OperationKind: "get_display_name"},
			{ID: "A-RESPOND", Order: 3, Kind: "respond", OperationID: "R-A", Result: "Jess"},
			{ID: "B-RESPOND", Order: 4, Kind: "respond", OperationID: "R-B", Result: "Jess"},
		},
	}
	truth, violations := deriveObject(history, contract)
	if len(violations) != 0 || len(truth.LegalWitnesses) != 2 {
		t.Fatalf("truth = %+v, violations = %v", truth, violations)
	}
	review := ObjectReview{
		HistoryID: "O-ALT", WitnessOrder: []string{"R-B", "R-A"},
		Contradiction: Contradiction{Kind: "none", OperationIDs: []string{}, EvidenceIDs: []string{}},
		Verdict:       VerdictSatisfies,
	}
	if violations := validateObjectReview(history, review, truth); len(violations) != 0 {
		t.Fatalf("alternative witness violations = %v", violations)
	}
}

func TestNoWitnessClaimIsIndependentlyRechecked(t *testing.T) {
	corpus := mustLoadCorpus(t)
	for historyIndex := range corpus.ObjectHistories {
		if corpus.ObjectHistories[historyIndex].ID != "O-1" {
			continue
		}
		for eventIndex := range corpus.ObjectHistories[historyIndex].Events {
			if corpus.ObjectHistories[historyIndex].Events[eventIndex].ID == "O1-R-RESPOND" {
				corpus.ObjectHistories[historyIndex].Events[eventIndex].Result = "Jessica"
			}
		}
	}
	evaluation := Evaluate(mustLoadReview(t, solutionPath), mustLoadContract(t), corpus)
	if !containsViolation(evaluation.Violations, "O-1: verdict does not match the legal witness space") ||
		!containsViolation(evaluation.Violations, "O-1: submitted order is not one legal sequential witness") {
		t.Fatalf("violations = %v, want independent witness-space rejection", evaluation.Violations)
	}
}

func TestInformationPathMustCitePredicateEvidence(t *testing.T) {
	review := mustLoadReview(t, solutionPath)
	for reviewIndex := range review.VisibilityReviews {
		if review.VisibilityReviews[reviewIndex].HistoryID != "V-4" {
			continue
		}
		review.VisibilityReviews[reviewIndex].InformationPaths[0].EvidenceIDs = []string{
			"V4-X", "V4-RX", "V4-Y-SUBMIT", "V4-Y-ACCEPT",
		}
	}
	evaluation := Evaluate(review, mustLoadContract(t), mustLoadCorpus(t))
	if !containsViolation(evaluation.Violations, "omits required evidence V4-Y-PREDICATE") {
		t.Fatalf("violations = %v, want predicate-evidence rejection", evaluation.Violations)
	}
}

func TestLoadReviewRejectsUnknownAndTrailingJSON(t *testing.T) {
	data, err := os.ReadFile(solutionPath)
	if err != nil {
		t.Fatal(err)
	}
	unknown := strings.Replace(string(data), `"name": "solution"`, `"name": "solution", "expected_verdicts": true`, 1)
	unknownPath := filepath.Join(t.TempDir(), "unknown.json")
	if err := os.WriteFile(unknownPath, []byte(unknown), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadReview(unknownPath); err == nil {
		t.Fatal("review with unknown field unexpectedly passed")
	}

	trailingPath := filepath.Join(t.TempDir(), "trailing.json")
	if err := os.WriteFile(trailingPath, append(data, []byte("\n{}\n")...), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadReview(trailingPath); err == nil {
		t.Fatal("review with trailing JSON unexpectedly passed")
	}
}

func TestLoadReviewRejectsMissingBoundaryClaims(t *testing.T) {
	data, err := os.ReadFile(solutionPath)
	if err != nil {
		t.Fatal(err)
	}
	fields := []string{
		"",
		"session_visibility_establishes_object_linearizability",
		"object_history_establishes_session_visibility",
		"one_dependency_establishes_general_causal_consistency",
		"one_legal_history_guarantees_future_availability",
	}
	for _, field := range fields {
		name := field
		if name == "" {
			name = "claims-section"
		}
		t.Run(name, func(t *testing.T) {
			var document map[string]any
			if err := json.Unmarshal(data, &document); err != nil {
				t.Fatal(err)
			}
			if field == "" {
				delete(document, "claims")
			} else {
				claims, ok := document["claims"].(map[string]any)
				if !ok {
					t.Fatal("solution claims are not an object")
				}
				delete(claims, field)
			}
			candidate, err := json.Marshal(document)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), name+".json")
			if err := os.WriteFile(path, candidate, 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadReview(path); err == nil {
				t.Fatalf("review missing %s unexpectedly passed", name)
			}
		})
	}
}

func TestDuplicateVisibilityOperationEventRejected(t *testing.T) {
	corpus := mustLoadCorpus(t)
	for historyIndex := range corpus.VisibilityHistories {
		if corpus.VisibilityHistories[historyIndex].ID != "V-2" {
			continue
		}
		history := &corpus.VisibilityHistories[historyIndex]
		history.Events = append(history.Events, VisibilityEvent{
			ID: "V2-RB-CONTRADICTS", Order: len(history.Events) + 1,
			Kind: "read_succeeded", SessionID: "J-88", OperationID: "R-2B",
			StateFactIDs: []string{"N-10", "N-11"}, ResultFactIDs: []string{"N-10", "N-11"},
		})
	}
	evaluation := Evaluate(mustLoadReview(t, solutionPath), mustLoadContract(t), corpus)
	if !containsViolation(evaluation.Violations, "operation R-2B has duplicate read_succeeded events") {
		t.Fatalf("violations = %v, want duplicate visibility-operation rejection", evaluation.Violations)
	}
}

func TestLoadContractAndCorpusRejectTrailingJSON(t *testing.T) {
	for _, test := range []struct {
		name string
		path string
		load func(string) error
	}{
		{name: "contract", path: contractPath, load: func(path string) error { _, err := LoadContract(path); return err }},
		{name: "histories", path: historiesPath, load: func(path string) error { _, err := LoadCorpus(path); return err }},
	} {
		t.Run(test.name, func(t *testing.T) {
			data, err := os.ReadFile(test.path)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), test.name+".json")
			if err := os.WriteFile(path, append(data, []byte("\n{}\n")...), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := test.load(path); err == nil {
				t.Fatalf("%s with trailing JSON unexpectedly passed", test.name)
			}
		})
	}
}

func TestBoundarySessionProofDoesNotEstablishObjectHistory(t *testing.T) {
	review := mustLoadReview(t, solutionPath)
	*review.Claims.SessionVisibilityEstablishesObjectLinearizability = true
	evaluation := Evaluate(review, mustLoadContract(t), mustLoadCorpus(t))
	if !containsViolation(evaluation.Violations, "does not establish object-wide linearizability") {
		t.Fatalf("violations = %v", evaluation.Violations)
	}
}

func TestBoundaryObjectHistoryDoesNotEstablishSessionVisibility(t *testing.T) {
	review := mustLoadReview(t, solutionPath)
	*review.Claims.ObjectHistoryEstablishesSessionVisibility = true
	evaluation := Evaluate(review, mustLoadContract(t), mustLoadCorpus(t))
	if !containsViolation(evaluation.Violations, "does not establish a session visibility guarantee") {
		t.Fatalf("violations = %v", evaluation.Violations)
	}
}

func TestBoundaryOneDependencyIsNotGeneralCausalConsistency(t *testing.T) {
	review := mustLoadReview(t, solutionPath)
	*review.Claims.OneDependencyEstablishesGeneralCausalConsistency = true
	evaluation := Evaluate(review, mustLoadContract(t), mustLoadCorpus(t))
	if !containsViolation(evaluation.Violations, "does not establish general causal consistency") {
		t.Fatalf("violations = %v", evaluation.Violations)
	}
}

func TestBoundaryOneLegalHistoryDoesNotGuaranteeAvailability(t *testing.T) {
	review := mustLoadReview(t, solutionPath)
	*review.Claims.OneLegalHistoryGuaranteesFutureAvailability = true
	evaluation := Evaluate(review, mustLoadContract(t), mustLoadCorpus(t))
	if !containsViolation(evaluation.Violations, "does not guarantee future availability") {
		t.Fatalf("violations = %v", evaluation.Violations)
	}
}

func TestBoundaryOverlapDoesNotPermitArbitraryResult(t *testing.T) {
	contract := objectContractByID(t, mustLoadContract(t), "C-O1")
	history := objectHistoryByID(t, mustLoadCorpus(t), "O-4")
	truth, violations := deriveObject(history, contract)
	if len(violations) != 0 || truth.Verdict != VerdictViolates || truth.Contradiction.Kind != "unexplained_result" {
		t.Fatalf("truth = %+v, violations = %v", truth, violations)
	}
}

func mustEvaluate(t *testing.T, reviewPath string) Evaluation {
	t.Helper()
	return Evaluate(mustLoadReview(t, reviewPath), mustLoadContract(t), mustLoadCorpus(t))
}

func mustLoadContract(t *testing.T) Contract {
	t.Helper()
	contract, err := LoadContract(contractPath)
	if err != nil {
		t.Fatal(err)
	}
	return contract
}

func mustLoadCorpus(t *testing.T) Corpus {
	t.Helper()
	corpus, err := LoadCorpus(historiesPath)
	if err != nil {
		t.Fatal(err)
	}
	return corpus
}

func mustLoadReview(t *testing.T, path string) Review {
	t.Helper()
	review, err := LoadReview(path)
	if err != nil {
		t.Fatal(err)
	}
	return review
}

func visibilityContractByID(t *testing.T, contract Contract, id string) VisibilityContract {
	t.Helper()
	for _, candidate := range contract.VisibilityContracts {
		if candidate.ID == id {
			return candidate
		}
	}
	t.Fatalf("visibility contract %s is missing", id)
	return VisibilityContract{}
}

func objectContractByID(t *testing.T, contract Contract, id string) ObjectContract {
	t.Helper()
	for _, candidate := range contract.ObjectContracts {
		if candidate.ID == id {
			return candidate
		}
	}
	t.Fatalf("object contract %s is missing", id)
	return ObjectContract{}
}

func visibilityHistoryByID(t *testing.T, corpus Corpus, id string) VisibilityHistory {
	t.Helper()
	for _, candidate := range corpus.VisibilityHistories {
		if candidate.ID == id {
			return candidate
		}
	}
	t.Fatalf("visibility history %s is missing", id)
	return VisibilityHistory{}
}

func objectHistoryByID(t *testing.T, corpus Corpus, id string) ObjectHistory {
	t.Helper()
	for _, candidate := range corpus.ObjectHistories {
		if candidate.ID == id {
			return candidate
		}
	}
	t.Fatalf("object history %s is missing", id)
	return ObjectHistory{}
}

func visibilityReviewByID(t *testing.T, review Review, id string) VisibilityReview {
	t.Helper()
	for _, candidate := range review.VisibilityReviews {
		if candidate.HistoryID == id {
			return candidate
		}
	}
	t.Fatalf("visibility review %s is missing", id)
	return VisibilityReview{}
}

func containsViolation(violations []string, fragment string) bool {
	for _, violation := range violations {
		if strings.Contains(violation, fragment) {
			return true
		}
	}
	return false
}
