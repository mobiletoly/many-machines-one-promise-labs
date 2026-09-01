package replay

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/mobiletoly/many-machines-one-promise-labs/practices/06-atomic-commit-adjudication/internal/strictjson"
)

const (
	PreparedValid       = "VALID"
	PreparedInvalid     = "INVALID"
	PreparedNotPrepared = "NOT_PREPARED"

	DecisionCommit         = "COMMIT"
	DecisionAbort          = "ABORT"
	DecisionNoneThroughCut = "NONE_THROUGH_CUT"
	DecisionUnknown        = "UNKNOWN"

	ContractValid  = "VALID"
	ContractBreach = "BREACH"

	DispositionApplyCommit    = "APPLY_COMMIT"
	DispositionApplyAbort     = "APPLY_ABORT"
	DispositionRemainPrepared = "REMAIN_PREPARED"
	DispositionReportBreach   = "REPORT_PROTOCOL_BREACH"
)

type Contract struct {
	ID                       string            `json:"id"`
	EvidenceAuthority        EvidenceAuthority `json:"evidence_authority"`
	PreparedStatuses         []string          `json:"prepared_statuses"`
	DecisionStatuses         []string          `json:"decision_statuses"`
	DecisionContractStatuses []string          `json:"decision_contract_statuses"`
	Dispositions             []string          `json:"dispositions"`
}

type EvidenceAuthority struct {
	ParticipantLogsCompleteThroughNamedCut          bool  `json:"participant_logs_complete_through_named_cut"`
	ParticipantEventsAuthoritative                  bool  `json:"participant_events_authoritative"`
	DecisionRecordsAuthoritative                    bool  `json:"decision_records_authoritative"`
	CompleteDecisionReadsEstablishAbsenceThroughCut bool  `json:"complete_decision_reads_establish_absence_through_cut"`
	UnavailableDecisionReadsEstablishAbsence        *bool `json:"unavailable_decision_reads_establish_absence"`
	CoordinatorNotesAreDecisionEvidence             *bool `json:"coordinator_notes_are_decision_evidence"`
}

type CaseSet struct {
	Cases []CaseDefinition `json:"cases"`
}

type CaseDefinition struct {
	ID                   string   `json:"id"`
	CutID                string   `json:"cut_id"`
	ThroughOrder         int      `json:"through_order"`
	TargetParticipant    string   `json:"target_participant"`
	RequiredParticipants []string `json:"required_participants"`
}

type ParticipantEvent struct {
	ID            string `json:"id"`
	CaseID        string `json:"case_id"`
	CutID         string `json:"cut_id"`
	ParticipantID string `json:"participant_id"`
	Order         int    `json:"order"`
	Kind          string `json:"kind"`
}

type DecisionEvent struct {
	ID                     string   `json:"id"`
	CaseID                 string   `json:"case_id"`
	CutID                  string   `json:"cut_id"`
	Source                 string   `json:"source"`
	Order                  int      `json:"order"`
	Kind                   string   `json:"kind"`
	Value                  string   `json:"value,omitempty"`
	PreparationEvidenceIDs []string `json:"preparation_evidence_ids,omitempty"`
	RecordID               string   `json:"record_id,omitempty"`
}

type Review struct {
	Name        string       `json:"name"`
	CaseReviews []CaseReview `json:"case_reviews"`
}

type CaseReview struct {
	CaseID                 string         `json:"case_id"`
	TargetParticipant      string         `json:"target_participant"`
	RequiredParticipants   []string       `json:"required_participants"`
	PreparedStatus         EvidenceStatus `json:"prepared_status"`
	DecisionStatus         EvidenceStatus `json:"decision_status"`
	DecisionContractStatus EvidenceStatus `json:"decision_contract_status"`
	Disposition            string         `json:"disposition"`
	SupportingEvidenceIDs  []string       `json:"supporting_evidence_ids"`
}

type EvidenceStatus struct {
	Value       string   `json:"value"`
	EvidenceIDs []string `json:"evidence_ids"`
}

type CaseTruth struct {
	CaseID                 string
	TargetParticipant      string
	RequiredParticipants   []string
	PreparedStatus         EvidenceStatus
	DecisionStatus         EvidenceStatus
	DecisionContractStatus EvidenceStatus
	Disposition            string
	SupportingEvidenceIDs  []string
}

type Corpus struct {
	Contract          Contract
	Cases             CaseSet
	ParticipantEvents []ParticipantEvent
	DecisionEvents    []DecisionEvent
}

type Evaluation struct {
	Truths     []CaseTruth
	Violations []string
}

func LoadCorpus(directory string) (Corpus, error) {
	var corpus Corpus
	if err := strictjson.Load(directory+"/contract.json", &corpus.Contract); err != nil {
		return Corpus{}, fmt.Errorf("load contract: %w", err)
	}
	if err := strictjson.Load(directory+"/cases.json", &corpus.Cases); err != nil {
		return Corpus{}, fmt.Errorf("load cases: %w", err)
	}
	if err := loadJSONLines(directory+"/participant-evidence.jsonl", &corpus.ParticipantEvents); err != nil {
		return Corpus{}, fmt.Errorf("load participant evidence: %w", err)
	}
	if err := loadJSONLines(directory+"/decision-evidence.jsonl", &corpus.DecisionEvents); err != nil {
		return Corpus{}, fmt.Errorf("load decision evidence: %w", err)
	}
	if err := ValidateCorpus(corpus); err != nil {
		return Corpus{}, err
	}
	return corpus, nil
}

func LoadReview(path string) (Review, error) {
	var review Review
	if err := strictjson.Load(path, &review); err != nil {
		return Review{}, err
	}
	return review, nil
}

func loadJSONLines[T any](path string, target *[]T) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Bytes()
		if strings.TrimSpace(string(line)) == "" {
			return fmt.Errorf("line %d is blank", lineNumber)
		}
		var record T
		if err := strictjson.Unmarshal(line, &record); err != nil {
			return fmt.Errorf("line %d: %w", lineNumber, err)
		}
		*target = append(*target, record)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if lineNumber == 0 {
		return fmt.Errorf("file is empty")
	}
	return nil
}

func ValidateCorpus(corpus Corpus) error {
	if err := validateContract(corpus.Contract); err != nil {
		return err
	}

	caseByID := map[string]CaseDefinition{}
	for _, definition := range corpus.Cases.Cases {
		if definition.ID == "" || definition.CutID == "" || definition.ThroughOrder <= 0 || definition.TargetParticipant == "" {
			return fmt.Errorf("case definition is incomplete")
		}
		if _, exists := caseByID[definition.ID]; exists {
			return fmt.Errorf("duplicate case %s", definition.ID)
		}
		if !uniqueNonempty(definition.RequiredParticipants) || !contains(definition.RequiredParticipants, definition.TargetParticipant) {
			return fmt.Errorf("case %s has invalid required participant inventory", definition.ID)
		}
		caseByID[definition.ID] = definition
	}
	if len(caseByID) == 0 {
		return fmt.Errorf("no cases declared")
	}

	evidenceIDs := map[string]struct{}{}
	participantOrder := map[string]struct{}{}
	for _, event := range corpus.ParticipantEvents {
		definition, ok := caseByID[event.CaseID]
		if !ok || event.CutID != definition.CutID || !contains(definition.RequiredParticipants, event.ParticipantID) {
			return fmt.Errorf("participant event %s does not belong to its declared case", event.ID)
		}
		if event.ID == "" || event.Order <= 0 || event.Order > definition.ThroughOrder {
			return fmt.Errorf("participant event %s has invalid identity or order", event.ID)
		}
		if !contains([]string{"prepared", "committed", "aborted", "resource_released", "cut_complete"}, event.Kind) {
			return fmt.Errorf("participant event %s has unknown kind %s", event.ID, event.Kind)
		}
		if err := addUnique(evidenceIDs, event.ID, "evidence id"); err != nil {
			return err
		}
		orderKey := fmt.Sprintf("%s/%s/%d", event.CaseID, event.ParticipantID, event.Order)
		if err := addUnique(participantOrder, orderKey, "participant event order"); err != nil {
			return err
		}
	}

	decisionOrder := map[string]struct{}{}
	for _, event := range corpus.DecisionEvents {
		definition, ok := caseByID[event.CaseID]
		if !ok || event.CutID != definition.CutID || event.ID == "" || event.Order <= 0 || event.Order > definition.ThroughOrder {
			return fmt.Errorf("decision event %s does not belong to its declared case", event.ID)
		}
		if err := validateDecisionShape(event); err != nil {
			return err
		}
		if err := addUnique(evidenceIDs, event.ID, "evidence id"); err != nil {
			return err
		}
		orderKey := fmt.Sprintf("%s/%s/%d", event.CaseID, event.Source, event.Order)
		if err := addUnique(decisionOrder, orderKey, "decision event order"); err != nil {
			return err
		}
	}

	for _, definition := range corpus.Cases.Cases {
		if err := validateCaseEvidence(definition, corpus.ParticipantEvents, corpus.DecisionEvents); err != nil {
			return err
		}
	}
	return nil
}

func validateContract(contract Contract) error {
	authority := contract.EvidenceAuthority
	if contract.ID != "atomic-commit-recovery-v1" ||
		!authority.ParticipantLogsCompleteThroughNamedCut ||
		!authority.ParticipantEventsAuthoritative ||
		!authority.DecisionRecordsAuthoritative ||
		!authority.CompleteDecisionReadsEstablishAbsenceThroughCut ||
		authority.UnavailableDecisionReadsEstablishAbsence == nil ||
		*authority.UnavailableDecisionReadsEstablishAbsence ||
		authority.CoordinatorNotesAreDecisionEvidence == nil ||
		*authority.CoordinatorNotesAreDecisionEvidence {
		return fmt.Errorf("contract authority declarations drifted")
	}
	if !sameSet(contract.PreparedStatuses, []string{PreparedValid, PreparedInvalid, PreparedNotPrepared}) ||
		!sameSet(contract.DecisionStatuses, []string{DecisionCommit, DecisionAbort, DecisionNoneThroughCut, DecisionUnknown}) ||
		!sameSet(contract.DecisionContractStatuses, []string{ContractValid, ContractBreach}) ||
		!sameSet(contract.Dispositions, []string{DispositionApplyCommit, DispositionApplyAbort, DispositionRemainPrepared, DispositionReportBreach}) {
		return fmt.Errorf("contract vocabulary drifted")
	}
	return nil
}

func validateDecisionShape(event DecisionEvent) error {
	switch event.Kind {
	case "decision_recorded":
		if event.Source != "D" || !contains([]string{DecisionCommit, DecisionAbort}, event.Value) || event.RecordID != "" {
			return fmt.Errorf("decision record %s has invalid shape", event.ID)
		}
	case "decision_read_complete":
		if event.Source != "D" || event.Value != "" || len(event.PreparationEvidenceIDs) != 0 {
			return fmt.Errorf("complete decision read %s has invalid shape", event.ID)
		}
	case "decision_read_unavailable":
		if event.Source != "D" || event.Value != "" || event.RecordID != "" || len(event.PreparationEvidenceIDs) != 0 {
			return fmt.Errorf("unavailable decision read %s has invalid shape", event.ID)
		}
	case "coordinator_note":
		if event.Source != "K" || !contains([]string{DecisionCommit, DecisionAbort}, event.Value) || event.RecordID != "" || len(event.PreparationEvidenceIDs) != 0 {
			return fmt.Errorf("coordinator note %s has invalid shape", event.ID)
		}
	default:
		return fmt.Errorf("decision event %s has unknown kind %s", event.ID, event.Kind)
	}
	return nil
}

func validateCaseEvidence(definition CaseDefinition, participantEvents []ParticipantEvent, decisionEvents []DecisionEvent) error {
	for _, participant := range definition.RequiredParticipants {
		state := "WORKING"
		cutCount := 0
		events := participantEventsFor(definition.ID, participant, participantEvents)
		for _, event := range events {
			switch event.Kind {
			case "prepared":
				if state != "WORKING" {
					return fmt.Errorf("case %s participant %s has illegal PREPARED transition", definition.ID, participant)
				}
				state = "PREPARED"
			case "resource_released":
				if state != "PREPARED" {
					return fmt.Errorf("case %s participant %s releases without PREPARED", definition.ID, participant)
				}
			case "committed":
				if state != "PREPARED" {
					return fmt.Errorf("case %s participant %s commits without PREPARED", definition.ID, participant)
				}
				state = "COMMITTED"
			case "aborted":
				if state != "WORKING" && state != "PREPARED" {
					return fmt.Errorf("case %s participant %s has illegal ABORT transition", definition.ID, participant)
				}
				state = "ABORTED"
			case "cut_complete":
				if event.Order != definition.ThroughOrder {
					return fmt.Errorf("case %s participant %s cut marker is not at the named cut", definition.ID, participant)
				}
				cutCount++
			}
		}
		if cutCount != 1 {
			return fmt.Errorf("case %s participant %s needs one complete lifecycle cut", definition.ID, participant)
		}
	}

	records := map[string]DecisionEvent{}
	readCount := 0
	for _, event := range decisionEventsFor(definition.ID, decisionEvents) {
		switch event.Kind {
		case "decision_recorded":
			if len(records) != 0 {
				return fmt.Errorf("case %s has multiple decision records", definition.ID)
			}
			records[event.ID] = event
			if !uniqueNonempty(event.PreparationEvidenceIDs) {
				return fmt.Errorf("decision record %s has duplicate preparation evidence", event.ID)
			}
			for _, evidenceID := range event.PreparationEvidenceIDs {
				prepared, ok := participantEventByID(evidenceID, participantEvents)
				if !ok || prepared.CaseID != definition.ID || prepared.Kind != "prepared" || prepared.Order >= event.Order {
					return fmt.Errorf("decision record %s cites invalid preparation evidence %s", event.ID, evidenceID)
				}
			}
		case "decision_read_complete":
			readCount++
		case "decision_read_unavailable":
			readCount++
		}
	}
	if readCount != 1 {
		return fmt.Errorf("case %s needs exactly one D read result", definition.ID)
	}
	for _, event := range decisionEventsFor(definition.ID, decisionEvents) {
		if event.Kind != "decision_read_complete" {
			continue
		}
		if event.RecordID == "" {
			if len(records) != 0 {
				return fmt.Errorf("case %s complete D read omits an available decision record", definition.ID)
			}
			continue
		}
		record, ok := records[event.RecordID]
		if !ok || record.Order >= event.Order {
			return fmt.Errorf("complete D read %s cites an unavailable decision record", event.ID)
		}
	}
	return nil
}

func Evaluate(corpus Corpus, review Review) (Evaluation, error) {
	if err := ValidateCorpus(corpus); err != nil {
		return Evaluation{}, err
	}
	truths, err := DeriveTruths(corpus)
	if err != nil {
		return Evaluation{}, err
	}
	violations := compareReview(truths, review)
	return Evaluation{Truths: truths, Violations: violations}, nil
}

func DeriveTruths(corpus Corpus) ([]CaseTruth, error) {
	truths := make([]CaseTruth, 0, len(corpus.Cases.Cases))
	for _, definition := range corpus.Cases.Cases {
		truth, err := deriveCase(definition, corpus.ParticipantEvents, corpus.DecisionEvents)
		if err != nil {
			return nil, err
		}
		truths = append(truths, truth)
	}
	return truths, nil
}

func deriveCase(definition CaseDefinition, participantEvents []ParticipantEvent, decisionEvents []DecisionEvent) (CaseTruth, error) {
	targetStatus, err := derivePreparedStatus(definition, definition.TargetParticipant, participantEvents)
	if err != nil {
		return CaseTruth{}, err
	}
	decisionStatus, decisionRecord, readEvent := deriveDecisionStatus(definition, decisionEvents)
	contractStatus := deriveDecisionContractStatus(definition, targetStatus, decisionRecord, readEvent, participantEvents)

	disposition := DispositionRemainPrepared
	if contractStatus.Value == ContractBreach {
		disposition = DispositionReportBreach
	} else if decisionStatus.Value == DecisionCommit && targetStatus.Value == PreparedValid {
		disposition = DispositionApplyCommit
	} else if decisionStatus.Value == DecisionAbort && targetStatus.Value == PreparedValid {
		disposition = DispositionApplyAbort
	}

	supporting := unionSets(targetStatus.EvidenceIDs, decisionStatus.EvidenceIDs, contractStatus.EvidenceIDs)
	return CaseTruth{
		CaseID:                 definition.ID,
		TargetParticipant:      definition.TargetParticipant,
		RequiredParticipants:   append([]string(nil), definition.RequiredParticipants...),
		PreparedStatus:         targetStatus,
		DecisionStatus:         decisionStatus,
		DecisionContractStatus: contractStatus,
		Disposition:            disposition,
		SupportingEvidenceIDs:  supporting,
	}, nil
}

func derivePreparedStatus(definition CaseDefinition, participant string, events []ParticipantEvent) (EvidenceStatus, error) {
	participantEvents := participantEventsFor(definition.ID, participant, events)
	state := "WORKING"
	preparedID := ""
	releases := []string{}
	terminalID := ""
	cutID := ""
	for _, event := range participantEvents {
		switch event.Kind {
		case "prepared":
			state = "PREPARED"
			preparedID = event.ID
		case "resource_released":
			releases = append(releases, event.ID)
		case "committed":
			state = "COMMITTED"
			terminalID = event.ID
		case "aborted":
			state = "ABORTED"
			terminalID = event.ID
		case "cut_complete":
			cutID = event.ID
		}
	}
	if cutID == "" {
		return EvidenceStatus{}, fmt.Errorf("case %s participant %s lacks a complete cut", definition.ID, participant)
	}
	if state != "PREPARED" {
		ids := []string{cutID}
		if terminalID != "" {
			ids = append(ids, terminalID)
		}
		return EvidenceStatus{Value: PreparedNotPrepared, EvidenceIDs: sortedUnique(ids)}, nil
	}
	if len(releases) != 0 {
		ids := append([]string{preparedID}, releases...)
		ids = append(ids, cutID)
		return EvidenceStatus{Value: PreparedInvalid, EvidenceIDs: sortedUnique(ids)}, nil
	}
	return EvidenceStatus{Value: PreparedValid, EvidenceIDs: sortedUnique([]string{preparedID, cutID})}, nil
}

func deriveDecisionStatus(definition CaseDefinition, events []DecisionEvent) (EvidenceStatus, *DecisionEvent, DecisionEvent) {
	caseEvents := decisionEventsFor(definition.ID, events)
	var read DecisionEvent
	records := map[string]DecisionEvent{}
	for _, event := range caseEvents {
		switch event.Kind {
		case "decision_recorded":
			records[event.ID] = event
		case "decision_read_complete", "decision_read_unavailable":
			read = event
		}
	}
	if read.Kind == "decision_read_unavailable" {
		return EvidenceStatus{Value: DecisionUnknown, EvidenceIDs: []string{read.ID}}, nil, read
	}
	if read.RecordID == "" {
		return EvidenceStatus{Value: DecisionNoneThroughCut, EvidenceIDs: []string{read.ID}}, nil, read
	}
	record := records[read.RecordID]
	return EvidenceStatus{Value: record.Value, EvidenceIDs: sortedUnique([]string{record.ID, read.ID})}, &record, read
}

func deriveDecisionContractStatus(definition CaseDefinition, target EvidenceStatus, record *DecisionEvent, read DecisionEvent, events []ParticipantEvent) EvidenceStatus {
	if record == nil {
		return EvidenceStatus{Value: ContractValid, EvidenceIDs: []string{read.ID}}
	}
	if record.Value == DecisionAbort {
		if target.Value == PreparedValid {
			return EvidenceStatus{Value: ContractValid, EvidenceIDs: []string{record.ID}}
		}
		return EvidenceStatus{Value: ContractBreach, EvidenceIDs: unionSets([]string{record.ID}, target.EvidenceIDs)}
	}

	evidence := []string{record.ID}
	breach := target.Value != PreparedValid
	for _, participant := range definition.RequiredParticipants {
		preparedID, invalidatingIDs, valid := validPreparationAtDecision(definition.ID, participant, *record, events)
		if valid {
			evidence = append(evidence, preparedID)
			continue
		}
		breach = true
		if preparedID != "" {
			evidence = append(evidence, preparedID)
		}
		evidence = append(evidence, invalidatingIDs...)
		if preparedID == "" {
			for _, event := range participantEventsFor(definition.ID, participant, events) {
				if event.Kind == "aborted" || event.Kind == "cut_complete" {
					evidence = append(evidence, event.ID)
				}
			}
		}
	}
	if breach {
		return EvidenceStatus{Value: ContractBreach, EvidenceIDs: sortedUnique(evidence)}
	}
	return EvidenceStatus{Value: ContractValid, EvidenceIDs: sortedUnique(evidence)}
}

func validPreparationAtDecision(caseID, participant string, record DecisionEvent, events []ParticipantEvent) (string, []string, bool) {
	preparedID := ""
	invalidating := []string{}
	for _, event := range participantEventsFor(caseID, participant, events) {
		if event.Order >= record.Order {
			continue
		}
		switch event.Kind {
		case "prepared":
			preparedID = event.ID
		case "resource_released", "aborted", "committed":
			if preparedID != "" {
				invalidating = append(invalidating, event.ID)
			}
		}
	}
	if preparedID == "" || len(invalidating) != 0 || !contains(record.PreparationEvidenceIDs, preparedID) {
		return preparedID, invalidating, false
	}
	return preparedID, nil, true
}

func compareReview(truths []CaseTruth, review Review) []string {
	violations := []string{}
	if review.Name != "operator-review" {
		violations = append(violations, "review name must be operator-review")
	}
	byCase := map[string]CaseReview{}
	for _, item := range review.CaseReviews {
		if _, exists := byCase[item.CaseID]; exists {
			violations = append(violations, fmt.Sprintf("duplicate case review %s", item.CaseID))
			continue
		}
		byCase[item.CaseID] = item
	}
	if len(review.CaseReviews) != len(truths) {
		violations = append(violations, "review must contain every declared case exactly once")
	}
	for _, truth := range truths {
		item, ok := byCase[truth.CaseID]
		if !ok {
			violations = append(violations, fmt.Sprintf("missing case review %s", truth.CaseID))
			continue
		}
		if item.TargetParticipant != truth.TargetParticipant {
			violations = append(violations, fmt.Sprintf("%s target participant is not certified", truth.CaseID))
		}
		if !sameSet(item.RequiredParticipants, truth.RequiredParticipants) {
			violations = append(violations, fmt.Sprintf("%s required participant inventory is not certified", truth.CaseID))
		}
		compareStatus := func(label string, actual, expected EvidenceStatus) {
			if actual.Value != expected.Value || !sameSet(actual.EvidenceIDs, expected.EvidenceIDs) {
				violations = append(violations, fmt.Sprintf("%s %s is not supported by the supplied evidence", truth.CaseID, label))
			}
		}
		compareStatus("prepared status", item.PreparedStatus, truth.PreparedStatus)
		compareStatus("decision status", item.DecisionStatus, truth.DecisionStatus)
		compareStatus("decision contract status", item.DecisionContractStatus, truth.DecisionContractStatus)
		if item.Disposition != truth.Disposition {
			violations = append(violations, fmt.Sprintf("%s disposition is not legal at the named cut", truth.CaseID))
		}
		if !sameSet(item.SupportingEvidenceIDs, truth.SupportingEvidenceIDs) {
			violations = append(violations, fmt.Sprintf("%s supporting evidence is incomplete or extraneous", truth.CaseID))
		}
	}
	for caseID := range byCase {
		found := false
		for _, truth := range truths {
			if truth.CaseID == caseID {
				found = true
				break
			}
		}
		if !found {
			violations = append(violations, fmt.Sprintf("unknown case review %s", caseID))
		}
	}
	sort.Strings(violations)
	return violations
}

func participantEventsFor(caseID, participant string, events []ParticipantEvent) []ParticipantEvent {
	result := []ParticipantEvent{}
	for _, event := range events {
		if event.CaseID == caseID && event.ParticipantID == participant {
			result = append(result, event)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Order == result[j].Order {
			return result[i].ID < result[j].ID
		}
		return result[i].Order < result[j].Order
	})
	return result
}

func decisionEventsFor(caseID string, events []DecisionEvent) []DecisionEvent {
	result := []DecisionEvent{}
	for _, event := range events {
		if event.CaseID == caseID {
			result = append(result, event)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Order == result[j].Order {
			return result[i].ID < result[j].ID
		}
		return result[i].Order < result[j].Order
	})
	return result
}

func participantEventByID(id string, events []ParticipantEvent) (ParticipantEvent, bool) {
	for _, event := range events {
		if event.ID == id {
			return event, true
		}
	}
	return ParticipantEvent{}, false
}

func addUnique(seen map[string]struct{}, value, label string) error {
	if value == "" {
		return fmt.Errorf("%s is empty", label)
	}
	if _, exists := seen[value]; exists {
		return fmt.Errorf("duplicate %s %s", label, value)
	}
	seen[value] = struct{}{}
	return nil
}

func uniqueNonempty(values []string) bool {
	if len(values) == 0 {
		return false
	}
	seen := map[string]struct{}{}
	for _, value := range values {
		if value == "" {
			return false
		}
		if _, exists := seen[value]; exists {
			return false
		}
		seen[value] = struct{}{}
	}
	return true
}

func sameSet(left, right []string) bool {
	if len(left) != len(right) || !uniqueNonempty(left) || !uniqueNonempty(right) {
		return false
	}
	leftSet := map[string]struct{}{}
	for _, value := range left {
		leftSet[value] = struct{}{}
	}
	for _, value := range right {
		if _, ok := leftSet[value]; !ok {
			return false
		}
	}
	return true
}

func unionSets(groups ...[]string) []string {
	seen := map[string]struct{}{}
	for _, group := range groups {
		for _, value := range group {
			if value != "" {
				seen[value] = struct{}{}
			}
		}
	}
	result := make([]string, 0, len(seen))
	for value := range seen {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func sortedUnique(values []string) []string {
	return unionSets(values)
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
