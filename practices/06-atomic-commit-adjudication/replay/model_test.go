package replay

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/mobiletoly/many-machines-one-promise-labs/practices/06-atomic-commit-adjudication/internal/strictjson"
)

func loadTestCorpus(t *testing.T) Corpus {
	t.Helper()
	corpus, err := LoadCorpus("..")
	if err != nil {
		t.Fatalf("load corpus: %v", err)
	}
	return corpus
}

func truthByID(t *testing.T, truths []CaseTruth, id string) CaseTruth {
	t.Helper()
	for _, truth := range truths {
		if truth.CaseID == id {
			return truth
		}
	}
	t.Fatalf("truth %s not found", id)
	return CaseTruth{}
}

func cloneCorpus(corpus Corpus) Corpus {
	clone := corpus
	clone.Cases.Cases = append([]CaseDefinition(nil), corpus.Cases.Cases...)
	for index := range clone.Cases.Cases {
		clone.Cases.Cases[index].RequiredParticipants = append([]string(nil), corpus.Cases.Cases[index].RequiredParticipants...)
	}
	clone.ParticipantEvents = append([]ParticipantEvent(nil), corpus.ParticipantEvents...)
	clone.DecisionEvents = append([]DecisionEvent(nil), corpus.DecisionEvents...)
	for index := range clone.DecisionEvents {
		clone.DecisionEvents[index].PreparationEvidenceIDs = append([]string(nil), corpus.DecisionEvents[index].PreparationEvidenceIDs...)
	}
	return clone
}

func TestReferenceReviewPasses(t *testing.T) {
	corpus := loadTestCorpus(t)
	review, err := LoadReview("../solution/adjudication.json")
	if err != nil {
		t.Fatalf("load solution: %v", err)
	}
	evaluation, err := Evaluate(corpus, review)
	if err != nil {
		t.Fatalf("evaluate solution: %v", err)
	}
	if len(evaluation.Violations) != 0 {
		t.Fatalf("solution violations: %v", evaluation.Violations)
	}
}

func TestStarterFailsWithoutLeakingReferenceReview(t *testing.T) {
	corpus := loadTestCorpus(t)
	review, err := LoadReview("../starter/adjudication.json")
	if err != nil {
		t.Fatalf("load starter: %v", err)
	}
	evaluation, err := Evaluate(corpus, review)
	if err != nil {
		t.Fatalf("evaluate starter: %v", err)
	}
	if len(evaluation.Violations) == 0 {
		t.Fatal("starter unexpectedly passed")
	}
	joined := strings.Join(evaluation.Violations, "\n")
	if !strings.Contains(joined, "T-201 disposition") {
		t.Fatalf("wait-all falsifier did not fail T-201: %s", joined)
	}
}

func TestBoundaryUnknownAndNoneThroughCutRemainDistinct(t *testing.T) {
	truths, err := DeriveTruths(loadTestCorpus(t))
	if err != nil {
		t.Fatal(err)
	}
	unknown := truthByID(t, truths, "T-203")
	none := truthByID(t, truths, "T-204")
	if unknown.DecisionStatus.Value != DecisionUnknown || none.DecisionStatus.Value != DecisionNoneThroughCut {
		t.Fatalf("decision statuses collapsed: %s, %s", unknown.DecisionStatus.Value, none.DecisionStatus.Value)
	}
	if unknown.Disposition != DispositionRemainPrepared || none.Disposition != DispositionRemainPrepared {
		t.Fatalf("blocking disposition drifted: %s, %s", unknown.Disposition, none.Disposition)
	}
}

func TestBoundaryCoordinatorNoteCannotAuthorizeCommit(t *testing.T) {
	truths, err := DeriveTruths(loadTestCorpus(t))
	if err != nil {
		t.Fatal(err)
	}
	truth := truthByID(t, truths, "T-203")
	if truth.DecisionStatus.Value != DecisionUnknown || truth.Disposition != DispositionRemainPrepared {
		t.Fatalf("coordinator note became decision authority: %+v", truth)
	}
	if !sameSet(truth.DecisionStatus.EvidenceIDs, []string{"D203-READ"}) {
		t.Fatalf("coordinator note entered D decision-status evidence: %v", truth.DecisionStatus.EvidenceIDs)
	}
	if contains(truth.SupportingEvidenceIDs, "K203-NOTE") {
		t.Fatalf("non-authoritative coordinator note became required supporting evidence: %v", truth.SupportingEvidenceIDs)
	}
}

func TestContractRejectsOmittedFalseAuthorityDeclarations(t *testing.T) {
	data, err := os.ReadFile("../contract.json")
	if err != nil {
		t.Fatal(err)
	}

	mutations := map[string]struct {
		needle      []byte
		replacement []byte
	}{
		"unavailable read absence": {
			needle: []byte("    \"unavailable_decision_reads_establish_absence\": false,\n"),
		},
		"coordinator decision authority": {
			needle: []byte(",\n    \"coordinator_notes_are_decision_evidence\": false"),
		},
	}

	for name, mutation := range mutations {
		t.Run(name, func(t *testing.T) {
			mutated := bytes.Replace(data, mutation.needle, mutation.replacement, 1)
			if bytes.Equal(mutated, data) {
				t.Fatal("contract mutation did not remove the target declaration")
			}
			var contract Contract
			if err := strictjson.Unmarshal(mutated, &contract); err != nil {
				t.Fatalf("mutated contract must remain valid JSON: %v", err)
			}
			corpus := cloneCorpus(loadTestCorpus(t))
			corpus.Contract = contract
			if err := ValidateCorpus(corpus); err == nil {
				t.Fatal("contract with omitted required false declaration unexpectedly passed")
			}
		})
	}
}

func TestBoundaryInvalidDecisionAndLostCapabilityAreBreaches(t *testing.T) {
	truths, err := DeriveTruths(loadTestCorpus(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"T-205", "T-206"} {
		truth := truthByID(t, truths, id)
		if truth.DecisionContractStatus.Value != ContractBreach || truth.Disposition != DispositionReportBreach {
			t.Fatalf("%s did not expose protocol breach: %+v", id, truth)
		}
	}
}

func TestBoundarySameCaseIDChangesWhenDecisionEvidenceChanges(t *testing.T) {
	corpus := cloneCorpus(loadTestCorpus(t))
	participants := corpus.ParticipantEvents[:0]
	for _, event := range corpus.ParticipantEvents {
		if event.ID != "P201-M-COMMIT" {
			participants = append(participants, event)
		}
	}
	corpus.ParticipantEvents = participants
	decisions := corpus.DecisionEvents[:0]
	for _, event := range corpus.DecisionEvents {
		if event.ID == "D201-REC" {
			continue
		}
		if event.ID == "D201-READ" {
			event.Kind = "decision_read_unavailable"
			event.RecordID = ""
		}
		decisions = append(decisions, event)
	}
	corpus.DecisionEvents = decisions
	if err := ValidateCorpus(corpus); err != nil {
		t.Fatalf("mutated corpus invalid: %v", err)
	}
	truths, err := DeriveTruths(corpus)
	if err != nil {
		t.Fatal(err)
	}
	truth := truthByID(t, truths, "T-201")
	if truth.DecisionStatus.Value != DecisionUnknown || truth.Disposition != DispositionRemainPrepared {
		t.Fatalf("case identity selected the old answer: %+v", truth)
	}
}

func TestBoundarySameCaseIDChangesWhenMissingParticipantPrepared(t *testing.T) {
	corpus := cloneCorpus(loadTestCorpus(t))
	for index := range corpus.ParticipantEvents {
		if corpus.ParticipantEvents[index].ID == "P205-J-ABORT" {
			corpus.ParticipantEvents[index].Kind = "prepared"
		}
	}
	for index := range corpus.DecisionEvents {
		if corpus.DecisionEvents[index].ID == "D205-REC" {
			corpus.DecisionEvents[index].PreparationEvidenceIDs = append(corpus.DecisionEvents[index].PreparationEvidenceIDs, "P205-J-ABORT")
		}
	}
	if err := ValidateCorpus(corpus); err != nil {
		t.Fatalf("mutated corpus invalid: %v", err)
	}
	truths, err := DeriveTruths(corpus)
	if err != nil {
		t.Fatal(err)
	}
	truth := truthByID(t, truths, "T-205")
	if truth.DecisionContractStatus.Value != ContractValid || truth.Disposition != DispositionApplyCommit {
		t.Fatalf("case identity selected breach after evidence changed: %+v", truth)
	}
}

func TestBoundarySameCaseIDChangesWhenPreparedCapabilitySurvives(t *testing.T) {
	corpus := cloneCorpus(loadTestCorpus(t))
	events := corpus.ParticipantEvents[:0]
	for _, event := range corpus.ParticipantEvents {
		if event.ID != "P206-I-RELEASE" {
			events = append(events, event)
		}
	}
	corpus.ParticipantEvents = events
	if err := ValidateCorpus(corpus); err != nil {
		t.Fatalf("mutated corpus invalid: %v", err)
	}
	truths, err := DeriveTruths(corpus)
	if err != nil {
		t.Fatal(err)
	}
	truth := truthByID(t, truths, "T-206")
	if truth.PreparedStatus.Value != PreparedValid || truth.Disposition != DispositionApplyCommit {
		t.Fatalf("case identity selected breach after capability changed: %+v", truth)
	}
}

func TestCorpusRejectsIncompleteParticipantCut(t *testing.T) {
	corpus := cloneCorpus(loadTestCorpus(t))
	events := corpus.ParticipantEvents[:0]
	for _, event := range corpus.ParticipantEvents {
		if event.ID != "P203-I-CUT" {
			events = append(events, event)
		}
	}
	corpus.ParticipantEvents = events
	if err := ValidateCorpus(corpus); err == nil || !strings.Contains(err.Error(), "complete lifecycle cut") {
		t.Fatalf("expected incomplete-cut error, got %v", err)
	}
}

func TestReviewRejectsDuplicateSemanticCase(t *testing.T) {
	corpus := loadTestCorpus(t)
	review, err := LoadReview("../solution/adjudication.json")
	if err != nil {
		t.Fatal(err)
	}
	review.CaseReviews = append(review.CaseReviews, review.CaseReviews[0])
	evaluation, err := Evaluate(corpus, review)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(evaluation.Violations, "\n"), "duplicate case review") {
		t.Fatalf("duplicate semantic case was accepted: %v", evaluation.Violations)
	}
}
