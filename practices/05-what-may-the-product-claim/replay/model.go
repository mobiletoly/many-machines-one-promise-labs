package replay

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

const (
	StateNeverAttempted = "never_attempted"
	StateUnknown        = "unknown"
	StatePending        = "pending"
	StateAccepted       = "accepted"
	StateRejected       = "rejected"

	ObligationNotTriggered = "not_triggered"
	ObligationUnresolved   = "unresolved"
	ObligationSatisfied    = "satisfied"
	ObligationUnsatisfied  = "unsatisfied"

	VerdictSupported   = "supported"
	VerdictUnsupported = "unsupported"
	VerdictPass        = "pass"
	VerdictFail        = "fail"

	ReasonEstablished       = "declared_predicate_established"
	ReasonNotEstablished    = "required_predicate_not_established"
	ReasonContradicted      = "authoritative_fact_contradicts_claim"
	ReasonDomainIncomplete  = "satisfaction_domain_incomplete"
	ReasonNoSatisfyingFact  = "no_recognized_satisfying_outcome"
	ReasonObligationDiffers = "obligation_state_differs"

	AtomicNotRequired    = "not_required"
	AtomicNotEstablished = "not_established"
)

type Contract struct {
	ID                string            `json:"id"`
	EvidenceAuthority EvidenceAuthority `json:"evidence_authority"`
	Rules             []Rule            `json:"rules"`
	ReturnObligation  ReturnObligation  `json:"return_obligation"`
	ClaimSemantics    []ClaimSemantic   `json:"claim_semantics"`
	StrongerContract  StrongerContract  `json:"stronger_contract"`
}

type EvidenceAuthority struct {
	CaseOperationInventoryComplete           bool `json:"case_operation_inventory_complete"`
	SubmissionEvidenceCompleteThroughEachCut bool `json:"submission_evidence_complete_through_each_cut"`
	AuthorityStateEventsAuthoritative        bool `json:"authority_state_events_authoritative"`
}

type Rule struct {
	ID         string      `json:"id"`
	Kind       string      `json:"kind"`
	Role       string      `json:"role,omitempty"`
	State      string      `json:"state,omitempty"`
	Status     string      `json:"status,omitempty"`
	Conditions []Condition `json:"conditions,omitempty"`
}

type Condition struct {
	Role  string `json:"role"`
	State string `json:"state"`
}

type ReturnObligation struct {
	ID                                        string   `json:"id"`
	TriggerRuleID                             string   `json:"trigger_rule_id"`
	SatisfactionRoles                         []string `json:"satisfaction_roles"`
	SatisfyingState                           string   `json:"satisfying_state"`
	UnsatisfiedRequiresCompleteTerminalDomain bool     `json:"unsatisfied_requires_complete_terminal_domain"`
}

type ClaimSemantic struct {
	ID            string `json:"id"`
	AssertsRuleID string `json:"asserts_rule_id"`
}

type StrongerContract struct {
	ID                                               string            `json:"id"`
	ForbiddenTerminalSplits                          []TerminalOutcome `json:"forbidden_terminal_splits"`
	HealthyCapableParticipantsMustAllowCommit        bool              `json:"healthy_capable_participants_must_allow_commit"`
	LaterCorrectionSatisfiesOriginalTerminalContract bool              `json:"later_correction_satisfies_original_terminal_contract"`
}

type WorkflowDesign struct {
	Cases         []CaseDesign   `json:"cases"`
	Architectures []Architecture `json:"architectures"`
}

type CaseDesign struct {
	ID         string      `json:"id"`
	Operations []Operation `json:"operations"`
	Cuts       []Cut       `json:"cuts"`
}

type Operation struct {
	ID        string `json:"id"`
	Role      string `json:"role"`
	Authority string `json:"authority"`
}

type Cut struct {
	ID              string `json:"id"`
	ThroughSequence int    `json:"through_sequence"`
}

type Architecture struct {
	ID                                     string            `json:"id"`
	SingleAtomicAuthority                  bool              `json:"single_atomic_authority"`
	PreterminalCommonDecision              bool              `json:"preterminal_common_decision"`
	AlwaysAbort                            bool              `json:"always_abort"`
	HealthyExecutionAllParticipantsCapable bool              `json:"healthy_execution_all_participants_capable"`
	Participants                           []Participant     `json:"participants"`
	ReachableTerminalOutcomes              []TerminalOutcome `json:"reachable_terminal_outcomes"`
}

type Participant struct {
	Role       string `json:"role"`
	Authority  string `json:"authority"`
	CanPrepare bool   `json:"can_prepare"`
}

type TerminalOutcome struct {
	ID          string `json:"id,omitempty"`
	Capture     string `json:"capture"`
	Reservation string `json:"reservation"`
}

type EvidenceEvent struct {
	ID          string `json:"id"`
	CaseID      string `json:"case_id"`
	Sequence    int    `json:"sequence"`
	OperationID string `json:"operation_id"`
	Kind        string `json:"kind"`
	State       string `json:"state,omitempty"`
}

type ClaimCandidates struct {
	Proposals []ClaimProposal `json:"proposals"`
}

type ClaimProposal struct {
	CaseID   string   `json:"case_id"`
	CutID    string   `json:"cut_id"`
	ClaimIDs []string `json:"claim_ids"`
}

type Review struct {
	Name                   string                 `json:"name"`
	CutReviews             []CutReview            `json:"cut_reviews"`
	StrongerContractReview StrongerContractReview `json:"stronger_contract_review"`
}

type CutReview struct {
	CaseID               string                 `json:"case_id"`
	CutID                string                 `json:"cut_id"`
	AuthorityStates      []AuthorityStateReview `json:"authority_states"`
	AppliedContractRules []AppliedRuleReview    `json:"applied_contract_rules"`
	Obligation           ObligationReview       `json:"obligation"`
	ClaimReviews         []ClaimReview          `json:"claim_reviews"`
}

type AuthorityStateReview struct {
	OperationID string   `json:"operation_id"`
	State       string   `json:"state"`
	EvidenceIDs []string `json:"evidence_ids"`
}

type AppliedRuleReview struct {
	RuleID      string   `json:"rule_id"`
	EvidenceIDs []string `json:"evidence_ids"`
}

type ObligationReview struct {
	ID                     string   `json:"id"`
	Status                 string   `json:"status"`
	EvidenceDomainComplete bool     `json:"evidence_domain_complete"`
	EvidenceIDs            []string `json:"evidence_ids"`
}

type ClaimReview struct {
	ClaimID     string   `json:"claim_id"`
	Verdict     string   `json:"verdict"`
	Reason      string   `json:"reason"`
	EvidenceIDs []string `json:"evidence_ids"`
}

type StrongerContractReview struct {
	ArchitectureReviews                      []ArchitectureReview `json:"architecture_reviews"`
	LaterCorrectionSatisfiesOriginalContract string               `json:"later_correction_satisfies_original_contract"`
	TwoPhaseCommitSelected                   string               `json:"two_phase_commit_selected"`
}

type ArchitectureReview struct {
	ArchitectureID                  string   `json:"architecture_id"`
	ContractFit                     string   `json:"contract_fit"`
	Reasons                         []string `json:"reasons"`
	DistributedAtomicCommitRequired string   `json:"distributed_atomic_commit_required"`
}

type OperationTruth struct {
	Operation   Operation
	State       string
	EvidenceIDs []string
}

type RuleTruth struct {
	Rule          Rule
	Holds         bool
	EvidenceIDs   []string
	FailureReason string
}

type ObligationTruth struct {
	ID                     string
	Status                 string
	EvidenceDomainComplete bool
	EvidenceIDs            []string
}

type ClaimTruth struct {
	ClaimID     string
	Verdict     string
	Reason      string
	EvidenceIDs []string
}

type CutTruth struct {
	CaseID          string
	CutID           string
	OperationTruths []OperationTruth
	RuleTruths      []RuleTruth
	Obligation      ObligationTruth
	ClaimTruths     []ClaimTruth
}

type ArchitectureTruth struct {
	ArchitectureID                  string
	ContractFit                     string
	Reasons                         []string
	DistributedAtomicCommitRequired string
}

type Evaluation struct {
	Review             Review
	CutTruths          []CutTruth
	ArchitectureTruths []ArchitectureTruth
	Violations         []string
}

func LoadContract(path string) (Contract, error) {
	var contract Contract
	if err := loadStrictJSON(path, &contract); err != nil {
		return Contract{}, fmt.Errorf("load contract: %w", err)
	}
	if err := validateContract(contract); err != nil {
		return Contract{}, fmt.Errorf("validate contract: %w", err)
	}
	return contract, nil
}

func LoadWorkflowDesign(path string) (WorkflowDesign, error) {
	var design WorkflowDesign
	if err := loadStrictJSON(path, &design); err != nil {
		return WorkflowDesign{}, fmt.Errorf("load workflow design: %w", err)
	}
	if err := validateWorkflowDesign(design); err != nil {
		return WorkflowDesign{}, fmt.Errorf("validate workflow design: %w", err)
	}
	return design, nil
}

func LoadEvidence(path string, design WorkflowDesign) ([]EvidenceEvent, error) {
	events, err := loadStrictJSONL[EvidenceEvent](path)
	if err != nil {
		return nil, fmt.Errorf("load authority evidence: %w", err)
	}
	if err := validateEvidence(events, design); err != nil {
		return nil, fmt.Errorf("validate authority evidence: %w", err)
	}
	return events, nil
}

func LoadClaimCandidates(path string, contract Contract, design WorkflowDesign) (ClaimCandidates, error) {
	var candidates ClaimCandidates
	if err := loadStrictJSON(path, &candidates); err != nil {
		return ClaimCandidates{}, fmt.Errorf("load claim candidates: %w", err)
	}
	if err := validateClaimCandidates(candidates, contract, design); err != nil {
		return ClaimCandidates{}, fmt.Errorf("validate claim candidates: %w", err)
	}
	return candidates, nil
}

func LoadReview(path string) (Review, error) {
	var review Review
	if err := loadStrictJSON(path, &review); err != nil {
		return Review{}, fmt.Errorf("load review: %w", err)
	}
	if review.Name == "" {
		return Review{}, fmt.Errorf("load review: name is required")
	}
	return review, nil
}

func Evaluate(review Review, contract Contract, design WorkflowDesign, events []EvidenceEvent, candidates ClaimCandidates) Evaluation {
	cutTruths := deriveCutTruths(contract, design, events, candidates)
	architectureTruths := deriveArchitectureTruths(contract, design)
	violations := validateReview(review, contract, design, candidates, cutTruths, architectureTruths)
	return Evaluation{
		Review:             review,
		CutTruths:          cutTruths,
		ArchitectureTruths: architectureTruths,
		Violations:         violations,
	}
}

func deriveCutTruths(contract Contract, design WorkflowDesign, events []EvidenceEvent, candidates ClaimCandidates) []CutTruth {
	eventsByCase := map[string][]EvidenceEvent{}
	for _, event := range events {
		eventsByCase[event.CaseID] = append(eventsByCase[event.CaseID], event)
	}
	proposalByCut := map[string]ClaimProposal{}
	for _, proposal := range candidates.Proposals {
		proposalByCut[cutKey(proposal.CaseID, proposal.CutID)] = proposal
	}

	truths := []CutTruth{}
	for _, caseDesign := range design.Cases {
		for _, cut := range caseDesign.Cuts {
			operations := deriveOperationTruths(caseDesign, cut, eventsByCase[caseDesign.ID])
			baseRules := deriveBaseRuleTruths(contract.Rules, operations)
			obligation := deriveObligationTruth(contract.ReturnObligation, baseRules, operations)
			rules := deriveAllRuleTruths(contract.Rules, operations, obligation)
			proposal := proposalByCut[cutKey(caseDesign.ID, cut.ID)]
			claims := deriveClaimTruths(contract.ClaimSemantics, proposal, rules)
			truths = append(truths, CutTruth{
				CaseID:          caseDesign.ID,
				CutID:           cut.ID,
				OperationTruths: operations,
				RuleTruths:      rules,
				Obligation:      obligation,
				ClaimTruths:     claims,
			})
		}
	}
	return truths
}

func deriveOperationTruths(caseDesign CaseDesign, cut Cut, events []EvidenceEvent) []OperationTruth {
	truths := make([]OperationTruth, 0, len(caseDesign.Operations))
	for _, operation := range caseDesign.Operations {
		var submission *EvidenceEvent
		var latestState *EvidenceEvent
		for index := range events {
			event := events[index]
			if event.Sequence > cut.ThroughSequence || event.OperationID != operation.ID {
				continue
			}
			if event.Kind == "submitted" {
				copy := event
				submission = &copy
			}
			if event.Kind == "authority_state" {
				copy := event
				latestState = &copy
			}
		}
		truth := OperationTruth{Operation: operation, State: StateNeverAttempted}
		switch {
		case submission == nil:
			truth.EvidenceIDs = []string{}
		case latestState == nil:
			truth.State = StateUnknown
			truth.EvidenceIDs = []string{submission.ID}
		default:
			truth.State = latestState.State
			truth.EvidenceIDs = []string{latestState.ID}
		}
		truths = append(truths, truth)
	}
	return truths
}

func deriveBaseRuleTruths(rules []Rule, operations []OperationTruth) []RuleTruth {
	truths := make([]RuleTruth, 0, len(rules))
	for _, rule := range rules {
		if rule.Kind == "obligation_status" {
			continue
		}
		truths = append(truths, deriveOperationRuleTruth(rule, operations))
	}
	return truths
}

func deriveAllRuleTruths(rules []Rule, operations []OperationTruth, obligation ObligationTruth) []RuleTruth {
	truths := make([]RuleTruth, 0, len(rules))
	for _, rule := range rules {
		if rule.Kind != "obligation_status" {
			truths = append(truths, deriveOperationRuleTruth(rule, operations))
			continue
		}
		truth := RuleTruth{Rule: rule, Holds: obligation.Status == rule.Status, EvidenceIDs: cloneStrings(obligation.EvidenceIDs)}
		if truth.Holds {
			truth.FailureReason = ""
		} else {
			switch {
			case rule.Status == ObligationUnsatisfied && obligation.Status == ObligationUnresolved:
				truth.FailureReason = ReasonDomainIncomplete
			case rule.Status == ObligationSatisfied:
				truth.FailureReason = ReasonNoSatisfyingFact
			default:
				truth.FailureReason = ReasonObligationDiffers
			}
		}
		truths = append(truths, truth)
	}
	return truths
}

func deriveOperationRuleTruth(rule Rule, operations []OperationTruth) RuleTruth {
	conditions := rule.Conditions
	if rule.Kind == "operation_state" {
		conditions = []Condition{{Role: rule.Role, State: rule.State}}
	}
	truth := RuleTruth{Rule: rule, Holds: true, EvidenceIDs: []string{}}
	hasContradiction := false
	for _, condition := range conditions {
		operation := operationByRole(operations, condition.Role)
		if operation.State == condition.State {
			truth.EvidenceIDs = append(truth.EvidenceIDs, operation.EvidenceIDs...)
			continue
		}
		truth.Holds = false
		truth.EvidenceIDs = append(truth.EvidenceIDs, operation.EvidenceIDs...)
		if isTerminal(operation.State) {
			hasContradiction = true
		}
	}
	truth.EvidenceIDs = uniqueStrings(truth.EvidenceIDs)
	if !truth.Holds {
		truth.FailureReason = ReasonNotEstablished
		if hasContradiction {
			truth.FailureReason = ReasonContradicted
		}
	}
	return truth
}

func deriveObligationTruth(obligation ReturnObligation, baseRules []RuleTruth, operations []OperationTruth) ObligationTruth {
	trigger := ruleTruthByID(baseRules, obligation.TriggerRuleID)
	truth := ObligationTruth{
		ID:                     obligation.ID,
		Status:                 ObligationNotTriggered,
		EvidenceDomainComplete: false,
		EvidenceIDs:            cloneStrings(trigger.EvidenceIDs),
	}
	if !trigger.Holds {
		return truth
	}

	allTerminal := true
	for _, role := range obligation.SatisfactionRoles {
		operation := operationByRole(operations, role)
		truth.EvidenceIDs = append(truth.EvidenceIDs, operation.EvidenceIDs...)
		if operation.State == obligation.SatisfyingState {
			truth.Status = ObligationSatisfied
			truth.EvidenceDomainComplete = allSatisfactionRolesTerminal(operations, obligation.SatisfactionRoles)
			truth.EvidenceIDs = uniqueStrings(truth.EvidenceIDs)
			return truth
		}
		if !isTerminal(operation.State) {
			allTerminal = false
		}
	}
	truth.EvidenceIDs = uniqueStrings(truth.EvidenceIDs)
	if allTerminal && obligation.UnsatisfiedRequiresCompleteTerminalDomain {
		truth.Status = ObligationUnsatisfied
		truth.EvidenceDomainComplete = true
		return truth
	}
	truth.Status = ObligationUnresolved
	truth.EvidenceDomainComplete = false
	return truth
}

func deriveClaimTruths(semantics []ClaimSemantic, proposal ClaimProposal, rules []RuleTruth) []ClaimTruth {
	semanticByID := map[string]ClaimSemantic{}
	for _, semantic := range semantics {
		semanticByID[semantic.ID] = semantic
	}
	truths := make([]ClaimTruth, 0, len(proposal.ClaimIDs))
	for _, claimID := range proposal.ClaimIDs {
		semantic := semanticByID[claimID]
		rule := ruleTruthByID(rules, semantic.AssertsRuleID)
		truth := ClaimTruth{ClaimID: claimID, Verdict: VerdictUnsupported, Reason: rule.FailureReason, EvidenceIDs: cloneStrings(rule.EvidenceIDs)}
		if rule.Holds {
			truth.Verdict = VerdictSupported
			truth.Reason = ReasonEstablished
		}
		truths = append(truths, truth)
	}
	return truths
}

func deriveArchitectureTruths(contract Contract, design WorkflowDesign) []ArchitectureTruth {
	truths := make([]ArchitectureTruth, 0, len(design.Architectures))
	for _, architecture := range design.Architectures {
		reasons := []string{}
		for _, outcome := range architecture.ReachableTerminalOutcomes {
			if forbiddenOutcome(contract.StrongerContract.ForbiddenTerminalSplits, outcome) {
				reasons = append(reasons, "forbidden_mixed_terminal_outcome_reachable")
				break
			}
		}
		if architecture.PreterminalCommonDecision && !architecture.SingleAtomicAuthority {
			for _, participant := range architecture.Participants {
				if !participant.CanPrepare {
					reasons = append(reasons, "required_participant_cannot_prepare")
					break
				}
			}
		}
		if contract.StrongerContract.HealthyCapableParticipantsMustAllowCommit &&
			architecture.HealthyExecutionAllParticipantsCapable && architecture.AlwaysAbort {
			reasons = append(reasons, "healthy_success_progress_not_met")
		}
		fit := VerdictPass
		if len(reasons) != 0 {
			fit = VerdictFail
		}
		atomic := AtomicNotEstablished
		if architecture.SingleAtomicAuthority && fit == VerdictPass {
			atomic = AtomicNotRequired
		}
		truths = append(truths, ArchitectureTruth{
			ArchitectureID:                  architecture.ID,
			ContractFit:                     fit,
			Reasons:                         reasons,
			DistributedAtomicCommitRequired: atomic,
		})
	}
	return truths
}

func validateReview(review Review, contract Contract, design WorkflowDesign, candidates ClaimCandidates, cutTruths []CutTruth, architectureTruths []ArchitectureTruth) []string {
	violations := []string{}
	reviewByCut := map[string]CutReview{}
	for _, candidate := range review.CutReviews {
		key := cutKey(candidate.CaseID, candidate.CutID)
		if _, exists := reviewByCut[key]; exists {
			violations = append(violations, fmt.Sprintf("cut review %s appears more than once", key))
		}
		reviewByCut[key] = candidate
	}
	if len(reviewByCut) != len(cutTruths) {
		violations = append(violations, fmt.Sprintf("cut review count = %d, want %d", len(reviewByCut), len(cutTruths)))
	}
	for _, truth := range cutTruths {
		key := cutKey(truth.CaseID, truth.CutID)
		candidate, ok := reviewByCut[key]
		if !ok {
			violations = append(violations, fmt.Sprintf("cut review %s is missing", key))
			continue
		}
		violations = append(violations, validateCutReview(candidate, truth)...)
	}

	architectureByID := map[string]ArchitectureReview{}
	for _, candidate := range review.StrongerContractReview.ArchitectureReviews {
		if _, exists := architectureByID[candidate.ArchitectureID]; exists {
			violations = append(violations, fmt.Sprintf("architecture review %s appears more than once", candidate.ArchitectureID))
		}
		architectureByID[candidate.ArchitectureID] = candidate
	}
	if len(architectureByID) != len(architectureTruths) {
		violations = append(violations, fmt.Sprintf("architecture review count = %d, want %d", len(architectureByID), len(architectureTruths)))
	}
	for _, truth := range architectureTruths {
		candidate, ok := architectureByID[truth.ArchitectureID]
		if !ok {
			violations = append(violations, fmt.Sprintf("architecture review %s is missing", truth.ArchitectureID))
			continue
		}
		if candidate.ContractFit != truth.ContractFit {
			violations = append(violations, fmt.Sprintf("architecture %s contract_fit = %s, want %s", truth.ArchitectureID, candidate.ContractFit, truth.ContractFit))
		}
		if !sameStringSet(candidate.Reasons, truth.Reasons) {
			violations = append(violations, fmt.Sprintf("architecture %s reasons do not match the declared contract", truth.ArchitectureID))
		}
		if candidate.DistributedAtomicCommitRequired != truth.DistributedAtomicCommitRequired {
			violations = append(violations, fmt.Sprintf("architecture %s distributed atomic commit verdict = %s, want %s", truth.ArchitectureID, candidate.DistributedAtomicCommitRequired, truth.DistributedAtomicCommitRequired))
		}
	}
	if review.StrongerContractReview.LaterCorrectionSatisfiesOriginalContract != VerdictUnsupported {
		violations = append(violations, "later correction cannot satisfy the original all-or-nothing terminal contract")
	}
	if review.StrongerContractReview.TwoPhaseCommitSelected != VerdictUnsupported {
		violations = append(violations, "the dossier does not select two-phase commit")
	}
	return violations
}

func validateCutReview(candidate CutReview, truth CutTruth) []string {
	violations := []string{}
	prefix := cutKey(truth.CaseID, truth.CutID)
	stateByOperation := map[string]AuthorityStateReview{}
	for _, state := range candidate.AuthorityStates {
		if _, exists := stateByOperation[state.OperationID]; exists {
			violations = append(violations, fmt.Sprintf("%s operation %s appears more than once", prefix, state.OperationID))
		}
		stateByOperation[state.OperationID] = state
	}
	if len(stateByOperation) != len(truth.OperationTruths) {
		violations = append(violations, fmt.Sprintf("%s authority state count = %d, want %d", prefix, len(stateByOperation), len(truth.OperationTruths)))
	}
	for _, expected := range truth.OperationTruths {
		candidateState, ok := stateByOperation[expected.Operation.ID]
		if !ok {
			violations = append(violations, fmt.Sprintf("%s operation %s is missing", prefix, expected.Operation.ID))
			continue
		}
		if candidateState.State != expected.State {
			violations = append(violations, fmt.Sprintf("%s operation %s state = %s, want %s", prefix, expected.Operation.ID, candidateState.State, expected.State))
		}
		if !sameStringSet(candidateState.EvidenceIDs, expected.EvidenceIDs) {
			violations = append(violations, fmt.Sprintf("%s operation %s evidence does not support its state", prefix, expected.Operation.ID))
		}
	}

	appliedByID := map[string]AppliedRuleReview{}
	for _, applied := range candidate.AppliedContractRules {
		if _, exists := appliedByID[applied.RuleID]; exists {
			violations = append(violations, fmt.Sprintf("%s rule %s appears more than once", prefix, applied.RuleID))
		}
		appliedByID[applied.RuleID] = applied
	}
	expectedApplied := 0
	for _, rule := range truth.RuleTruths {
		if !rule.Holds {
			continue
		}
		expectedApplied++
		candidateRule, ok := appliedByID[rule.Rule.ID]
		if !ok {
			violations = append(violations, fmt.Sprintf("%s supported rule %s is missing", prefix, rule.Rule.ID))
			continue
		}
		if !sameStringSet(candidateRule.EvidenceIDs, rule.EvidenceIDs) {
			violations = append(violations, fmt.Sprintf("%s rule %s evidence does not match", prefix, rule.Rule.ID))
		}
	}
	if len(appliedByID) != expectedApplied {
		violations = append(violations, fmt.Sprintf("%s applied rule count = %d, want %d", prefix, len(appliedByID), expectedApplied))
	}

	if candidate.Obligation.ID != truth.Obligation.ID {
		violations = append(violations, fmt.Sprintf("%s obligation id = %s, want %s", prefix, candidate.Obligation.ID, truth.Obligation.ID))
	}
	if candidate.Obligation.Status != truth.Obligation.Status {
		violations = append(violations, fmt.Sprintf("%s obligation status = %s, want %s", prefix, candidate.Obligation.Status, truth.Obligation.Status))
	}
	if candidate.Obligation.EvidenceDomainComplete != truth.Obligation.EvidenceDomainComplete {
		violations = append(violations, fmt.Sprintf("%s obligation evidence completeness is incorrect", prefix))
	}
	if !sameStringSet(candidate.Obligation.EvidenceIDs, truth.Obligation.EvidenceIDs) {
		violations = append(violations, fmt.Sprintf("%s obligation evidence does not match", prefix))
	}

	claimByID := map[string]ClaimReview{}
	for _, claim := range candidate.ClaimReviews {
		if _, exists := claimByID[claim.ClaimID]; exists {
			violations = append(violations, fmt.Sprintf("%s claim %s appears more than once", prefix, claim.ClaimID))
		}
		claimByID[claim.ClaimID] = claim
	}
	if len(claimByID) != len(truth.ClaimTruths) {
		violations = append(violations, fmt.Sprintf("%s claim review count = %d, want %d", prefix, len(claimByID), len(truth.ClaimTruths)))
	}
	for _, expected := range truth.ClaimTruths {
		candidateClaim, ok := claimByID[expected.ClaimID]
		if !ok {
			violations = append(violations, fmt.Sprintf("%s claim %s is missing", prefix, expected.ClaimID))
			continue
		}
		if candidateClaim.Verdict != expected.Verdict || candidateClaim.Reason != expected.Reason {
			violations = append(violations, fmt.Sprintf("%s claim %s verdict or reason is unsupported", prefix, expected.ClaimID))
		}
		if !sameStringSet(candidateClaim.EvidenceIDs, expected.EvidenceIDs) {
			violations = append(violations, fmt.Sprintf("%s claim %s evidence does not match", prefix, expected.ClaimID))
		}
	}
	return violations
}

func validateContract(contract Contract) error {
	if contract.ID == "" || !contract.EvidenceAuthority.CaseOperationInventoryComplete ||
		!contract.EvidenceAuthority.SubmissionEvidenceCompleteThroughEachCut ||
		!contract.EvidenceAuthority.AuthorityStateEventsAuthoritative {
		return fmt.Errorf("contract must name complete operation, submission, and authority-state evidence")
	}
	ruleIDs := map[string]bool{}
	validRoles := map[string]bool{"capture": true, "reservation": true, "return_primary": true, "return_alternative": true}
	for _, rule := range contract.Rules {
		if rule.ID == "" || ruleIDs[rule.ID] {
			return fmt.Errorf("rule ids must be unique and non-empty")
		}
		ruleIDs[rule.ID] = true
		switch rule.Kind {
		case "operation_state":
			if !validRoles[rule.Role] || !validState(rule.State) {
				return fmt.Errorf("rule %s has an invalid role or state", rule.ID)
			}
		case "all_operation_states":
			if len(rule.Conditions) == 0 {
				return fmt.Errorf("rule %s has no conditions", rule.ID)
			}
			for _, condition := range rule.Conditions {
				if !validRoles[condition.Role] || !validState(condition.State) {
					return fmt.Errorf("rule %s has an invalid condition", rule.ID)
				}
			}
		case "obligation_status":
			if !validObligationStatus(rule.Status) {
				return fmt.Errorf("rule %s has an invalid obligation status", rule.ID)
			}
		default:
			return fmt.Errorf("rule %s has unsupported kind %s", rule.ID, rule.Kind)
		}
	}
	if contract.ReturnObligation.ID == "" || !ruleIDs[contract.ReturnObligation.TriggerRuleID] ||
		len(contract.ReturnObligation.SatisfactionRoles) == 0 || contract.ReturnObligation.SatisfyingState != StateAccepted ||
		!contract.ReturnObligation.UnsatisfiedRequiresCompleteTerminalDomain {
		return fmt.Errorf("return obligation contract is incomplete")
	}
	claimIDs := map[string]bool{}
	for _, semantic := range contract.ClaimSemantics {
		if semantic.ID == "" || claimIDs[semantic.ID] || !ruleIDs[semantic.AssertsRuleID] {
			return fmt.Errorf("claim semantics must be unique and reference declared rules")
		}
		claimIDs[semantic.ID] = true
	}
	if contract.StrongerContract.ID == "" || len(contract.StrongerContract.ForbiddenTerminalSplits) == 0 ||
		!contract.StrongerContract.HealthyCapableParticipantsMustAllowCommit ||
		contract.StrongerContract.LaterCorrectionSatisfiesOriginalTerminalContract {
		return fmt.Errorf("stronger contract must forbid mixed outcomes, require healthy progress, and reject later correction as original atomicity")
	}
	return nil
}

func validateWorkflowDesign(design WorkflowDesign) error {
	if len(design.Cases) == 0 || len(design.Architectures) == 0 {
		return fmt.Errorf("cases and architectures are required")
	}
	caseIDs := map[string]bool{}
	for _, candidate := range design.Cases {
		if candidate.ID == "" || caseIDs[candidate.ID] {
			return fmt.Errorf("case ids must be unique and non-empty")
		}
		caseIDs[candidate.ID] = true
		if len(candidate.Operations) != 4 || len(candidate.Cuts) == 0 {
			return fmt.Errorf("case %s must declare four operations and at least one cut", candidate.ID)
		}
		roles := map[string]bool{}
		operationIDs := map[string]bool{}
		for _, operation := range candidate.Operations {
			if operation.ID == "" || operation.Authority == "" || operationIDs[operation.ID] || roles[operation.Role] {
				return fmt.Errorf("case %s operation inventory is invalid", candidate.ID)
			}
			operationIDs[operation.ID] = true
			roles[operation.Role] = true
		}
		for _, role := range []string{"capture", "reservation", "return_primary", "return_alternative"} {
			if !roles[role] {
				return fmt.Errorf("case %s lacks role %s", candidate.ID, role)
			}
		}
		cutIDs := map[string]bool{}
		lastSequence := 0
		for _, cut := range candidate.Cuts {
			if cut.ID == "" || cutIDs[cut.ID] || cut.ThroughSequence <= lastSequence {
				return fmt.Errorf("case %s cuts must be unique and increasing", candidate.ID)
			}
			cutIDs[cut.ID] = true
			lastSequence = cut.ThroughSequence
		}
	}
	architectureIDs := map[string]bool{}
	for _, architecture := range design.Architectures {
		if architecture.ID == "" || architectureIDs[architecture.ID] || len(architecture.Participants) != 2 {
			return fmt.Errorf("architecture ids must be unique and each design must name two participants")
		}
		architectureIDs[architecture.ID] = true
	}
	return nil
}

func validateEvidence(events []EvidenceEvent, design WorkflowDesign) error {
	caseByID := map[string]CaseDesign{}
	for _, candidate := range design.Cases {
		caseByID[candidate.ID] = candidate
	}
	eventIDs := map[string]bool{}
	sequenceByCase := map[string]map[int]bool{}
	terminalByOperation := map[string]bool{}
	submitted := map[string]bool{}
	for _, event := range events {
		candidate, ok := caseByID[event.CaseID]
		if !ok || event.ID == "" || eventIDs[event.ID] || event.Sequence <= 0 {
			return fmt.Errorf("evidence event identity is invalid")
		}
		eventIDs[event.ID] = true
		if sequenceByCase[event.CaseID] == nil {
			sequenceByCase[event.CaseID] = map[int]bool{}
		}
		if sequenceByCase[event.CaseID][event.Sequence] {
			return fmt.Errorf("case %s repeats sequence %d", event.CaseID, event.Sequence)
		}
		sequenceByCase[event.CaseID][event.Sequence] = true
		if !operationExists(candidate, event.OperationID) {
			return fmt.Errorf("event %s references unknown operation %s", event.ID, event.OperationID)
		}
		key := event.CaseID + "/" + event.OperationID
		switch event.Kind {
		case "submitted":
			if event.State != "" || submitted[key] {
				return fmt.Errorf("event %s has invalid submission evidence", event.ID)
			}
			submitted[key] = true
		case "authority_state":
			if !submitted[key] || !validAuthorityState(event.State) || terminalByOperation[key] {
				return fmt.Errorf("event %s has invalid authority-state transition", event.ID)
			}
			if isTerminal(event.State) {
				terminalByOperation[key] = true
			}
		default:
			return fmt.Errorf("event %s has unsupported kind %s", event.ID, event.Kind)
		}
	}
	for _, candidate := range design.Cases {
		maxCut := candidate.Cuts[len(candidate.Cuts)-1].ThroughSequence
		for sequence := 1; sequence <= maxCut; sequence++ {
			if !sequenceByCase[candidate.ID][sequence] {
				return fmt.Errorf("case %s lacks evidence sequence %d", candidate.ID, sequence)
			}
		}
	}
	return nil
}

func validateClaimCandidates(candidates ClaimCandidates, contract Contract, design WorkflowDesign) error {
	claimIDs := map[string]bool{}
	for _, semantic := range contract.ClaimSemantics {
		claimIDs[semantic.ID] = true
	}
	validCuts := map[string]bool{}
	for _, candidate := range design.Cases {
		for _, cut := range candidate.Cuts {
			validCuts[cutKey(candidate.ID, cut.ID)] = true
		}
	}
	seen := map[string]bool{}
	for _, proposal := range candidates.Proposals {
		key := cutKey(proposal.CaseID, proposal.CutID)
		if !validCuts[key] || seen[key] || len(proposal.ClaimIDs) == 0 {
			return fmt.Errorf("claim proposal %s is invalid", key)
		}
		seen[key] = true
		claimsInProposal := map[string]bool{}
		for _, claimID := range proposal.ClaimIDs {
			if !claimIDs[claimID] || claimsInProposal[claimID] {
				return fmt.Errorf("claim proposal %s contains invalid claim %s", key, claimID)
			}
			claimsInProposal[claimID] = true
		}
	}
	if len(seen) != len(validCuts) {
		return fmt.Errorf("every cut must have one claim proposal")
	}
	return nil
}

func loadStrictJSON(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("unexpected trailing JSON value")
		}
		return fmt.Errorf("parse trailing JSON: %w", err)
	}
	return nil
}

func loadStrictJSONL[T any](path string) ([]T, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	items := []T{}
	scanner := bufio.NewScanner(file)
	line := 0
	for scanner.Scan() {
		line++
		if strings.TrimSpace(scanner.Text()) == "" {
			return nil, fmt.Errorf("line %d is empty", line)
		}
		decoder := json.NewDecoder(strings.NewReader(scanner.Text()))
		decoder.DisallowUnknownFields()
		var item T
		if err := decoder.Decode(&item); err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}
		var trailing any
		if err := decoder.Decode(&trailing); err != io.EOF {
			if err == nil {
				return nil, fmt.Errorf("line %d has trailing JSON", line)
			}
			return nil, fmt.Errorf("line %d trailing JSON: %w", line, err)
		}
		items = append(items, item)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("input is empty")
	}
	return items, nil
}

func operationByRole(operations []OperationTruth, role string) OperationTruth {
	for _, operation := range operations {
		if operation.Operation.Role == role {
			return operation
		}
	}
	panic("validated role missing: " + role)
}

func ruleTruthByID(rules []RuleTruth, id string) RuleTruth {
	for _, rule := range rules {
		if rule.Rule.ID == id {
			return rule
		}
	}
	panic("validated rule missing: " + id)
}

func allSatisfactionRolesTerminal(operations []OperationTruth, roles []string) bool {
	for _, role := range roles {
		if !isTerminal(operationByRole(operations, role).State) {
			return false
		}
	}
	return true
}

func forbiddenOutcome(forbidden []TerminalOutcome, candidate TerminalOutcome) bool {
	for _, outcome := range forbidden {
		if outcome.Capture == candidate.Capture && outcome.Reservation == candidate.Reservation {
			return true
		}
	}
	return false
}

func operationExists(candidate CaseDesign, id string) bool {
	for _, operation := range candidate.Operations {
		if operation.ID == id {
			return true
		}
	}
	return false
}

func validState(state string) bool {
	return state == StateAccepted || state == StateRejected
}

func validAuthorityState(state string) bool {
	return state == StatePending || state == StateAccepted || state == StateRejected
}

func validObligationStatus(status string) bool {
	return status == ObligationUnresolved || status == ObligationSatisfied || status == ObligationUnsatisfied
}

func isTerminal(state string) bool {
	return state == StateAccepted || state == StateRejected
}

func cutKey(caseID, cutID string) string {
	return caseID + "/" + cutID
}

func cloneStrings(values []string) []string {
	return append([]string(nil), values...)
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}

func sameStringSet(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	a := cloneStrings(left)
	b := cloneStrings(right)
	sort.Strings(a)
	sort.Strings(b)
	for index := range a {
		if a[index] != b[index] {
			return false
		}
	}
	return true
}
