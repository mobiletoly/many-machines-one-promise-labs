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
	"time"
)

const (
	VerdictPass        = "pass"
	VerdictFail        = "fail"
	ClaimUnsupported   = "unsupported"
	RelationAfter      = "after_established_boundary"
	RelationAt         = "at_established_boundary"
	RelationNotSeen    = "not_observed"
	ReasonGeneric      = "generic_signal"
	ReasonState        = "state_inadmissible"
	ReasonInterpret    = "interpretation_invalid"
	ReasonPath         = "path_ineligible"
	ReasonResult       = "result_invalid"
	EventConfiguration = "serving_configuration_established"
	EventRequest       = "request_succeeded"
)

type Contract struct {
	Capability                CapabilityContract        `json:"capability"`
	Disruption                DisruptionContract        `json:"disruption"`
	AcceptedHistoryAuthority  AcceptedHistoryAuthority  `json:"accepted_history_authority"`
	RecoveryEvidenceAuthority RecoveryEvidenceAuthority `json:"recovery_evidence_authority"`
	ServingEvidenceAuthority  ServingEvidenceAuthority  `json:"serving_evidence_authority"`
	ProtectedAcknowledgement  ProtectedAcknowledgement  `json:"protected_acknowledgement"`
}

type CapabilityContract struct {
	ID                    string                `json:"id"`
	RequiredResultMeaning string                `json:"required_result_meaning"`
	EligiblePathIDs       []string              `json:"eligible_path_ids"`
	BuildInterpretations  []BuildInterpretation `json:"build_interpretations"`
}

type BuildInterpretation struct {
	BuildID    string `json:"build_id"`
	StatusCode int    `json:"status_code"`
	Meaning    string `json:"meaning"`
}

type DisruptionContract struct {
	At         string `json:"at"`
	RPOSeconds int    `json:"rpo_seconds"`
	RTOSeconds int    `json:"rto_seconds"`
}

type AcceptedHistoryAuthority struct {
	ID                        string `json:"id"`
	CompleteThroughDisruption bool   `json:"complete_through_disruption"`
	OrderField                string `json:"order_field"`
}

type RecoveryEvidenceAuthority struct {
	ID                        string `json:"id"`
	CompleteMembershipByState bool   `json:"complete_membership_by_state"`
}

type ServingEvidenceAuthority struct {
	ID                                                        string `json:"id"`
	CapabilityUsableEventKind                                 string `json:"capability_usable_event_kind"`
	RequiresAdmissibleStateValidInterpretationAndEligiblePath bool   `json:"requires_admissible_state_valid_interpretation_and_eligible_path"`
}

type ProtectedAcknowledgement struct {
	OperationClass string `json:"operation_class"`
	Result         string `json:"result"`
}

type AcceptedFact struct {
	ID                    string `json:"id"`
	Position              int    `json:"position"`
	AcceptedAt            string `json:"accepted_at"`
	OperationID           string `json:"operation_id"`
	OperationClass        string `json:"operation_class"`
	AcknowledgementResult string `json:"acknowledgement_result"`
	StatusCode            *int   `json:"status_code,omitempty"`
}

type RecoveredState struct {
	EvidenceID       string   `json:"evidence_id"`
	StateID          string   `json:"state_id"`
	AccountedFactIDs []string `json:"accounted_fact_ids"`
	StatusCode       int      `json:"status_code"`
	StatusCodeFactID string   `json:"status_code_fact_id"`
}

type ServingEvent struct {
	ID            string `json:"id"`
	At            string `json:"at"`
	Kind          string `json:"kind"`
	StateID       string `json:"state_id"`
	BuildID       string `json:"build_id"`
	PathID        string `json:"path_id"`
	ResultMeaning string `json:"result_meaning"`
	Audience      string `json:"audience"`
}

type Decision struct {
	Name         string        `json:"name"`
	StateReviews []StateReview `json:"state_reviews"`
	RTOReview    RTOReview     `json:"rto_review"`
	Claims       Claims        `json:"claims"`
}

type StateReview struct {
	StateID                        string   `json:"state_id"`
	LatestCompleteFactID           string   `json:"latest_complete_fact_id"`
	CompletePrefixFactIDs          []string `json:"complete_prefix_fact_ids"`
	MembershipEvidenceID           string   `json:"membership_evidence_id"`
	PreservedProtectedOperationIDs []string `json:"preserved_protected_operation_ids"`
	MissingProtectedOperationIDs   []string `json:"missing_protected_operation_ids"`
	RPOVerdict                     string   `json:"rpo_verdict"`
	AcknowledgementVerdict         string   `json:"acknowledgement_verdict"`
	AdmissibilityVerdict           string   `json:"admissibility_verdict"`
}

type RTOReview struct {
	EndBoundaryEventID          string          `json:"end_boundary_event_id"`
	EndBoundaryAt               string          `json:"end_boundary_at"`
	StateID                     string          `json:"state_id"`
	BuildID                     string          `json:"build_id"`
	PathID                      string          `json:"path_id"`
	EvidenceIDs                 []string        `json:"evidence_ids"`
	RejectedEarlierEvents       []RejectedEvent `json:"rejected_earlier_events"`
	FirstCustomerResultEventID  string          `json:"first_customer_result_event_id"`
	FirstCustomerResultRelation string          `json:"first_customer_result_relation"`
	RTOVerdict                  string          `json:"rto_verdict"`
}

type RejectedEvent struct {
	EventID string   `json:"event_id"`
	Reasons []string `json:"reasons"`
}

type Claims struct {
	AllCapabilitiesRecovered         string `json:"all_capabilities_recovered"`
	FutureDisruptionsMeetObjectives  string `json:"future_disruptions_meet_objectives"`
	RPOSelectsRecoveryMechanism      string `json:"rpo_selects_recovery_mechanism"`
	FirstCustomerResultDefinesRTOEnd string `json:"first_customer_result_defines_rto_end"`
}

type StateTruth struct {
	StateID                        string
	EvidenceID                     string
	LatestCompleteFactID           string
	CompletePrefixFactIDs          []string
	PreservedProtectedOperationIDs []string
	MissingProtectedOperationIDs   []string
	RPOPass                        bool
	AcknowledgementPass            bool
	Admissible                     bool
}

type EventTruth struct {
	Event     ServingEvent
	Reasons   []string
	Qualifies bool
}

type RTOTruth struct {
	BoundaryEvent         ServingEvent
	RequiredEvidenceIDs   []string
	RejectedEarlier       []RejectedEvent
	FirstCustomerResult   ServingEvent
	FirstCustomerRelation string
	Pass                  bool
}

type Evaluation struct {
	Decision   Decision
	StateTruth map[string]StateTruth
	RTOTruth   RTOTruth
	Violations []string
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

func LoadAcceptedHistory(path string) ([]AcceptedFact, error) {
	facts, err := loadStrictJSONL[AcceptedFact](path)
	if err != nil {
		return nil, fmt.Errorf("load accepted history: %w", err)
	}
	if err := validateAcceptedHistory(facts); err != nil {
		return nil, fmt.Errorf("validate accepted history: %w", err)
	}
	return facts, nil
}

func LoadRecoveryEvidence(path string, facts []AcceptedFact) ([]RecoveredState, error) {
	states, err := loadStrictJSONL[RecoveredState](path)
	if err != nil {
		return nil, fmt.Errorf("load recovery evidence: %w", err)
	}
	if err := validateRecoveryEvidence(states, facts); err != nil {
		return nil, fmt.Errorf("validate recovery evidence: %w", err)
	}
	return states, nil
}

func LoadServingEvents(path string, states []RecoveredState) ([]ServingEvent, error) {
	events, err := loadStrictJSONL[ServingEvent](path)
	if err != nil {
		return nil, fmt.Errorf("load serving events: %w", err)
	}
	if err := validateServingEvents(events, states); err != nil {
		return nil, fmt.Errorf("validate serving events: %w", err)
	}
	return events, nil
}

func LoadDecision(path string) (Decision, error) {
	var decision Decision
	if err := loadStrictJSON(path, &decision); err != nil {
		return Decision{}, fmt.Errorf("load decision: %w", err)
	}
	if decision.Name == "" || len(decision.StateReviews) == 0 {
		return Decision{}, fmt.Errorf("load decision: name and state_reviews are required")
	}
	return decision, nil
}

func Evaluate(decision Decision, contract Contract, facts []AcceptedFact, states []RecoveredState, events []ServingEvent) Evaluation {
	if violations := validateDossierTimes(contract, facts, events); len(violations) != 0 {
		return Evaluation{Decision: decision, Violations: violations}
	}
	stateTruth := make(map[string]StateTruth, len(states))
	stateByID := make(map[string]RecoveredState, len(states))
	for _, state := range states {
		stateByID[state.StateID] = state
		stateTruth[state.StateID] = deriveStateTruth(contract, facts, state)
	}
	rtoTruth := deriveRTOTruth(contract, facts, stateByID, stateTruth, events)
	violations := validateDecision(decision, facts, states, events, stateTruth, rtoTruth)
	return Evaluation{Decision: decision, StateTruth: stateTruth, RTOTruth: rtoTruth, Violations: violations}
}

func validateDossierTimes(contract Contract, facts []AcceptedFact, events []ServingEvent) []string {
	disruption := mustTime(contract.Disruption.At)
	violations := []string{}
	for _, fact := range facts {
		if mustTime(fact.AcceptedAt).After(disruption) {
			violations = append(violations, fmt.Sprintf("accepted fact %s occurs after the disruption boundary", fact.ID))
		}
	}
	for _, event := range events {
		if mustTime(event.At).Before(disruption) {
			violations = append(violations, fmt.Sprintf("serving event %s occurs before the disruption boundary", event.ID))
		}
	}
	return violations
}

func deriveStateTruth(contract Contract, facts []AcceptedFact, state RecoveredState) StateTruth {
	accounted := stringSet(state.AccountedFactIDs)
	prefix := make([]string, 0, len(facts))
	latestID := ""
	for _, fact := range facts {
		if !accounted[fact.ID] {
			break
		}
		prefix = append(prefix, fact.ID)
		latestID = fact.ID
	}

	preserved := []string{}
	missing := []string{}
	for _, fact := range facts {
		if fact.OperationClass != contract.ProtectedAcknowledgement.OperationClass ||
			fact.AcknowledgementResult != contract.ProtectedAcknowledgement.Result {
			continue
		}
		if accounted[fact.ID] {
			preserved = append(preserved, fact.OperationID)
		} else {
			missing = append(missing, fact.OperationID)
		}
	}

	rpoFloor := mustTime(contract.Disruption.At).Add(-time.Duration(contract.Disruption.RPOSeconds) * time.Second)
	rpoPass := true
	for _, fact := range facts {
		if mustTime(fact.AcceptedAt).After(rpoFloor) {
			break
		}
		if !accounted[fact.ID] {
			rpoPass = false
			break
		}
	}
	ackPass := len(missing) == 0
	return StateTruth{
		StateID: state.StateID, EvidenceID: state.EvidenceID,
		LatestCompleteFactID: latestID, CompletePrefixFactIDs: prefix,
		PreservedProtectedOperationIDs: preserved, MissingProtectedOperationIDs: missing,
		RPOPass: rpoPass, AcknowledgementPass: ackPass, Admissible: rpoPass && ackPass,
	}
}

func deriveRTOTruth(contract Contract, facts []AcceptedFact, states map[string]RecoveredState, truths map[string]StateTruth, events []ServingEvent) RTOTruth {
	eventTruths := make([]EventTruth, 0, len(events))
	var boundary ServingEvent
	var firstCustomer ServingEvent
	for _, event := range events {
		truth := deriveEventTruth(contract, states, truths, event)
		eventTruths = append(eventTruths, truth)
		if boundary.ID == "" && truth.Qualifies {
			boundary = event
		}
		if firstCustomer.ID == "" && event.Kind == EventRequest && event.Audience == "customer" && truth.Qualifies {
			firstCustomer = event
		}
	}

	rejected := []RejectedEvent{}
	for _, truth := range eventTruths {
		if boundary.ID != "" && !mustTime(truth.Event.At).Before(mustTime(boundary.At)) {
			break
		}
		if !truth.Qualifies {
			rejected = append(rejected, RejectedEvent{EventID: truth.Event.ID, Reasons: truth.Reasons})
		}
	}

	requiredEvidence := []string{}
	if boundary.ID != "" {
		state := states[boundary.StateID]
		stateTruth := truths[boundary.StateID]
		requiredEvidence = append(requiredEvidence, boundary.ID, state.EvidenceID)
		requiredEvidence = append(requiredEvidence, stateTruth.CompletePrefixFactIDs...)
		protectedFactByOperation := make(map[string]string)
		for _, fact := range facts {
			protectedFactByOperation[fact.OperationID] = fact.ID
		}
		for _, operationID := range stateTruth.PreservedProtectedOperationIDs {
			requiredEvidence = append(requiredEvidence, protectedFactByOperation[operationID])
		}
		requiredEvidence = uniqueStrings(requiredEvidence)
	}

	relation := RelationNotSeen
	if firstCustomer.ID != "" && boundary.ID != "" {
		if mustTime(firstCustomer.At).Equal(mustTime(boundary.At)) {
			relation = RelationAt
		} else if mustTime(firstCustomer.At).After(mustTime(boundary.At)) {
			relation = RelationAfter
		}
	}
	deadline := mustTime(contract.Disruption.At).Add(time.Duration(contract.Disruption.RTOSeconds) * time.Second)
	pass := boundary.ID != "" && !mustTime(boundary.At).After(deadline)
	return RTOTruth{
		BoundaryEvent: boundary, RequiredEvidenceIDs: requiredEvidence,
		RejectedEarlier: rejected, FirstCustomerResult: firstCustomer,
		FirstCustomerRelation: relation, Pass: pass,
	}
}

func deriveEventTruth(contract Contract, states map[string]RecoveredState, truths map[string]StateTruth, event ServingEvent) EventTruth {
	if event.Kind != contract.ServingEvidenceAuthority.CapabilityUsableEventKind && event.Kind != EventRequest {
		return EventTruth{Event: event, Reasons: []string{ReasonGeneric}}
	}
	reasons := []string{}
	state, stateExists := states[event.StateID]
	stateTruth, truthExists := truths[event.StateID]
	if !stateExists || !truthExists || !stateTruth.Admissible {
		reasons = append(reasons, ReasonState)
	}
	if stateExists && interpretation(contract, event.BuildID, state.StatusCode) != contract.Capability.RequiredResultMeaning {
		reasons = append(reasons, ReasonInterpret)
	}
	if !contains(contract.Capability.EligiblePathIDs, event.PathID) {
		reasons = append(reasons, ReasonPath)
	}
	if event.Kind == EventRequest && event.ResultMeaning != contract.Capability.RequiredResultMeaning {
		reasons = append(reasons, ReasonResult)
	}
	return EventTruth{Event: event, Reasons: reasons, Qualifies: len(reasons) == 0}
}

func validateDecision(decision Decision, facts []AcceptedFact, states []RecoveredState, events []ServingEvent, truths map[string]StateTruth, rto RTOTruth) []string {
	violations := []string{}
	reviews := make(map[string]StateReview, len(decision.StateReviews))
	for _, review := range decision.StateReviews {
		if _, exists := reviews[review.StateID]; exists {
			violations = append(violations, fmt.Sprintf("state %s is reviewed more than once", review.StateID))
			continue
		}
		reviews[review.StateID] = review
	}
	if len(reviews) != len(truths) {
		violations = append(violations, "state review set does not match the recovery evidence")
	}
	for _, state := range states {
		truth := truths[state.StateID]
		review, ok := reviews[state.StateID]
		if !ok {
			violations = append(violations, fmt.Sprintf("state %s has no review", state.StateID))
			continue
		}
		prefix := state.StateID + ": "
		if review.LatestCompleteFactID != truth.LatestCompleteFactID {
			violations = append(violations, prefix+"latest complete cut is not supported")
		}
		if !sameStringSet(review.CompletePrefixFactIDs, truth.CompletePrefixFactIDs) {
			violations = append(violations, prefix+"complete-prefix evidence is incorrect")
		}
		if review.MembershipEvidenceID != truth.EvidenceID {
			violations = append(violations, prefix+"membership evidence is incorrect")
		}
		if !sameStringSet(review.PreservedProtectedOperationIDs, truth.PreservedProtectedOperationIDs) {
			violations = append(violations, prefix+"preserved protected operations are incorrect")
		}
		if !sameStringSet(review.MissingProtectedOperationIDs, truth.MissingProtectedOperationIDs) {
			violations = append(violations, prefix+"missing protected operations are incorrect")
		}
		if review.RPOVerdict != verdict(truth.RPOPass) {
			violations = append(violations, prefix+"RPO verdict is incorrect")
		}
		if review.AcknowledgementVerdict != verdict(truth.AcknowledgementPass) {
			violations = append(violations, prefix+"acknowledgement verdict is incorrect")
		}
		if review.AdmissibilityVerdict != verdict(truth.Admissible) {
			violations = append(violations, prefix+"admissibility verdict is incorrect")
		}
	}

	knownEvidence := map[string]bool{}
	for _, fact := range facts {
		knownEvidence[fact.ID] = true
	}
	for _, state := range states {
		knownEvidence[state.EvidenceID] = true
	}
	for _, event := range events {
		knownEvidence[event.ID] = true
	}
	for _, evidenceID := range decision.RTOReview.EvidenceIDs {
		if !knownEvidence[evidenceID] {
			violations = append(violations, fmt.Sprintf("RTO review cites unknown evidence %s", evidenceID))
		}
	}
	if rto.BoundaryEvent.ID == "" {
		violations = append(violations, "the dossier contains no evidence-supported recovery boundary")
	} else {
		review := decision.RTOReview
		if review.EndBoundaryEventID != rto.BoundaryEvent.ID || review.EndBoundaryAt != rto.BoundaryEvent.At ||
			review.StateID != rto.BoundaryEvent.StateID || review.BuildID != rto.BoundaryEvent.BuildID || review.PathID != rto.BoundaryEvent.PathID {
			violations = append(violations, "RTO end boundary is not the earliest supported capability-usable boundary")
		}
		if !containsAll(review.EvidenceIDs, rto.RequiredEvidenceIDs) {
			violations = append(violations, "RTO review omits evidence required for the selected boundary")
		}
		if !sameRejectedEvents(review.RejectedEarlierEvents, rto.RejectedEarlier) {
			violations = append(violations, "earlier recovery signals are not rejected with supported reasons")
		}
		if review.FirstCustomerResultEventID != rto.FirstCustomerResult.ID ||
			review.FirstCustomerResultRelation != rto.FirstCustomerRelation {
			violations = append(violations, "first customer result is related to the recovery boundary incorrectly")
		}
		if review.RTOVerdict != verdict(rto.Pass) {
			violations = append(violations, "RTO verdict is incorrect")
		}
	}

	for name, value := range map[string]string{
		"all capabilities recovered":            decision.Claims.AllCapabilitiesRecovered,
		"future disruptions meet objectives":    decision.Claims.FutureDisruptionsMeetObjectives,
		"RPO selects a recovery mechanism":      decision.Claims.RPOSelectsRecoveryMechanism,
		"first customer result defines RTO end": decision.Claims.FirstCustomerResultDefinesRTOEnd,
	} {
		if value != ClaimUnsupported {
			violations = append(violations, name+" is a stronger unsupported claim")
		}
	}
	return violations
}

func validateContract(contract Contract) error {
	if contract.Capability.ID == "" || contract.Capability.RequiredResultMeaning == "" || len(contract.Capability.EligiblePathIDs) == 0 {
		return fmt.Errorf("capability identity, result meaning, and eligible paths are required")
	}
	if _, err := time.Parse(time.RFC3339, contract.Disruption.At); err != nil {
		return fmt.Errorf("disruption time: %w", err)
	}
	if contract.Disruption.RPOSeconds <= 0 || contract.Disruption.RTOSeconds <= 0 {
		return fmt.Errorf("positive RPO and RTO bounds are required")
	}
	if !contract.AcceptedHistoryAuthority.CompleteThroughDisruption || contract.AcceptedHistoryAuthority.OrderField != "position" {
		return fmt.Errorf("accepted-history authority must declare complete ordered coverage")
	}
	if !contract.RecoveryEvidenceAuthority.CompleteMembershipByState {
		return fmt.Errorf("recovery evidence must declare complete state membership")
	}
	if contract.ServingEvidenceAuthority.ID == "" || contract.ServingEvidenceAuthority.CapabilityUsableEventKind == "" ||
		!contract.ServingEvidenceAuthority.RequiresAdmissibleStateValidInterpretationAndEligiblePath {
		return fmt.Errorf("serving evidence authority must declare the capability-usable event contract")
	}
	if contract.ProtectedAcknowledgement.OperationClass == "" || contract.ProtectedAcknowledgement.Result == "" {
		return fmt.Errorf("protected acknowledgement scope is required")
	}
	seen := map[string]bool{}
	for _, interpretation := range contract.Capability.BuildInterpretations {
		key := fmt.Sprintf("%s/%d", interpretation.BuildID, interpretation.StatusCode)
		if interpretation.BuildID == "" || interpretation.Meaning == "" || seen[key] {
			return fmt.Errorf("build interpretations must be complete and unique")
		}
		seen[key] = true
	}
	return nil
}

func validateAcceptedHistory(facts []AcceptedFact) error {
	if len(facts) == 0 {
		return fmt.Errorf("accepted history is empty")
	}
	ids := map[string]bool{}
	operations := map[string]bool{}
	lastPosition := 0
	var lastTime time.Time
	for _, fact := range facts {
		acceptedAt, err := time.Parse(time.RFC3339, fact.AcceptedAt)
		if err != nil {
			return fmt.Errorf("fact %s accepted_at: %w", fact.ID, err)
		}
		if fact.ID == "" || fact.OperationID == "" || fact.OperationClass == "" || fact.AcknowledgementResult == "" {
			return fmt.Errorf("accepted fact fields are required")
		}
		if ids[fact.ID] || operations[fact.OperationID] {
			return fmt.Errorf("accepted facts and operations must be unique")
		}
		if fact.Position <= lastPosition {
			return fmt.Errorf("accepted-history positions must increase")
		}
		if !lastTime.IsZero() && acceptedAt.Before(lastTime) {
			return fmt.Errorf("accepted-history times must not regress")
		}
		ids[fact.ID], operations[fact.OperationID] = true, true
		lastPosition, lastTime = fact.Position, acceptedAt
	}
	return nil
}

func validateRecoveryEvidence(states []RecoveredState, facts []AcceptedFact) error {
	if len(states) == 0 {
		return fmt.Errorf("recovery evidence is empty")
	}
	knownFacts := map[string]bool{}
	factByID := map[string]AcceptedFact{}
	for _, fact := range facts {
		knownFacts[fact.ID] = true
		factByID[fact.ID] = fact
	}
	stateIDs := map[string]bool{}
	evidenceIDs := map[string]bool{}
	for _, state := range states {
		if state.StateID == "" || state.EvidenceID == "" || len(state.AccountedFactIDs) == 0 || state.StatusCodeFactID == "" {
			return fmt.Errorf("recovered state fields are required")
		}
		if stateIDs[state.StateID] || evidenceIDs[state.EvidenceID] {
			return fmt.Errorf("state and evidence identifiers must be unique")
		}
		stateIDs[state.StateID], evidenceIDs[state.EvidenceID] = true, true
		seenFacts := map[string]bool{}
		for _, factID := range state.AccountedFactIDs {
			if !knownFacts[factID] {
				return fmt.Errorf("state %s cites unknown fact %s", state.StateID, factID)
			}
			if seenFacts[factID] {
				return fmt.Errorf("state %s repeats fact %s", state.StateID, factID)
			}
			seenFacts[factID] = true
		}
		if !seenFacts[state.StatusCodeFactID] {
			return fmt.Errorf("state %s status source is not accounted for", state.StateID)
		}
		statusFact := factByID[state.StatusCodeFactID]
		if statusFact.StatusCode == nil || *statusFact.StatusCode != state.StatusCode {
			return fmt.Errorf("state %s status code is not supported by fact %s", state.StateID, state.StatusCodeFactID)
		}
	}
	return nil
}

func validateServingEvents(events []ServingEvent, states []RecoveredState) error {
	if len(events) == 0 {
		return fmt.Errorf("serving events are empty")
	}
	knownStates := map[string]bool{}
	for _, state := range states {
		knownStates[state.StateID] = true
	}
	ids := map[string]bool{}
	var lastTime time.Time
	for _, event := range events {
		at, err := time.Parse(time.RFC3339, event.At)
		if err != nil {
			return fmt.Errorf("event %s at: %w", event.ID, err)
		}
		if event.ID == "" || event.Kind == "" || event.StateID == "" || event.BuildID == "" || event.PathID == "" {
			return fmt.Errorf("serving event fields are required")
		}
		if ids[event.ID] {
			return fmt.Errorf("serving event %s is duplicated", event.ID)
		}
		if !knownStates[event.StateID] {
			return fmt.Errorf("event %s cites unknown state %s", event.ID, event.StateID)
		}
		if !lastTime.IsZero() && at.Before(lastTime) {
			return fmt.Errorf("serving event times must not regress")
		}
		if event.Kind == EventRequest && (event.ResultMeaning == "" || event.Audience == "") {
			return fmt.Errorf("request event %s lacks result meaning or audience", event.ID)
		}
		if event.Kind != EventRequest && (event.ResultMeaning != "" || event.Audience != "") {
			return fmt.Errorf("non-request event %s carries request-only fields", event.ID)
		}
		ids[event.ID], lastTime = true, at
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
	result := []T{}
	scanner := bufio.NewScanner(file)
	line := 0
	for scanner.Scan() {
		line++
		if strings.TrimSpace(scanner.Text()) == "" {
			continue
		}
		var value T
		decoder := json.NewDecoder(strings.NewReader(scanner.Text()))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&value); err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}
		var trailing any
		if err := decoder.Decode(&trailing); err != io.EOF {
			if err == nil {
				return nil, fmt.Errorf("line %d: unexpected trailing JSON value", line)
			}
			return nil, fmt.Errorf("line %d trailing JSON: %w", line, err)
		}
		result = append(result, value)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func interpretation(contract Contract, buildID string, statusCode int) string {
	for _, candidate := range contract.Capability.BuildInterpretations {
		if candidate.BuildID == buildID && candidate.StatusCode == statusCode {
			return candidate.Meaning
		}
	}
	return ""
}

func verdict(value bool) string {
	if value {
		return VerdictPass
	}
	return VerdictFail
}
func mustTime(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}
	return parsed
}
func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
func stringSet(values []string) map[string]bool {
	result := map[string]bool{}
	for _, value := range values {
		result[value] = true
	}
	return result
}
func sameStringSet(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	a, b := append([]string{}, left...), append([]string{}, right...)
	sort.Strings(a)
	sort.Strings(b)
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func containsAll(actual, required []string) bool {
	set := stringSet(actual)
	for _, value := range required {
		if !set[value] {
			return false
		}
	}
	return true
}
func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, value := range values {
		if value != "" && !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}

func sameRejectedEvents(actual, expected []RejectedEvent) bool {
	if len(actual) != len(expected) {
		return false
	}
	actualByID := map[string][]string{}
	for _, item := range actual {
		if _, exists := actualByID[item.EventID]; exists {
			return false
		}
		actualByID[item.EventID] = item.Reasons
	}
	for _, item := range expected {
		reasons, exists := actualByID[item.EventID]
		if !exists || !sameStringSet(reasons, item.Reasons) {
			return false
		}
	}
	return true
}
