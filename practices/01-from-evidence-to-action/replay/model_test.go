package replay

import (
	"strings"
	"testing"
)

const (
	evidencePath = "../evidence.jsonl"
	starterPath  = "../starter/policy.json"
	solutionPath = "../solution/policy.json"
)

func TestStarterPolicyViolatesContract(t *testing.T) {
	policy := mustLoadPolicy(t, starterPath)
	evidence := mustLoadEvidence(t)
	evaluation := Evaluate(policy, evidence)
	if len(evaluation.Violations) == 0 {
		t.Fatal("starter policy unexpectedly satisfied the declared contract")
	}
	if evaluation.Violations[0] != "initial evidence also fits a capable X with impaired evidence paths; it does not establish crashed" {
		t.Fatalf("first violation = %q, want unsupported crash inference", evaluation.Violations[0])
	}
}

func TestSolutionPolicySatisfiesContract(t *testing.T) {
	policy := mustLoadPolicy(t, solutionPath)
	evidence := mustLoadEvidence(t)
	evaluation := Evaluate(policy, evidence)
	if len(evaluation.Violations) != 0 {
		t.Fatalf("solution violations = %v", evaluation.Violations)
	}
}

func TestPolicyChangesFutureEvidence(t *testing.T) {
	evidence := mustLoadEvidence(t)
	withoutTrial := mustLoadPolicy(t, solutionPath)
	withoutTrial.EvidenceAction = EvidenceNone
	withoutTrial.MaxEvidenceCallsPerRound = 0
	withoutTrial.RestoreOn = RestoreNone

	evaluation := Evaluate(withoutTrial, evidence)
	gamma := reportByName(t, evaluation, "gamma")
	if gamma.Restored {
		t.Fatal("gamma restored without an admitted evidence-producing interaction")
	}
	for _, event := range gamma.Events {
		if event.Action == "representative_shipping_quote" {
			t.Fatal("representative evidence existed even though policy admitted no trial")
		}
	}
}

func TestOracleRejectsEvidenceOutsideTheDeclaredRecord(t *testing.T) {
	policy := mustLoadPolicy(t, solutionPath)
	evidence := []Evidence{{
		At:          "10:00:00.000",
		Observer:    "inside-x",
		Kind:        "execution_truth",
		Participant: "shipping-x",
		Detail:      "shipping-x crashed",
	}}

	evaluation := Evaluate(policy, evidence)
	if !containsViolation(evaluation.Violations, "unsupported kind \"execution_truth\"") {
		t.Fatalf("violations = %v, want unsupported-evidence rejection", evaluation.Violations)
	}
}

func TestOracleRejectsIncompleteEvidence(t *testing.T) {
	policy := mustLoadPolicy(t, solutionPath)
	evidence := mustLoadEvidence(t)
	evaluation := Evaluate(policy, evidence[:len(evidence)-1])
	if !containsViolation(evaluation.Violations, "detector suspicion outputs, want 1") {
		t.Fatalf("violations = %v, want incomplete-evidence rejection", evaluation.Violations)
	}
}

func TestBoundaryHealthEndpointCannotRestoreShippingQuoteAdmission(t *testing.T) {
	evidence := mustLoadEvidence(t)
	policy := mustLoadPolicy(t, solutionPath)
	policy.EvidenceAction = EvidenceHealthCheck
	policy.RestoreOn = RestoreHealthCheck

	evaluation := Evaluate(policy, evidence)
	alpha := reportByName(t, evaluation, "alpha")
	if !alpha.Restored || alpha.RestoredBy != RestoreHealthCheck {
		t.Fatal("boundary setup did not restore from the health endpoint")
	}
	if !containsViolation(evaluation.Violations, "shipping quotes remained incapable") {
		t.Fatalf("violations = %v, want unrepresentative-evidence failure", evaluation.Violations)
	}
}

func TestBoundaryPolicyDoesNotControlAnotherCaller(t *testing.T) {
	policy := mustLoadPolicy(t, solutionPath)
	if !PolicyApplies(policy, "checkout-p", "shipping_quote") {
		t.Fatal("solution policy does not cover its declared caller and operation")
	}
	if PolicyApplies(policy, "checkout-p2", "shipping_quote") {
		t.Fatal("checkout-p policy unexpectedly controls checkout-p2")
	}
}

func TestBoundarySuccessfulTrialDoesNotEstablishFullHealth(t *testing.T) {
	evidence := mustLoadEvidence(t)
	policy := mustLoadPolicy(t, solutionPath)
	policy.PostRestoreClaim = "x_fully_healthy"

	evaluation := Evaluate(policy, evidence)
	if !containsViolation(evaluation.Violations, "post-restore claim \"x_fully_healthy\" exceeds") {
		t.Fatalf("violations = %v, want stronger-claim rejection", evaluation.Violations)
	}
}

func mustLoadPolicy(t *testing.T, path string) Policy {
	t.Helper()
	policy, err := LoadPolicy(path)
	if err != nil {
		t.Fatal(err)
	}
	return policy
}

func mustLoadEvidence(t *testing.T) []Evidence {
	t.Helper()
	evidence, err := LoadEvidence(evidencePath)
	if err != nil {
		t.Fatal(err)
	}
	return evidence
}

func reportByName(t *testing.T, evaluation Evaluation, name string) ScenarioReport {
	t.Helper()
	for _, report := range evaluation.Reports {
		if report.Name == name {
			return report
		}
	}
	t.Fatalf("scenario %q not found", name)
	return ScenarioReport{}
}

func containsViolation(violations []string, fragment string) bool {
	for _, violation := range violations {
		if strings.Contains(violation, fragment) {
			return true
		}
	}
	return false
}
