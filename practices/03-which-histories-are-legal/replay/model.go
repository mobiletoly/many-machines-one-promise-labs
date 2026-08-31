package replay

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
)

const (
	KindReadYourWrites       = "read_your_writes"
	KindMonotonicReads       = "monotonic_reads"
	KindDependentObservation = "dependent_observation"

	VerdictSatisfies = "satisfies"
	VerdictViolates  = "violates"
)

type Contract struct {
	VisibilityContracts []VisibilityContract `json:"visibility_contracts"`
	ObjectContracts     []ObjectContract     `json:"object_contracts"`
}

type VisibilityContract struct {
	ID                    string `json:"id"`
	Kind                  string `json:"kind"`
	Scope                 string `json:"scope"`
	SessionID             string `json:"session_id,omitempty"`
	TargetReadID          string `json:"target_read_id"`
	ExposedOperationID    string `json:"exposed_operation_id,omitempty"`
	AcceptancePredicateID string `json:"acceptance_predicate_id,omitempty"`
	RequiredInputRole     string `json:"required_input_role,omitempty"`
}

type ObjectContract struct {
	ID           string `json:"id"`
	ObjectID     string `json:"object_id"`
	InitialValue string `json:"initial_value"`
	SetOperation string `json:"set_operation"`
	GetOperation string `json:"get_operation"`
	SetResult    string `json:"set_result"`
}

type Corpus struct {
	VisibilityHistories []VisibilityHistory `json:"visibility_histories"`
	ObjectHistories     []ObjectHistory     `json:"object_histories"`
}

type VisibilityHistory struct {
	ID         string            `json:"id"`
	ContractID string            `json:"contract_id"`
	Events     []VisibilityEvent `json:"events"`
}

type VisibilityEvent struct {
	ID                 string   `json:"id"`
	Order              int      `json:"order"`
	Kind               string   `json:"kind"`
	SessionID          string   `json:"session_id,omitempty"`
	OperationID        string   `json:"operation_id,omitempty"`
	FactID             string   `json:"fact_id,omitempty"`
	FactType           string   `json:"fact_type,omitempty"`
	Field              string   `json:"field,omitempty"`
	Value              string   `json:"value,omitempty"`
	Version            int      `json:"version,omitempty"`
	Result             string   `json:"result,omitempty"`
	StateFactIDs       []string `json:"state_fact_ids,omitempty"`
	ResultFactIDs      []string `json:"result_fact_ids,omitempty"`
	WithdrawsFactID    string   `json:"withdraws_fact_id,omitempty"`
	SourceReadID       string   `json:"source_read_id,omitempty"`
	SelectedFactID     string   `json:"selected_fact_id,omitempty"`
	SelectedVersion    int      `json:"selected_version,omitempty"`
	PredicateID        string   `json:"predicate_id,omitempty"`
	PredicateInputRole string   `json:"predicate_input_role,omitempty"`
	PredicateFactID    string   `json:"predicate_fact_id,omitempty"`
	PredicateVersion   int      `json:"predicate_version,omitempty"`
	PredicatePassed    bool     `json:"predicate_passed,omitempty"`
	ReplicaID          string   `json:"replica_id,omitempty"`
	ExposesOperationID string   `json:"exposes_operation_id,omitempty"`
}

type ObjectHistory struct {
	ID         string        `json:"id"`
	ContractID string        `json:"contract_id"`
	Events     []ObjectEvent `json:"events"`
}

type ObjectEvent struct {
	ID            string `json:"id"`
	Order         int    `json:"order"`
	Kind          string `json:"kind"`
	OperationID   string `json:"operation_id"`
	OperationKind string `json:"operation_kind,omitempty"`
	Argument      string `json:"argument,omitempty"`
	Result        string `json:"result,omitempty"`
}

type Review struct {
	Name              string             `json:"name"`
	VisibilityReviews []VisibilityReview `json:"visibility_reviews"`
	ObjectReviews     []ObjectReview     `json:"object_reviews"`
	Claims            *Claims            `json:"claims"`
}

type VisibilityReview struct {
	HistoryID                    string                `json:"history_id"`
	Scope                        string                `json:"scope"`
	RequiredFacts                []FactRequirement     `json:"required_facts"`
	StateEvidenceEventID         string                `json:"state_evidence_event_id"`
	LaterExplainingChanges       []ExplainingChange    `json:"later_explaining_changes"`
	InformationPaths             []InformationPath     `json:"information_paths"`
	ChronologicalNonDependencies []ChronologicalNonDep `json:"chronological_non_dependencies"`
	Verdict                      string                `json:"verdict"`
	MissingFactIDs               []string              `json:"missing_fact_ids"`
}

type FactRequirement struct {
	FactID      string   `json:"fact_id"`
	Basis       string   `json:"basis"`
	EvidenceIDs []string `json:"evidence_ids"`
}

type ExplainingChange struct {
	FactID        string   `json:"fact_id"`
	AffectsFactID string   `json:"affects_fact_id"`
	EvidenceIDs   []string `json:"evidence_ids"`
}

type InformationPath struct {
	FromFactID    string   `json:"from_fact_id"`
	ToOperationID string   `json:"to_operation_id"`
	PredicateID   string   `json:"predicate_id"`
	InputRole     string   `json:"input_role"`
	EvidenceIDs   []string `json:"evidence_ids"`
}

type ChronologicalNonDep struct {
	FactID      string   `json:"fact_id"`
	Reason      string   `json:"reason"`
	EvidenceIDs []string `json:"evidence_ids"`
}

type ObjectReview struct {
	HistoryID     string        `json:"history_id"`
	ForcedEdges   []ForcedEdge  `json:"forced_edges"`
	WitnessOrder  []string      `json:"witness_order"`
	Contradiction Contradiction `json:"contradiction"`
	Verdict       string        `json:"verdict"`
}

type ForcedEdge struct {
	Before      string   `json:"before"`
	After       string   `json:"after"`
	EvidenceIDs []string `json:"evidence_ids"`
}

type Contradiction struct {
	Kind         string   `json:"kind"`
	OperationIDs []string `json:"operation_ids"`
	EvidenceIDs  []string `json:"evidence_ids"`
}

type Claims struct {
	SessionVisibilityEstablishesObjectLinearizability *bool `json:"session_visibility_establishes_object_linearizability"`
	ObjectHistoryEstablishesSessionVisibility         *bool `json:"object_history_establishes_session_visibility"`
	OneDependencyEstablishesGeneralCausalConsistency  *bool `json:"one_dependency_establishes_general_causal_consistency"`
	OneLegalHistoryGuaranteesFutureAvailability       *bool `json:"one_legal_history_guarantees_future_availability"`
}

type Evaluation struct {
	Review              Review
	VisibilitySatisfies int
	VisibilityViolates  int
	ObjectSatisfies     int
	ObjectViolates      int
	Violations          []string
}

type factTruth struct {
	FactID           string
	Basis            string
	RequiredEvidence []string
	AllowedEvidence  []string
}

type changeTruth struct {
	FactID           string
	AffectsFactID    string
	RequiredEvidence []string
	AllowedEvidence  []string
}

type visibilityTruth struct {
	Scope           string
	RequiredFacts   []factTruth
	StateEventID    string
	Changes         []changeTruth
	Paths           []InformationPath
	NonDependencies []ChronologicalNonDep
	MissingFacts    []string
	Verdict         string
}

type completedOperation struct {
	ID             string
	Kind           string
	Argument       string
	Result         string
	InvokeOrder    int
	RespondOrder   int
	InvokeEventID  string
	RespondEventID string
}

type objectTruth struct {
	ForcedEdges    []ForcedEdge
	LegalWitnesses [][]string
	Contradiction  Contradiction
	Verdict        string
}

func LoadContract(path string) (Contract, error) {
	var contract Contract
	if err := loadStrictJSON(path, &contract); err != nil {
		return Contract{}, fmt.Errorf("load contract: %w", err)
	}
	if err := validateContract(contract); err != nil {
		return Contract{}, err
	}
	return contract, nil
}

func LoadCorpus(path string) (Corpus, error) {
	var corpus Corpus
	if err := loadStrictJSON(path, &corpus); err != nil {
		return Corpus{}, fmt.Errorf("load histories: %w", err)
	}
	return corpus, nil
}

func LoadReview(path string) (Review, error) {
	var review Review
	if err := loadStrictJSON(path, &review); err != nil {
		return Review{}, fmt.Errorf("load review: %w", err)
	}
	if review.Name == "" {
		return Review{}, fmt.Errorf("review name is empty")
	}
	if err := validateClaimsPresence(review.Claims); err != nil {
		return Review{}, err
	}
	return review, nil
}

func loadStrictJSON(path string, target any) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
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

func Evaluate(review Review, contract Contract, corpus Corpus) Evaluation {
	evaluation := Evaluation{Review: review}
	contractVisibility := map[string]VisibilityContract{}
	for _, candidate := range contract.VisibilityContracts {
		contractVisibility[candidate.ID] = candidate
	}
	contractObjects := map[string]ObjectContract{}
	for _, candidate := range contract.ObjectContracts {
		contractObjects[candidate.ID] = candidate
	}

	visibilityReviews, violations := indexVisibilityReviews(review.VisibilityReviews)
	evaluation.Violations = append(evaluation.Violations, violations...)
	objectReviews, violations := indexObjectReviews(review.ObjectReviews)
	evaluation.Violations = append(evaluation.Violations, violations...)

	seenHistories := map[string]bool{}
	for _, history := range corpus.VisibilityHistories {
		if seenHistories[history.ID] {
			evaluation.Violations = append(evaluation.Violations, fmt.Sprintf("duplicate history id %s", history.ID))
			continue
		}
		seenHistories[history.ID] = true
		candidateContract, ok := contractVisibility[history.ContractID]
		if !ok {
			evaluation.Violations = append(evaluation.Violations, fmt.Sprintf("%s references unknown visibility contract %s", history.ID, history.ContractID))
			continue
		}
		truth, historyViolations := deriveVisibility(history, candidateContract)
		evaluation.Violations = append(evaluation.Violations, prefixViolations(history.ID, historyViolations)...)
		if truth.Verdict == VerdictSatisfies {
			evaluation.VisibilitySatisfies++
		} else {
			evaluation.VisibilityViolates++
		}
		submitted, ok := visibilityReviews[history.ID]
		if !ok {
			evaluation.Violations = append(evaluation.Violations, fmt.Sprintf("missing visibility review for %s", history.ID))
			continue
		}
		evaluation.Violations = append(evaluation.Violations, validateVisibilityReview(history, submitted, truth)...)
		delete(visibilityReviews, history.ID)
	}
	for id := range visibilityReviews {
		evaluation.Violations = append(evaluation.Violations, fmt.Sprintf("visibility review references unknown history %s", id))
	}

	for _, history := range corpus.ObjectHistories {
		if seenHistories[history.ID] {
			evaluation.Violations = append(evaluation.Violations, fmt.Sprintf("duplicate history id %s", history.ID))
			continue
		}
		seenHistories[history.ID] = true
		candidateContract, ok := contractObjects[history.ContractID]
		if !ok {
			evaluation.Violations = append(evaluation.Violations, fmt.Sprintf("%s references unknown object contract %s", history.ID, history.ContractID))
			continue
		}
		truth, historyViolations := deriveObject(history, candidateContract)
		evaluation.Violations = append(evaluation.Violations, prefixViolations(history.ID, historyViolations)...)
		if truth.Verdict == VerdictSatisfies {
			evaluation.ObjectSatisfies++
		} else {
			evaluation.ObjectViolates++
		}
		submitted, ok := objectReviews[history.ID]
		if !ok {
			evaluation.Violations = append(evaluation.Violations, fmt.Sprintf("missing object review for %s", history.ID))
			continue
		}
		evaluation.Violations = append(evaluation.Violations, validateObjectReview(history, submitted, truth)...)
		delete(objectReviews, history.ID)
	}
	for id := range objectReviews {
		evaluation.Violations = append(evaluation.Violations, fmt.Sprintf("object review references unknown history %s", id))
	}

	evaluation.Violations = append(evaluation.Violations, validateClaims(review.Claims)...)
	if len(evaluation.Violations) != 0 {
		evaluation.Violations = append([]string{"review does not prove the declared histories"}, evaluation.Violations...)
	}
	return evaluation
}

func validateContract(contract Contract) error {
	seen := map[string]bool{}
	for _, candidate := range contract.VisibilityContracts {
		if candidate.ID == "" || seen[candidate.ID] {
			return fmt.Errorf("contract has an empty or duplicate id %q", candidate.ID)
		}
		seen[candidate.ID] = true
		if candidate.Scope == "" || candidate.TargetReadID == "" {
			return fmt.Errorf("visibility contract %s lacks scope or target read", candidate.ID)
		}
		switch candidate.Kind {
		case KindReadYourWrites, KindMonotonicReads:
			if candidate.SessionID == "" {
				return fmt.Errorf("visibility contract %s lacks a session", candidate.ID)
			}
		case KindDependentObservation:
			if candidate.ExposedOperationID == "" || candidate.AcceptancePredicateID == "" || candidate.RequiredInputRole == "" {
				return fmt.Errorf("visibility contract %s lacks dependency semantics", candidate.ID)
			}
		default:
			return fmt.Errorf("visibility contract %s has unsupported kind %q", candidate.ID, candidate.Kind)
		}
	}
	for _, candidate := range contract.ObjectContracts {
		if candidate.ID == "" || seen[candidate.ID] {
			return fmt.Errorf("contract has an empty or duplicate id %q", candidate.ID)
		}
		seen[candidate.ID] = true
		if candidate.ObjectID == "" || candidate.SetOperation == "" || candidate.GetOperation == "" || candidate.SetResult == "" {
			return fmt.Errorf("object contract %s is incomplete", candidate.ID)
		}
	}
	if len(contract.VisibilityContracts) == 0 || len(contract.ObjectContracts) == 0 {
		return fmt.Errorf("contract must declare both proof families")
	}
	return nil
}

func deriveVisibility(history VisibilityHistory, contract VisibilityContract) (visibilityTruth, []string) {
	truth := visibilityTruth{Scope: contract.Scope}
	_, violations := indexVisibilityEvents(history)
	target, ok := findVisibilityEventByOperation(history.Events, contract.TargetReadID, "read_succeeded")
	if !ok {
		return truth, append(violations, fmt.Sprintf("target read %s is missing", contract.TargetReadID))
	}
	truth.StateEventID = target.ID

	switch contract.Kind {
	case KindReadYourWrites:
		for _, event := range history.Events {
			if event.Kind != "write_accepted" || event.SessionID != contract.SessionID || event.Order >= target.Order {
				continue
			}
			allowed := []string{event.ID, target.ID}
			if response, found := findVisibilityEventByOperation(history.Events, event.OperationID, "write_responded"); found {
				allowed = append(allowed, response.ID)
			}
			truth.RequiredFacts = append(truth.RequiredFacts, factTruth{
				FactID: event.FactID, Basis: "accepted_in_same_session",
				RequiredEvidence: []string{event.ID, target.ID}, AllowedEvidence: allowed,
			})
		}
	case KindMonotonicReads:
		for _, earlier := range history.Events {
			if earlier.Kind != "read_succeeded" || earlier.SessionID != contract.SessionID || earlier.Order >= target.Order {
				continue
			}
			for _, factID := range earlier.ResultFactIDs {
				if containsFactTruth(truth.RequiredFacts, factID) {
					continue
				}
				truth.RequiredFacts = append(truth.RequiredFacts, factTruth{
					FactID: factID, Basis: "earlier_successful_observation",
					RequiredEvidence: []string{earlier.ID, target.ID}, AllowedEvidence: []string{earlier.ID, target.ID},
				})
			}
		}
		for _, event := range history.Events {
			if event.Kind != "fact_accepted" || event.WithdrawsFactID == "" || event.Order >= target.Order {
				continue
			}
			if containsFactTruth(truth.RequiredFacts, event.WithdrawsFactID) && contains(target.StateFactIDs, event.FactID) {
				truth.Changes = append(truth.Changes, changeTruth{
					FactID: event.FactID, AffectsFactID: event.WithdrawsFactID,
					RequiredEvidence: []string{event.ID, target.ID}, AllowedEvidence: []string{event.ID, target.ID},
				})
			}
		}
		violations = append(violations, validateActiveNoticeResult(history, target)...)
	case KindDependentObservation:
		dependency, dependencyViolations := deriveDependency(history, contract, target)
		violations = append(violations, dependencyViolations...)
		if dependency != nil {
			truth.RequiredFacts = append(truth.RequiredFacts, dependency.fact)
			truth.Paths = append(truth.Paths, dependency.path)
			truth.NonDependencies = append(truth.NonDependencies, dependency.nonDependencies...)
		}
	}

	for _, required := range truth.RequiredFacts {
		if !contains(target.StateFactIDs, required.FactID) {
			truth.MissingFacts = append(truth.MissingFacts, required.FactID)
		}
	}
	sort.Strings(truth.MissingFacts)
	if len(truth.MissingFacts) == 0 {
		truth.Verdict = VerdictSatisfies
	} else {
		truth.Verdict = VerdictViolates
	}
	return truth, violations
}

type dependencyTruth struct {
	fact            factTruth
	path            InformationPath
	nonDependencies []ChronologicalNonDep
}

func deriveDependency(history VisibilityHistory, contract VisibilityContract, target VisibilityEvent) (*dependencyTruth, []string) {
	var violations []string
	if target.ExposesOperationID != contract.ExposedOperationID {
		violations = append(violations, "target result exposes the wrong operation")
		return nil, violations
	}
	submission, ok := findVisibilityEventByOperation(history.Events, contract.ExposedOperationID, "operation_submitted")
	if !ok {
		return nil, append(violations, "dependent operation submission is missing")
	}
	predicate, ok := findVisibilityEventByOperation(history.Events, contract.ExposedOperationID, "predicate_evaluated")
	if !ok {
		return nil, append(violations, "dependent operation predicate evidence is missing")
	}
	accepted, ok := findVisibilityEventByOperation(history.Events, contract.ExposedOperationID, "operation_accepted")
	if !ok {
		return nil, append(violations, "dependent operation acceptance is missing")
	}
	sourceRead, ok := findVisibilityEventByOperation(history.Events, submission.SourceReadID, "read_succeeded")
	if !ok {
		return nil, append(violations, "source observation is missing")
	}
	factEvent, ok := findVisibilityEventByFact(history.Events, submission.SelectedFactID, "fact_accepted")
	if !ok {
		return nil, append(violations, "selected accepted fact is missing")
	}

	pathIDs := []string{factEvent.ID, sourceRead.ID, submission.ID, predicate.ID, accepted.ID}
	pathExists := contains(sourceRead.ResultFactIDs, submission.SelectedFactID) && sourceRead.Version == submission.SelectedVersion &&
		factEvent.Version == submission.SelectedVersion && factEvent.Order < sourceRead.Order && sourceRead.Order < submission.Order &&
		submission.Order < predicate.Order && predicate.Order < accepted.Order && accepted.Order < target.Order
	if !pathExists {
		violations = append(violations, "connected records do not establish the claimed X-to-Y information path")
	}

	relevant := predicate.PredicateID == contract.AcceptancePredicateID && predicate.PredicateInputRole == contract.RequiredInputRole &&
		predicate.PredicateFactID == submission.SelectedFactID && predicate.PredicateVersion == submission.SelectedVersion &&
		predicate.PredicatePassed
	if !relevant {
		violations = append(violations, "connected path does not show that the X-derived fact participates in Y's declared acceptance rule")
	}
	if !pathExists || !relevant {
		return nil, violations
	}

	truth := &dependencyTruth{
		fact: factTruth{
			FactID: submission.SelectedFactID, Basis: "consumed_dependency_input",
			RequiredEvidence: append(append([]string{}, pathIDs...), target.ID),
			AllowedEvidence:  append(append([]string{}, pathIDs...), target.ID),
		},
		path: InformationPath{
			FromFactID: submission.SelectedFactID, ToOperationID: submission.OperationID,
			PredicateID: predicate.PredicateID, InputRole: predicate.PredicateInputRole, EvidenceIDs: pathIDs,
		},
	}
	for _, event := range history.Events {
		if event.Kind != "fact_accepted" || event.Order >= submission.Order || event.FactID == submission.SelectedFactID {
			continue
		}
		truth.nonDependencies = append(truth.nonDependencies, ChronologicalNonDep{
			FactID: event.FactID, Reason: "not_consumed_by_declared_rule", EvidenceIDs: []string{event.ID, predicate.ID},
		})
	}
	return truth, violations
}

func validateActiveNoticeResult(history VisibilityHistory, target VisibilityEvent) []string {
	facts := map[string]VisibilityEvent{}
	for _, event := range history.Events {
		if event.FactID != "" {
			facts[event.FactID] = event
		}
	}
	active := map[string]bool{}
	for _, factID := range target.StateFactIDs {
		fact, ok := facts[factID]
		if !ok {
			return []string{fmt.Sprintf("target state references unknown fact %s", factID)}
		}
		if fact.FactType == "notice" {
			active[factID] = true
		}
	}
	for _, factID := range target.StateFactIDs {
		fact := facts[factID]
		if fact.FactType == "withdrawal" && fact.WithdrawsFactID != "" {
			delete(active, fact.WithdrawsFactID)
		}
	}
	want := keys(active)
	if !sameStringSet(want, target.ResultFactIDs) {
		return []string{"target result does not match the declared active-notice view of its serving state"}
	}
	return nil
}

func deriveObject(history ObjectHistory, contract ObjectContract) (objectTruth, []string) {
	operations, events, violations := completeOperations(history, contract)
	truth := objectTruth{}
	if len(violations) != 0 {
		return truth, violations
	}
	for _, left := range operations {
		for _, right := range operations {
			if left.ID == right.ID || left.RespondOrder >= right.InvokeOrder {
				continue
			}
			truth.ForcedEdges = append(truth.ForcedEdges, ForcedEdge{
				Before: left.ID, After: right.ID, EvidenceIDs: []string{left.RespondEventID, right.InvokeEventID},
			})
		}
	}
	sort.Slice(truth.ForcedEdges, func(i, j int) bool {
		return edgeKey(truth.ForcedEdges[i]) < edgeKey(truth.ForcedEdges[j])
	})
	for _, order := range permutations(operationIDs(operations)) {
		if respectsEdges(order, truth.ForcedEdges) && legalSequentialOrder(order, operations, contract) {
			truth.LegalWitnesses = append(truth.LegalWitnesses, order)
		}
	}
	if len(truth.LegalWitnesses) != 0 {
		truth.Verdict = VerdictSatisfies
		truth.Contradiction = Contradiction{Kind: "none", OperationIDs: []string{}, EvidenceIDs: []string{}}
		return truth, nil
	}
	truth.Verdict = VerdictViolates
	if unexplained, ok := unexplainedRead(operations, contract); ok {
		truth.Contradiction = Contradiction{
			Kind: "unexplained_result", OperationIDs: []string{unexplained.ID}, EvidenceIDs: []string{unexplained.RespondEventID},
		}
		return truth, nil
	}
	if len(truth.ForcedEdges) != 0 {
		edge := truth.ForcedEdges[0]
		responseID := ""
		if operation, ok := operationByID(operations, edge.After); ok {
			responseID = operation.RespondEventID
		}
		evidence := append([]string{}, edge.EvidenceIDs...)
		if responseID != "" {
			evidence = append(evidence, responseID)
		}
		truth.Contradiction = Contradiction{
			Kind: "forced_order_result_conflict", OperationIDs: []string{edge.Before, edge.After}, EvidenceIDs: evidence,
		}
		return truth, nil
	}
	truth.Contradiction = Contradiction{Kind: "no_sequential_witness", OperationIDs: operationIDs(operations), EvidenceIDs: keysObjectEvents(events)}
	return truth, nil
}

func completeOperations(history ObjectHistory, contract ObjectContract) ([]completedOperation, map[string]ObjectEvent, []string) {
	var violations []string
	events := map[string]ObjectEvent{}
	operations := map[string]*completedOperation{}
	for index, event := range history.Events {
		if event.ID == "" || event.Order != index+1 {
			violations = append(violations, fmt.Sprintf("event %d lacks a stable id or order", index+1))
		}
		if _, exists := events[event.ID]; exists {
			violations = append(violations, fmt.Sprintf("duplicate event id %s", event.ID))
		}
		events[event.ID] = event
		if event.OperationID == "" {
			violations = append(violations, fmt.Sprintf("event %s lacks an operation id", event.ID))
			continue
		}
		operation := operations[event.OperationID]
		if operation == nil {
			operation = &completedOperation{ID: event.OperationID}
			operations[event.OperationID] = operation
		}
		switch event.Kind {
		case "invoke":
			if operation.InvokeOrder != 0 {
				violations = append(violations, fmt.Sprintf("operation %s has duplicate invocation", event.OperationID))
			}
			operation.Kind, operation.Argument = event.OperationKind, event.Argument
			operation.InvokeOrder, operation.InvokeEventID = event.Order, event.ID
		case "respond":
			if operation.RespondOrder != 0 {
				violations = append(violations, fmt.Sprintf("operation %s has duplicate response", event.OperationID))
			}
			operation.Result = event.Result
			operation.RespondOrder, operation.RespondEventID = event.Order, event.ID
		default:
			violations = append(violations, fmt.Sprintf("event %s has unsupported kind %q", event.ID, event.Kind))
		}
	}
	completed := make([]completedOperation, 0, len(operations))
	for _, operation := range operations {
		if operation.InvokeOrder == 0 || operation.RespondOrder == 0 || operation.InvokeOrder >= operation.RespondOrder {
			violations = append(violations, fmt.Sprintf("operation %s is not one completed invocation", operation.ID))
			continue
		}
		if operation.Kind != contract.SetOperation && operation.Kind != contract.GetOperation {
			violations = append(violations, fmt.Sprintf("operation %s has unsupported operation kind %q", operation.ID, operation.Kind))
		}
		completed = append(completed, *operation)
	}
	sort.Slice(completed, func(i, j int) bool { return completed[i].ID < completed[j].ID })
	return completed, events, violations
}

func legalSequentialOrder(order []string, operations []completedOperation, contract ObjectContract) bool {
	state := contract.InitialValue
	for _, id := range order {
		operation, ok := operationByID(operations, id)
		if !ok {
			return false
		}
		switch operation.Kind {
		case contract.SetOperation:
			if operation.Result != contract.SetResult {
				return false
			}
			state = operation.Argument
		case contract.GetOperation:
			if operation.Result != state {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func unexplainedRead(operations []completedOperation, contract ObjectContract) (completedOperation, bool) {
	possible := map[string]bool{contract.InitialValue: true}
	for _, operation := range operations {
		if operation.Kind == contract.SetOperation {
			possible[operation.Argument] = true
		}
	}
	for _, operation := range operations {
		if operation.Kind == contract.GetOperation && !possible[operation.Result] {
			return operation, true
		}
	}
	return completedOperation{}, false
}

func validateVisibilityReview(history VisibilityHistory, submitted VisibilityReview, truth visibilityTruth) []string {
	var violations []string
	prefix := history.ID + ": "
	if submitted.Scope != truth.Scope {
		violations = append(violations, prefix+"scope does not match the declared successful-observation boundary")
	}
	if submitted.StateEvidenceEventID != truth.StateEventID {
		violations = append(violations, prefix+"state evidence does not identify the target successful read")
	}
	events := map[string]bool{}
	for _, event := range history.Events {
		events[event.ID] = true
	}
	actualFacts, duplicates := indexFactRequirements(submitted.RequiredFacts)
	if duplicates {
		violations = append(violations, prefix+"required facts contain duplicates")
	}
	if len(actualFacts) != len(truth.RequiredFacts) {
		violations = append(violations, prefix+"required fact set is incomplete or contains unsupported facts")
	}
	for _, expected := range truth.RequiredFacts {
		actual, ok := actualFacts[expected.FactID]
		if !ok {
			violations = append(violations, prefix+fmt.Sprintf("required fact %s is missing", expected.FactID))
			continue
		}
		if actual.Basis != expected.Basis {
			violations = append(violations, prefix+fmt.Sprintf("required fact %s uses unsupported basis %q", expected.FactID, actual.Basis))
		}
		violations = append(violations, validateEvidence(prefix+"required fact "+expected.FactID, actual.EvidenceIDs, expected.RequiredEvidence, expected.AllowedEvidence, events)...)
	}
	violations = append(violations, compareChanges(prefix, submitted.LaterExplainingChanges, truth.Changes, events)...)
	violations = append(violations, comparePaths(prefix, submitted.InformationPaths, truth.Paths, events)...)
	violations = append(violations, compareNonDependencies(prefix, submitted.ChronologicalNonDependencies, truth.NonDependencies, events)...)
	if submitted.Verdict != truth.Verdict {
		violations = append(violations, prefix+"verdict does not follow from required fact coverage")
	}
	if !sameStringSet(submitted.MissingFactIDs, truth.MissingFacts) {
		violations = append(violations, prefix+"missing fact contradiction does not match the serving evidence")
	}
	return violations
}

func validateObjectReview(history ObjectHistory, submitted ObjectReview, truth objectTruth) []string {
	var violations []string
	prefix := history.ID + ": "
	events := map[string]bool{}
	for _, event := range history.Events {
		events[event.ID] = true
	}
	if !sameEdges(submitted.ForcedEdges, truth.ForcedEdges) {
		violations = append(violations, prefix+"forced edges do not follow from response-before-invocation evidence")
	}
	for _, edge := range submitted.ForcedEdges {
		expected, ok := edgeByKey(truth.ForcedEdges, edgeKey(edge))
		if !ok {
			continue
		}
		violations = append(violations, validateEvidence(prefix+"forced edge "+edge.Before+"->"+edge.After,
			edge.EvidenceIDs, expected.EvidenceIDs, expected.EvidenceIDs, events)...)
	}
	if submitted.Verdict != truth.Verdict {
		violations = append(violations, prefix+"verdict does not match the legal witness space")
	}
	if truth.Verdict == VerdictSatisfies {
		if !containsWitness(truth.LegalWitnesses, submitted.WitnessOrder) {
			violations = append(violations, prefix+"submitted order is not one legal sequential witness")
		}
		if submitted.Contradiction.Kind != "none" || len(submitted.Contradiction.OperationIDs) != 0 || len(submitted.Contradiction.EvidenceIDs) != 0 {
			violations = append(violations, prefix+"legal history must not claim a contradiction")
		}
		return violations
	}
	if len(submitted.WitnessOrder) != 0 {
		violations = append(violations, prefix+"history has no legal witness, but the review supplies one")
	}
	if submitted.Contradiction.Kind != truth.Contradiction.Kind ||
		!sameStringSet(submitted.Contradiction.OperationIDs, truth.Contradiction.OperationIDs) {
		violations = append(violations, prefix+"contradiction does not identify the smallest unsupported history claim")
	}
	violations = append(violations, validateEvidence(prefix+"contradiction", submitted.Contradiction.EvidenceIDs,
		truth.Contradiction.EvidenceIDs, truth.Contradiction.EvidenceIDs, events)...)
	return violations
}

func validateClaimsPresence(claims *Claims) error {
	if claims == nil {
		return fmt.Errorf("review claims section is missing")
	}
	checks := []struct {
		name  string
		value *bool
	}{
		{name: "session_visibility_establishes_object_linearizability", value: claims.SessionVisibilityEstablishesObjectLinearizability},
		{name: "object_history_establishes_session_visibility", value: claims.ObjectHistoryEstablishesSessionVisibility},
		{name: "one_dependency_establishes_general_causal_consistency", value: claims.OneDependencyEstablishesGeneralCausalConsistency},
		{name: "one_legal_history_guarantees_future_availability", value: claims.OneLegalHistoryGuaranteesFutureAvailability},
	}
	for _, check := range checks {
		if check.value == nil {
			return fmt.Errorf("review claim %s is missing", check.name)
		}
	}
	return nil
}

func validateClaims(claims *Claims) []string {
	var violations []string
	if err := validateClaimsPresence(claims); err != nil {
		return []string{err.Error()}
	}
	if *claims.SessionVisibilityEstablishesObjectLinearizability {
		violations = append(violations, "session visibility proof does not establish object-wide linearizability")
	}
	if *claims.ObjectHistoryEstablishesSessionVisibility {
		violations = append(violations, "one legal object history does not establish a session visibility guarantee")
	}
	if *claims.OneDependencyEstablishesGeneralCausalConsistency {
		violations = append(violations, "one scoped dependency does not establish general causal consistency")
	}
	if *claims.OneLegalHistoryGuaranteesFutureAvailability {
		violations = append(violations, "one legal history does not guarantee future availability")
	}
	return violations
}

func indexVisibilityEvents(history VisibilityHistory) (map[string]VisibilityEvent, []string) {
	events := map[string]VisibilityEvent{}
	operationEvents := map[string]string{}
	var violations []string
	for index, event := range history.Events {
		if event.ID == "" || event.Order != index+1 {
			violations = append(violations, fmt.Sprintf("event %d lacks a stable id or order", index+1))
		}
		if _, exists := events[event.ID]; exists {
			violations = append(violations, fmt.Sprintf("duplicate event id %s", event.ID))
		}
		events[event.ID] = event
		if event.OperationID != "" {
			key := event.OperationID + "\x00" + event.Kind
			if earlierID, exists := operationEvents[key]; exists {
				violations = append(violations, fmt.Sprintf("operation %s has duplicate %s events %s and %s", event.OperationID, event.Kind, earlierID, event.ID))
			} else {
				operationEvents[key] = event.ID
			}
		}
	}
	return events, violations
}

func indexVisibilityReviews(reviews []VisibilityReview) (map[string]VisibilityReview, []string) {
	indexed := map[string]VisibilityReview{}
	var violations []string
	for _, review := range reviews {
		if review.HistoryID == "" {
			violations = append(violations, "visibility review has an empty history id")
			continue
		}
		if _, exists := indexed[review.HistoryID]; exists {
			violations = append(violations, fmt.Sprintf("duplicate visibility review for %s", review.HistoryID))
		}
		indexed[review.HistoryID] = review
	}
	return indexed, violations
}

func indexObjectReviews(reviews []ObjectReview) (map[string]ObjectReview, []string) {
	indexed := map[string]ObjectReview{}
	var violations []string
	for _, review := range reviews {
		if review.HistoryID == "" {
			violations = append(violations, "object review has an empty history id")
			continue
		}
		if _, exists := indexed[review.HistoryID]; exists {
			violations = append(violations, fmt.Sprintf("duplicate object review for %s", review.HistoryID))
		}
		indexed[review.HistoryID] = review
	}
	return indexed, violations
}

func indexFactRequirements(facts []FactRequirement) (map[string]FactRequirement, bool) {
	indexed := map[string]FactRequirement{}
	duplicate := false
	for _, fact := range facts {
		if _, exists := indexed[fact.FactID]; exists {
			duplicate = true
		}
		indexed[fact.FactID] = fact
	}
	return indexed, duplicate
}

func compareChanges(prefix string, actual []ExplainingChange, expected []changeTruth, events map[string]bool) []string {
	var violations []string
	if len(actual) != len(expected) {
		violations = append(violations, prefix+"later explaining changes are incomplete or unsupported")
	}
	for _, want := range expected {
		found := false
		for _, candidate := range actual {
			if candidate.FactID != want.FactID || candidate.AffectsFactID != want.AffectsFactID {
				continue
			}
			found = true
			violations = append(violations, validateEvidence(prefix+"explaining change "+want.FactID,
				candidate.EvidenceIDs, want.RequiredEvidence, want.AllowedEvidence, events)...)
		}
		if !found {
			violations = append(violations, prefix+fmt.Sprintf("later explaining change %s is missing", want.FactID))
		}
	}
	return violations
}

func comparePaths(prefix string, actual, expected []InformationPath, events map[string]bool) []string {
	var violations []string
	if len(actual) != len(expected) {
		violations = append(violations, prefix+"information paths are incomplete or unsupported")
	}
	for _, want := range expected {
		found := false
		for _, candidate := range actual {
			if candidate.FromFactID != want.FromFactID || candidate.ToOperationID != want.ToOperationID ||
				candidate.PredicateID != want.PredicateID || candidate.InputRole != want.InputRole {
				continue
			}
			found = true
			violations = append(violations, validateEvidence(prefix+"information path "+want.FromFactID+"->"+want.ToOperationID,
				candidate.EvidenceIDs, want.EvidenceIDs, want.EvidenceIDs, events)...)
		}
		if !found {
			violations = append(violations, prefix+"path exists only as connectivity or omits declared dependency relevance")
		}
	}
	return violations
}

func compareNonDependencies(prefix string, actual, expected []ChronologicalNonDep, events map[string]bool) []string {
	var violations []string
	if len(actual) != len(expected) {
		violations = append(violations, prefix+"chronological non-dependencies are incomplete or unsupported")
	}
	for _, want := range expected {
		found := false
		for _, candidate := range actual {
			if candidate.FactID != want.FactID || candidate.Reason != want.Reason {
				continue
			}
			found = true
			violations = append(violations, validateEvidence(prefix+"chronological non-dependency "+want.FactID,
				candidate.EvidenceIDs, want.EvidenceIDs, want.EvidenceIDs, events)...)
		}
		if !found {
			violations = append(violations, prefix+fmt.Sprintf("chronological fact %s receives unsupported dependency credit", want.FactID))
		}
	}
	return violations
}

func validateEvidence(label string, actual, required, allowed []string, events map[string]bool) []string {
	var violations []string
	if hasDuplicates(actual) {
		violations = append(violations, label+" repeats evidence ids")
	}
	for _, id := range actual {
		if !events[id] {
			violations = append(violations, label+fmt.Sprintf(" references absent evidence %s", id))
		}
		if !contains(allowed, id) {
			violations = append(violations, label+fmt.Sprintf(" cites irrelevant evidence %s", id))
		}
	}
	for _, id := range required {
		if !contains(actual, id) {
			violations = append(violations, label+fmt.Sprintf(" omits required evidence %s", id))
		}
	}
	return violations
}

func sameEdges(actual, expected []ForcedEdge) bool {
	if len(actual) != len(expected) {
		return false
	}
	actualKeys := make([]string, 0, len(actual))
	expectedKeys := make([]string, 0, len(expected))
	for _, edge := range actual {
		actualKeys = append(actualKeys, edgeKey(edge))
	}
	for _, edge := range expected {
		expectedKeys = append(expectedKeys, edgeKey(edge))
	}
	sort.Strings(actualKeys)
	sort.Strings(expectedKeys)
	return sameStringSlice(actualKeys, expectedKeys)
}

func edgeByKey(edges []ForcedEdge, key string) (ForcedEdge, bool) {
	for _, edge := range edges {
		if edgeKey(edge) == key {
			return edge, true
		}
	}
	return ForcedEdge{}, false
}

func edgeKey(edge ForcedEdge) string { return edge.Before + "->" + edge.After }

func findVisibilityEventByOperation(events []VisibilityEvent, operationID, kind string) (VisibilityEvent, bool) {
	for _, event := range events {
		if event.OperationID == operationID && event.Kind == kind {
			return event, true
		}
	}
	return VisibilityEvent{}, false
}

func findVisibilityEventByFact(events []VisibilityEvent, factID, kind string) (VisibilityEvent, bool) {
	for _, event := range events {
		if event.FactID == factID && event.Kind == kind {
			return event, true
		}
	}
	return VisibilityEvent{}, false
}

func containsFactTruth(facts []factTruth, id string) bool {
	for _, fact := range facts {
		if fact.FactID == id {
			return true
		}
	}
	return false
}

func operationByID(operations []completedOperation, id string) (completedOperation, bool) {
	for _, operation := range operations {
		if operation.ID == id {
			return operation, true
		}
	}
	return completedOperation{}, false
}

func operationIDs(operations []completedOperation) []string {
	ids := make([]string, 0, len(operations))
	for _, operation := range operations {
		ids = append(ids, operation.ID)
	}
	return ids
}

func respectsEdges(order []string, edges []ForcedEdge) bool {
	positions := map[string]int{}
	for index, id := range order {
		positions[id] = index
	}
	for _, edge := range edges {
		if positions[edge.Before] >= positions[edge.After] {
			return false
		}
	}
	return true
}

func permutations(values []string) [][]string {
	if len(values) == 0 {
		return [][]string{{}}
	}
	var result [][]string
	for index, value := range values {
		rest := append([]string{}, values[:index]...)
		rest = append(rest, values[index+1:]...)
		for _, suffix := range permutations(rest) {
			candidate := append([]string{value}, suffix...)
			result = append(result, candidate)
		}
	}
	return result
}

func containsWitness(witnesses [][]string, candidate []string) bool {
	for _, witness := range witnesses {
		if sameStringSlice(witness, candidate) {
			return true
		}
	}
	return false
}

func sameStringSet(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	leftCopy := append([]string{}, left...)
	rightCopy := append([]string{}, right...)
	sort.Strings(leftCopy)
	sort.Strings(rightCopy)
	return sameStringSlice(leftCopy, rightCopy)
}

func sameStringSlice(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func hasDuplicates(values []string) bool {
	seen := map[string]bool{}
	for _, value := range values {
		if seen[value] {
			return true
		}
		seen[value] = true
	}
	return false
}

func keys(values map[string]bool) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func keysObjectEvents(values map[string]ObjectEvent) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func prefixViolations(prefix string, violations []string) []string {
	result := make([]string, 0, len(violations))
	for _, violation := range violations {
		result = append(result, prefix+": "+violation)
	}
	return result
}
