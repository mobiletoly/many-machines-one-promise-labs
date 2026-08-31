package replay

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	contractPath   = "../contract.json"
	designPath     = "../workflow-design.json"
	evidencePath   = "../authority-evidence.jsonl"
	candidatesPath = "../claim-candidates.json"
	starterPath    = "../starter/product-claim-review.json"
	solutionPath   = "../solution/product-claim-review.json"
)

func TestStarterReviewFails(t *testing.T) {
	evaluation := mustEvaluate(t, starterPath)
	if len(evaluation.Violations) == 0 {
		t.Fatal("starter review unexpectedly passed")
	}
	if !containsViolation(evaluation.Violations, "cut review count") {
		t.Fatalf("violations = %v", evaluation.Violations)
	}
}

func TestSolutionReviewPasses(t *testing.T) {
	evaluation := mustEvaluate(t, solutionPath)
	if len(evaluation.Violations) != 0 {
		t.Fatalf("solution violations = %v", evaluation.Violations)
	}
}

func TestUnknownPendingRejectedAndNeverAttemptedRemainDistinct(t *testing.T) {
	evaluation := mustEvaluate(t, solutionPath)
	k1 := cutTruth(t, evaluation, "P-95", "K1")
	if operationTruth(t, k1, "R-95").State != StateUnknown || operationTruth(t, k1, "F-95").State != StateNeverAttempted {
		t.Fatalf("K1 operation truths = %+v", k1.OperationTruths)
	}
	k3 := cutTruth(t, evaluation, "P-95", "K3")
	if operationTruth(t, k3, "F-95").State != StatePending {
		t.Fatalf("K3 operation truths = %+v", k3.OperationTruths)
	}
	k4 := cutTruth(t, evaluation, "P-95", "K4")
	if operationTruth(t, k4, "F-95").State != StateRejected || operationTruth(t, k4, "G-95").State != StateUnknown {
		t.Fatalf("K4 operation truths = %+v", k4.OperationTruths)
	}
}

func TestLaterEvidenceDoesNotRewriteEarlierCut(t *testing.T) {
	evaluation := mustEvaluate(t, solutionPath)
	k1 := cutTruth(t, evaluation, "P-95", "K1")
	k5 := cutTruth(t, evaluation, "P-95", "K5")
	if operationTruth(t, k1, "R-95").State != StateUnknown || operationTruth(t, k5, "R-95").State != StateRejected {
		t.Fatalf("K1/K5 reservation truths = %+v / %+v", k1.OperationTruths, k5.OperationTruths)
	}
	if claimTruth(t, k1, "payment-captured").Verdict != VerdictSupported || claimTruth(t, k5, "payment-captured").Verdict != VerdictSupported {
		t.Fatal("later return evidence rewrote the earlier accepted capture")
	}
}

func TestRejectedAttemptDoesNotProveUnsatisfiedObligation(t *testing.T) {
	evaluation := mustEvaluate(t, solutionPath)
	k4 := cutTruth(t, evaluation, "P-95", "K4")
	if k4.Obligation.Status != ObligationUnresolved || k4.Obligation.EvidenceDomainComplete {
		t.Fatalf("K4 obligation = %+v", k4.Obligation)
	}
	claim := claimTruth(t, k4, "return-unsatisfied")
	if claim.Verdict != VerdictUnsupported || claim.Reason != ReasonDomainIncomplete {
		t.Fatalf("K4 return-unsatisfied = %+v", claim)
	}
}

func TestCompleteRejectedSatisfactionDomainProvesUnsatisfied(t *testing.T) {
	evaluation := mustEvaluate(t, solutionPath)
	u1 := cutTruth(t, evaluation, "P-96", "U1")
	if u1.Obligation.Status != ObligationUnsatisfied || !u1.Obligation.EvidenceDomainComplete {
		t.Fatalf("U1 obligation = %+v", u1.Obligation)
	}
}

func TestAcceptedAlternativeSatisfiesWithoutRewritingHistory(t *testing.T) {
	evaluation := mustEvaluate(t, solutionPath)
	k5 := cutTruth(t, evaluation, "P-95", "K5")
	if k5.Obligation.Status != ObligationSatisfied || operationTruth(t, k5, "F-95").State != StateRejected || operationTruth(t, k5, "G-95").State != StateAccepted {
		t.Fatalf("K5 truth = %+v", k5)
	}
}

func TestReviewOrderDoesNotMatter(t *testing.T) {
	review := mustLoadReview(t, solutionPath)
	review.CutReviews[0], review.CutReviews[6] = review.CutReviews[6], review.CutReviews[0]
	review.CutReviews[1].AuthorityStates[0], review.CutReviews[1].AuthorityStates[3] = review.CutReviews[1].AuthorityStates[3], review.CutReviews[1].AuthorityStates[0]
	review.StrongerContractReview.ArchitectureReviews[0], review.StrongerContractReview.ArchitectureReviews[3] = review.StrongerContractReview.ArchitectureReviews[3], review.StrongerContractReview.ArchitectureReviews[0]
	evaluation := evaluateReview(t, review)
	if len(evaluation.Violations) != 0 {
		t.Fatalf("reordered review violations = %v", evaluation.Violations)
	}
}

func TestReviewRejectsUnknownAndTrailingJSON(t *testing.T) {
	data, err := os.ReadFile(solutionPath)
	if err != nil {
		t.Fatal(err)
	}
	unknown := strings.Replace(string(data), `"name": "reference"`, `"name": "reference", "answer": 42`, 1)
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

func TestEvidenceRejectsMissingSequence(t *testing.T) {
	design := mustLoadDossier(t).design
	data, err := os.ReadFile(evidencePath)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	data = []byte(strings.Join(append(lines[:4], lines[5:]...), "\n") + "\n")
	path := filepath.Join(t.TempDir(), "missing.jsonl")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadEvidence(path, design); err == nil {
		t.Fatal("evidence with a missing sequence unexpectedly passed")
	}
}

func TestBoundaryMixedOutcomeFailsStrongerContract(t *testing.T) {
	evaluation := mustEvaluate(t, solutionPath)
	truth := architectureTruth(t, evaluation, "independent-effects")
	if truth.ContractFit != VerdictFail || !contains(truth.Reasons, "forbidden_mixed_terminal_outcome_reachable") {
		t.Fatalf("independent architecture truth = %+v", truth)
	}
}

func TestBoundaryOneAtomicAuthorityDoesNotRequireDistributedCommit(t *testing.T) {
	evaluation := mustEvaluate(t, solutionPath)
	truth := architectureTruth(t, evaluation, "single-commerce-authority")
	if truth.ContractFit != VerdictPass || truth.DistributedAtomicCommitRequired != AtomicNotRequired {
		t.Fatalf("single-authority truth = %+v", truth)
	}
}

func TestBoundaryCannotPrepareFailsCommonDecision(t *testing.T) {
	evaluation := mustEvaluate(t, solutionPath)
	truth := architectureTruth(t, evaluation, "coordinator-over-immediate-capture")
	if truth.ContractFit != VerdictFail || !contains(truth.Reasons, "required_participant_cannot_prepare") {
		t.Fatalf("cannot-prepare truth = %+v", truth)
	}
}

func TestBoundaryAlwaysAbortFailsDeclaredProgress(t *testing.T) {
	evaluation := mustEvaluate(t, solutionPath)
	truth := architectureTruth(t, evaluation, "always-abort")
	if truth.ContractFit != VerdictFail || !contains(truth.Reasons, "healthy_success_progress_not_met") {
		t.Fatalf("always-abort truth = %+v", truth)
	}
}

func TestBoundaryLaterCorrectionDoesNotRepairOriginalAtomicClaim(t *testing.T) {
	review := mustLoadReview(t, solutionPath)
	review.StrongerContractReview.LaterCorrectionSatisfiesOriginalContract = VerdictSupported
	evaluation := evaluateReview(t, review)
	if !containsViolation(evaluation.Violations, "later correction cannot satisfy") {
		t.Fatalf("violations = %v", evaluation.Violations)
	}
}

func TestBoundaryDoesNotSelectTwoPhaseCommit(t *testing.T) {
	review := mustLoadReview(t, solutionPath)
	review.StrongerContractReview.TwoPhaseCommitSelected = VerdictSupported
	evaluation := evaluateReview(t, review)
	if !containsViolation(evaluation.Violations, "does not select two-phase commit") {
		t.Fatalf("violations = %v", evaluation.Violations)
	}
}

type dossier struct {
	contract   Contract
	design     WorkflowDesign
	evidence   []EvidenceEvent
	candidates ClaimCandidates
}

func mustLoadDossier(t *testing.T) dossier {
	t.Helper()
	contract, err := LoadContract(contractPath)
	if err != nil {
		t.Fatal(err)
	}
	design, err := LoadWorkflowDesign(designPath)
	if err != nil {
		t.Fatal(err)
	}
	evidence, err := LoadEvidence(evidencePath, design)
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := LoadClaimCandidates(candidatesPath, contract, design)
	if err != nil {
		t.Fatal(err)
	}
	return dossier{contract: contract, design: design, evidence: evidence, candidates: candidates}
}

func mustEvaluate(t *testing.T, path string) Evaluation {
	t.Helper()
	return evaluateReview(t, mustLoadReview(t, path))
}

func evaluateReview(t *testing.T, review Review) Evaluation {
	t.Helper()
	dossier := mustLoadDossier(t)
	return Evaluate(review, dossier.contract, dossier.design, dossier.evidence, dossier.candidates)
}

func mustLoadReview(t *testing.T, path string) Review {
	t.Helper()
	review, err := LoadReview(path)
	if err != nil {
		t.Fatal(err)
	}
	return review
}

func cutTruth(t *testing.T, evaluation Evaluation, caseID, cutID string) CutTruth {
	t.Helper()
	for _, truth := range evaluation.CutTruths {
		if truth.CaseID == caseID && truth.CutID == cutID {
			return truth
		}
	}
	t.Fatalf("cut %s/%s not found", caseID, cutID)
	return CutTruth{}
}

func operationTruth(t *testing.T, cut CutTruth, operationID string) OperationTruth {
	t.Helper()
	for _, truth := range cut.OperationTruths {
		if truth.Operation.ID == operationID {
			return truth
		}
	}
	t.Fatalf("operation %s not found", operationID)
	return OperationTruth{}
}

func claimTruth(t *testing.T, cut CutTruth, claimID string) ClaimTruth {
	t.Helper()
	for _, truth := range cut.ClaimTruths {
		if truth.ClaimID == claimID {
			return truth
		}
	}
	t.Fatalf("claim %s not found", claimID)
	return ClaimTruth{}
}

func architectureTruth(t *testing.T, evaluation Evaluation, architectureID string) ArchitectureTruth {
	t.Helper()
	for _, truth := range evaluation.ArchitectureTruths {
		if truth.ArchitectureID == architectureID {
			return truth
		}
	}
	t.Fatalf("architecture %s not found", architectureID)
	return ArchitectureTruth{}
}

func containsViolation(violations []string, fragment string) bool {
	for _, violation := range violations {
		if strings.Contains(violation, fragment) {
			return true
		}
	}
	return false
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
