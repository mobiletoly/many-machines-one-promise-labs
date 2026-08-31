package replay

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	contractPath = "../contract.json"
	historyPath  = "../accepted-history.jsonl"
	evidencePath = "../recovery-evidence.jsonl"
	eventsPath   = "../serving-events.jsonl"
	starterPath  = "../starter/recovery-decision.json"
	solutionPath = "../solution/recovery-decision.json"
)

func TestStarterDecisionFails(t *testing.T) {
	evaluation := mustEvaluate(t, starterPath)
	if len(evaluation.Violations) == 0 {
		t.Fatal("starter decision unexpectedly passed")
	}
	if !containsViolation(evaluation.Violations, "latest complete cut is not supported") {
		t.Fatalf("violations = %v", evaluation.Violations)
	}
}

func TestSolutionDecisionPasses(t *testing.T) {
	evaluation := mustEvaluate(t, solutionPath)
	if len(evaluation.Violations) != 0 {
		t.Fatalf("solution violations = %v", evaluation.Violations)
	}
	if evaluation.RTOTruth.BoundaryEvent.ID != "E8" || !evaluation.RTOTruth.Pass {
		t.Fatalf("RTO truth = %+v", evaluation.RTOTruth)
	}
}

func TestGapDoesNotBecomeACompletePrefix(t *testing.T) {
	evaluation := mustEvaluate(t, solutionPath)
	state := evaluation.StateTruth["R-A"]
	if state.LatestCompleteFactID != "H4" || contains(state.CompletePrefixFactIDs, "H8") {
		t.Fatalf("R-A truth = %+v", state)
	}
}

func TestMissingFactRequiredThroughRPOFloorFails(t *testing.T) {
	evaluation := mustEvaluate(t, solutionPath)
	state := evaluation.StateTruth["R-A"]
	if state.RPOPass || state.Admissible {
		t.Fatalf("R-A truth = %+v, want missing H5 at or before the floor to fail RPO", state)
	}
}

func TestQuietIntervalAfterCompletePrefixStillMeetsRPO(t *testing.T) {
	contract, facts, states, _ := mustLoadDossier(t)
	for index := range facts {
		if facts[index].ID == "H5" {
			facts[index].AcceptedAt = "2026-08-30T11:59:35Z"
		}
	}
	truth := deriveStateTruth(contract, facts, states[0])
	if !truth.RPOPass {
		t.Fatalf("R-A truth = %+v, want coverage through a quiet RPO floor", truth)
	}
}

func TestRPOAndProtectedAcknowledgementAreIndependent(t *testing.T) {
	evaluation := mustEvaluate(t, solutionPath)
	state := evaluation.StateTruth["R-B"]
	if !state.RPOPass || state.AcknowledgementPass || state.Admissible {
		t.Fatalf("R-B truth = %+v", state)
	}
}

func TestEvidenceOrderDoesNotMatter(t *testing.T) {
	decision := mustLoadDecision(t, solutionPath)
	review := decision.StateReviews[2]
	review.CompletePrefixFactIDs[0], review.CompletePrefixFactIDs[7] = review.CompletePrefixFactIDs[7], review.CompletePrefixFactIDs[0]
	decision.StateReviews[2] = review
	decision.RTOReview.EvidenceIDs[0], decision.RTOReview.EvidenceIDs[9] = decision.RTOReview.EvidenceIDs[9], decision.RTOReview.EvidenceIDs[0]
	evaluation := evaluateDecision(t, decision)
	if len(evaluation.Violations) != 0 {
		t.Fatalf("alternative evidence order violations = %v", evaluation.Violations)
	}
}

func TestDecisionRejectsUnknownAndTrailingJSON(t *testing.T) {
	data, err := os.ReadFile(solutionPath)
	if err != nil {
		t.Fatal(err)
	}
	unknown := strings.Replace(string(data), `"name": "solution"`, `"name": "solution", "answer": 42`, 1)
	unknownPath := filepath.Join(t.TempDir(), "unknown.json")
	if err := os.WriteFile(unknownPath, []byte(unknown), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadDecision(unknownPath); err == nil {
		t.Fatal("decision with unknown field unexpectedly passed")
	}
	trailingPath := filepath.Join(t.TempDir(), "trailing.json")
	if err := os.WriteFile(trailingPath, append(data, []byte("\n{}\n")...), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadDecision(trailingPath); err == nil {
		t.Fatal("decision with trailing JSON unexpectedly passed")
	}
}

func TestJSONLRejectsUnknownAndTrailingValues(t *testing.T) {
	unknownPath := filepath.Join(t.TempDir(), "unknown.jsonl")
	unknown := `{"id":"H1","position":1,"accepted_at":"2026-08-30T11:00:00Z","operation_id":"O1","operation_class":"x","acknowledgement_result":"y","answer":42}`
	if err := os.WriteFile(unknownPath, []byte(unknown+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadAcceptedHistory(unknownPath); err == nil {
		t.Fatal("JSONL with unknown field unexpectedly passed")
	}
	trailingPath := filepath.Join(t.TempDir(), "trailing.jsonl")
	trailing := `{"id":"H1","position":1,"accepted_at":"2026-08-30T11:00:00Z","operation_id":"O1","operation_class":"x","acknowledgement_result":"y"} {}`
	if err := os.WriteFile(trailingPath, []byte(trailing+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadAcceptedHistory(trailingPath); err == nil {
		t.Fatal("JSONL with trailing value unexpectedly passed")
	}
}

func TestDossierRejectsHistoryAfterDisruption(t *testing.T) {
	contract, facts, states, events := mustLoadDossier(t)
	facts[len(facts)-1].AcceptedAt = "2026-08-30T12:00:01Z"
	evaluation := Evaluate(mustLoadDecision(t, solutionPath), contract, facts, states, events)
	if !containsViolation(evaluation.Violations, "occurs after the disruption boundary") {
		t.Fatalf("violations = %v", evaluation.Violations)
	}
}

func TestBoundaryGenericHealthDoesNotEndRTO(t *testing.T) {
	evaluation := mustEvaluate(t, solutionPath)
	if evaluation.RTOTruth.BoundaryEvent.ID == "E2" {
		t.Fatal("generic health became the RTO end")
	}
	if !rejectedFor(evaluation.RTOTruth.RejectedEarlier, "E2", ReasonGeneric) {
		t.Fatalf("rejected = %+v", evaluation.RTOTruth.RejectedEarlier)
	}
}

func TestBoundaryFirstCustomerResultDoesNotMoveEstablishedRTOEnd(t *testing.T) {
	evaluation := mustEvaluate(t, solutionPath)
	if evaluation.RTOTruth.BoundaryEvent.ID != "E8" || evaluation.RTOTruth.FirstCustomerResult.ID != "E9" || evaluation.RTOTruth.FirstCustomerRelation != RelationAfter {
		t.Fatalf("RTO truth = %+v", evaluation.RTOTruth)
	}
}

func TestBoundaryValidStateWithWrongInterpretationDoesNotQualify(t *testing.T) {
	evaluation := mustEvaluate(t, solutionPath)
	if !rejectedFor(evaluation.RTOTruth.RejectedEarlier, "E6", ReasonInterpret) {
		t.Fatalf("rejected = %+v", evaluation.RTOTruth.RejectedEarlier)
	}
}

func TestBoundaryValidStateWithIneligiblePathDoesNotQualify(t *testing.T) {
	evaluation := mustEvaluate(t, solutionPath)
	if !rejectedFor(evaluation.RTOTruth.RejectedEarlier, "E7", ReasonPath) {
		t.Fatalf("rejected = %+v", evaluation.RTOTruth.RejectedEarlier)
	}
}

func TestBoundaryPassingReviewDoesNotSupportStrongerClaims(t *testing.T) {
	decision := mustLoadDecision(t, solutionPath)
	decision.Claims.AllCapabilitiesRecovered = "supported"
	evaluation := evaluateDecision(t, decision)
	if !containsViolation(evaluation.Violations, "all capabilities recovered is a stronger unsupported claim") {
		t.Fatalf("violations = %v", evaluation.Violations)
	}
}

func mustEvaluate(t *testing.T, path string) Evaluation {
	t.Helper()
	return evaluateDecision(t, mustLoadDecision(t, path))
}

func evaluateDecision(t *testing.T, decision Decision) Evaluation {
	t.Helper()
	contract, facts, states, events := mustLoadDossier(t)
	return Evaluate(decision, contract, facts, states, events)
}

func mustLoadDossier(t *testing.T) (Contract, []AcceptedFact, []RecoveredState, []ServingEvent) {
	t.Helper()
	contract, err := LoadContract(contractPath)
	if err != nil {
		t.Fatal(err)
	}
	facts, err := LoadAcceptedHistory(historyPath)
	if err != nil {
		t.Fatal(err)
	}
	states, err := LoadRecoveryEvidence(evidencePath, facts)
	if err != nil {
		t.Fatal(err)
	}
	events, err := LoadServingEvents(eventsPath, states)
	if err != nil {
		t.Fatal(err)
	}
	return contract, facts, states, events
}

func mustLoadDecision(t *testing.T, path string) Decision {
	t.Helper()
	decision, err := LoadDecision(path)
	if err != nil {
		t.Fatal(err)
	}
	return decision
}

func containsViolation(violations []string, fragment string) bool {
	for _, violation := range violations {
		if strings.Contains(violation, fragment) {
			return true
		}
	}
	return false
}

func rejectedFor(events []RejectedEvent, id, reason string) bool {
	for _, event := range events {
		if event.EventID == id && contains(event.Reasons, reason) {
			return true
		}
	}
	return false
}
