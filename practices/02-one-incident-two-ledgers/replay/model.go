package replay

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"sort"
)

const (
	KindAttemptAccepted    = "attempt_accepted"
	KindAttemptSelected    = "attempt_selected"
	KindResultEmitted      = "result_emitted"
	KindResourceRetained   = "resource_retained"
	KindPermitWaiting      = "permit_waiting"
	KindProducerClaimPause = "producer_claim_paused"

	ScopedClaim = "browse_cannot_consume_checkout_allocation_under_declared_replay"
)

type Contract struct {
	Window                     string `json:"window"`
	WindowStartMS              int    `json:"window_start_ms"`
	WindowEndMS                int    `json:"window_end_ms"`
	TotalPermits               int    `json:"total_permits"`
	ServiceIntervalMS          int    `json:"service_interval_ms"`
	EligibleOperations         int    `json:"eligible_operations"`
	ArrivalIntervalMS          int    `json:"arrival_interval_ms"`
	GoodDeadlineMS             int    `json:"good_deadline_ms"`
	SLOTargetPercent           int    `json:"slo_target_percent"`
	RetryEveryNthOperation     int    `json:"retry_every_nth_operation"`
	RetryAfterMS               int    `json:"retry_after_ms"`
	PreWindowAttempts          int    `json:"pre_window_attempts"`
	StartBrowseRetainedPermits int    `json:"start_browse_retained_permits"`
	BoundaryBrowseDemand       int    `json:"boundary_browse_demand"`
	HealthyBrowseLatencyMS     int    `json:"healthy_browse_latency_ms"`
}

type Event struct {
	Sequence      int    `json:"sequence"`
	AtMS          int    `json:"at_ms"`
	Kind          string `json:"kind"`
	ContractValid bool   `json:"contract_valid,omitempty"`
	OperationID   string `json:"operation_id,omitempty"`
	AttemptID     string `json:"attempt_id,omitempty"`
	AttemptKind   string `json:"attempt_kind,omitempty"`
	WorkClass     string `json:"work_class,omitempty"`
	PermitID      string `json:"permit_id,omitempty"`
	State         string `json:"state,omitempty"`
}

type LoadLedger struct {
	EligibleLogicalOperations             int     `json:"eligible_logical_operations"`
	CheckoutAttemptsForEligibleOperations int     `json:"checkout_attempts_for_eligible_operations"`
	RetryAttempts                         int     `json:"retry_attempts"`
	RetryAmplification                    float64 `json:"retry_amplification"`
	PreWindowQueuedAttempts               int     `json:"pre_window_queued_attempts"`
	BrowsePermitsRetained                 int     `json:"browse_permits_retained"`
	TotalPermits                          int     `json:"total_permits"`
	CheckoutAttemptCapacityPerInterval    int     `json:"checkout_attempt_capacity_per_interval"`
}

type ProductLedger struct {
	GoodLogicalOperations      int     `json:"good_logical_operations"`
	BadLogicalOperations       int     `json:"bad_logical_operations"`
	SLIPercent                 float64 `json:"sli_percent"`
	SLOTargetPercent           int     `json:"slo_target_percent"`
	BadEventAllowance          int     `json:"bad_event_allowance"`
	BudgetExceededByOperations int     `json:"budget_exceeded_by_operations"`
	MaxFirstResultLatencyMS    int     `json:"max_first_result_latency_ms"`
	SLOPass                    bool    `json:"slo_pass"`
}

type ResourceRetention struct {
	StartDesignRetainsYAcrossX           bool   `json:"start_design_retains_y_across_x"`
	RetentionEvidenceKind                string `json:"retention_evidence_kind"`
	ProductRequiresLongTransaction       bool   `json:"product_requires_long_transaction"`
	AlternativeEvidenceSchemeMayReleaseY bool   `json:"alternative_evidence_scheme_may_release_y"`
}

type Backpressure struct {
	Present                      bool   `json:"present"`
	WaitingEvidenceKind          string `json:"waiting_evidence_kind"`
	ProducerEvidenceKind         string `json:"producer_evidence_kind"`
	FullQueueAloneIsBackpressure bool   `json:"full_queue_alone_is_backpressure"`
}

type Allocation struct {
	BrowsePermits   int  `json:"browse_permits"`
	CheckoutPermits int  `json:"checkout_permits"`
	Borrowing       bool `json:"borrowing"`
}

type ScenarioPrediction struct {
	CheckoutAttempts        int     `json:"checkout_attempts"`
	RetryAttempts           int     `json:"retry_attempts"`
	BrowsePermitsRetained   int     `json:"browse_permits_retained"`
	BrowseAttemptsWaiting   int     `json:"browse_attempts_waiting"`
	ProducerClaimPaused     bool    `json:"producer_claim_paused"`
	GoodLogicalOperations   int     `json:"good_logical_operations"`
	BadLogicalOperations    int     `json:"bad_logical_operations"`
	SLIPercent              float64 `json:"sli_percent"`
	MaxFirstResultLatencyMS int     `json:"max_first_result_latency_ms"`
	SLOPass                 bool    `json:"slo_pass"`
}

type BoundaryPrediction struct {
	DependencyXCapable       bool `json:"dependency_x_capable"`
	BrowseLatencyMS          int  `json:"browse_latency_ms"`
	SharedPoolBrowseAdmitted int  `json:"shared_pool_browse_admitted"`
	IsolatedBrowseAdmitted   int  `json:"isolated_browse_admitted"`
	IsolatedBrowseWaiting    int  `json:"isolated_browse_waiting"`
	IdleCheckoutPermits      int  `json:"idle_checkout_permits"`
	FungibilityLost          bool `json:"fungibility_lost"`
}

type Claims struct {
	ScopedClaim           string `json:"scoped_claim"`
	FutureSLOGuaranteed   bool   `json:"future_slo_guaranteed"`
	IsOnlyLegalCorrection bool   `json:"is_only_legal_correction"`
	XFailureProven        bool   `json:"x_failure_proven"`
}

type Analysis struct {
	Name                string             `json:"name"`
	LoadLedger          LoadLedger         `json:"load_ledger"`
	ProductLedger       ProductLedger      `json:"product_ledger"`
	ResourceRetention   ResourceRetention  `json:"resource_retention"`
	Backpressure        Backpressure       `json:"backpressure"`
	Allocation          Allocation         `json:"allocation"`
	CorrectedPrediction ScenarioPrediction `json:"corrected_prediction"`
	BoundaryPrediction  BoundaryPrediction `json:"boundary_prediction"`
	Claims              Claims             `json:"claims"`
}

type DerivedEvidence struct {
	LoadLedger       LoadLedger
	ProductLedger    ProductLedger
	ResourceRetained bool
	PermitWaiting    bool
	ProducerPaused   bool
}

type Evaluation struct {
	Analysis            Analysis
	Derived             DerivedEvidence
	CorrectedPrediction ScenarioPrediction
	BoundaryPrediction  BoundaryPrediction
	Violations          []string
}

type attempt struct {
	operationID   string
	attemptID     string
	attemptKind   string
	firstAccepted int
}

func LoadContract(path string) (Contract, error) {
	file, err := os.Open(path)
	if err != nil {
		return Contract{}, fmt.Errorf("open contract: %w", err)
	}
	defer file.Close()

	var contract Contract
	if err := decodeStrictJSON(file, &contract); err != nil {
		return Contract{}, fmt.Errorf("parse contract: %w", err)
	}
	if err := validateContract(contract); err != nil {
		return Contract{}, err
	}
	return contract, nil
}

func LoadAnalysis(path string) (Analysis, error) {
	file, err := os.Open(path)
	if err != nil {
		return Analysis{}, fmt.Errorf("open analysis: %w", err)
	}
	defer file.Close()

	var analysis Analysis
	if err := decodeStrictJSON(file, &analysis); err != nil {
		return Analysis{}, fmt.Errorf("parse analysis: %w", err)
	}
	return analysis, nil
}

func LoadEvidence(path string) ([]Event, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open evidence: %w", err)
	}
	defer file.Close()

	var events []Event
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var event Event
		if err := decodeStrictJSON(bytes.NewReader(scanner.Bytes()), &event); err != nil {
			return nil, fmt.Errorf("parse evidence line %d: %w", len(events)+1, err)
		}
		events = append(events, event)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read evidence: %w", err)
	}
	if len(events) == 0 {
		return nil, fmt.Errorf("evidence is empty")
	}
	return events, nil
}

func decodeStrictJSON(reader io.Reader, target any) error {
	decoder := json.NewDecoder(reader)
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

func WriteEvidence(path string, events []Event) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create evidence: %w", err)
	}
	encoder := json.NewEncoder(file)
	encoder.SetEscapeHTML(false)
	for _, event := range events {
		if err := encoder.Encode(event); err != nil {
			file.Close()
			return fmt.Errorf("write evidence: %w", err)
		}
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close evidence: %w", err)
	}
	return nil
}

func GenerateStartEvidence(contract Contract) []Event {
	checkoutEvents, _ := simulateCheckout(contract, contract.TotalPermits-contract.StartBrowseRetainedPermits)
	for index := range checkoutEvents {
		if checkoutEvents[index].Kind == KindAttemptSelected {
			checkoutEvents[index].PermitID = fmt.Sprintf("Y-%02d", contract.TotalPermits)
		}
	}

	events := make([]Event, 0, len(checkoutEvents)+contract.StartBrowseRetainedPermits+2)
	for len(checkoutEvents) > 0 && checkoutEvents[0].AtMS < contract.WindowStartMS {
		events = append(events, checkoutEvents[0])
		checkoutEvents = checkoutEvents[1:]
	}
	for permit := 1; permit <= contract.StartBrowseRetainedPermits; permit++ {
		events = append(events, Event{
			AtMS:        contract.WindowStartMS - 1,
			Kind:        KindResourceRetained,
			OperationID: fmt.Sprintf("B-%03d", permit),
			WorkClass:   "browse",
			PermitID:    fmt.Sprintf("Y-%02d", permit),
			State:       "HOLDING_Y_ACROSS_X",
		})
	}

	waitingInserted := false
	for _, event := range checkoutEvents {
		events = append(events, event)
		if !waitingInserted && event.Kind == KindAttemptSelected && event.AtMS == contract.WindowStartMS {
			events = append(events,
				Event{
					AtMS:        contract.WindowStartMS,
					Kind:        KindPermitWaiting,
					OperationID: fmt.Sprintf("B-%03d", contract.StartBrowseRetainedPermits+1),
					WorkClass:   "browse",
					State:       "WAITING_FOR_Y",
				},
				Event{
					AtMS:      contract.WindowStartMS,
					Kind:      KindProducerClaimPause,
					WorkClass: "browse",
					State:     "CLAIM_PAUSED",
				},
			)
			waitingInserted = true
		}
	}
	for index := range events {
		events[index].Sequence = index + 1
	}
	return events
}

func Evaluate(analysis Analysis, contract Contract, events []Event) Evaluation {
	derived, evidenceViolations := Derive(contract, events)
	corrected := PredictScenario(contract, analysis.Allocation)
	boundary := PredictBoundary(contract, analysis.Allocation)

	evaluation := Evaluation{
		Analysis:            analysis,
		Derived:             derived,
		CorrectedPrediction: corrected,
		BoundaryPrediction:  boundary,
	}
	evaluation.Violations = append(evaluation.Violations, evidenceViolations...)
	evaluation.Violations = append(evaluation.Violations,
		compareLoadLedger(analysis.LoadLedger, derived.LoadLedger)...)
	evaluation.Violations = append(evaluation.Violations,
		compareProductLedger(analysis.ProductLedger, derived.ProductLedger)...)
	evaluation.Violations = append(evaluation.Violations,
		validateRetention(analysis.ResourceRetention, derived)...)
	evaluation.Violations = append(evaluation.Violations,
		validateBackpressure(analysis.Backpressure, derived)...)
	evaluation.Violations = append(evaluation.Violations,
		validateAllocation(analysis.Allocation, contract, corrected)...)
	evaluation.Violations = append(evaluation.Violations,
		compareScenario("corrected_prediction", analysis.CorrectedPrediction, corrected)...)
	evaluation.Violations = append(evaluation.Violations,
		compareBoundary(analysis.BoundaryPrediction, boundary)...)
	evaluation.Violations = append(evaluation.Violations,
		validateClaims(analysis.Claims)...)

	if len(evaluation.Violations) != 0 {
		evaluation.Violations = append([]string{"analysis does not match the declared incident"}, evaluation.Violations...)
	}
	return evaluation
}

func Derive(contract Contract, events []Event) (DerivedEvidence, []string) {
	var violations []string
	canonical := GenerateStartEvidence(contract)
	if !reflect.DeepEqual(events, canonical) {
		violations = append(violations, "raw evidence differs from the declared deterministic incident")
	}

	firstAccepted := map[string]int{}
	workClass := map[string]string{}
	attemptsByOperation := map[string]int{}
	retriesByOperation := map[string]int{}
	selectedAt := map[string]int{}
	firstResult := map[string]int{}
	browsePermits := map[string]bool{}
	checkoutPermits := map[string]bool{}
	resourceRetained := false
	permitWaiting := false
	producerPaused := false

	for index, event := range events {
		if event.Sequence != index+1 {
			violations = append(violations, fmt.Sprintf("event sequence %d appears at line %d", event.Sequence, index+1))
		}
		switch event.Kind {
		case KindAttemptAccepted:
			if event.OperationID == "" || event.AttemptID == "" || event.WorkClass != "checkout" {
				violations = append(violations, fmt.Sprintf("attempt event %d lacks declared checkout identity", event.Sequence))
				continue
			}
			if previous, ok := firstAccepted[event.OperationID]; !ok || event.AtMS < previous {
				firstAccepted[event.OperationID] = event.AtMS
			}
			workClass[event.OperationID] = event.WorkClass
			attemptsByOperation[event.OperationID]++
			if event.AttemptKind == "retry" {
				retriesByOperation[event.OperationID]++
			}
		case KindAttemptSelected:
			selectedAt[event.AttemptID] = event.AtMS
			if event.PermitID != "" {
				checkoutPermits[event.PermitID] = true
			}
		case KindResultEmitted:
			if !event.ContractValid {
				continue
			}
			if previous, ok := firstResult[event.OperationID]; !ok || event.AtMS < previous {
				firstResult[event.OperationID] = event.AtMS
			}
		case KindResourceRetained:
			resourceRetained = true
			browsePermits[event.PermitID] = true
		case KindPermitWaiting:
			permitWaiting = event.State == "WAITING_FOR_Y"
		case KindProducerClaimPause:
			producerPaused = event.State == "CLAIM_PAUSED"
		default:
			violations = append(violations, fmt.Sprintf("event %d has unsupported kind %q", event.Sequence, event.Kind))
		}
	}

	eligible := map[string]bool{}
	for operation, acceptedAt := range firstAccepted {
		if workClass[operation] == "checkout" && acceptedAt >= contract.WindowStartMS && acceptedAt < contract.WindowEndMS {
			eligible[operation] = true
		}
	}

	load := LoadLedger{
		EligibleLogicalOperations:          len(eligible),
		BrowsePermitsRetained:              len(browsePermits),
		TotalPermits:                       contract.TotalPermits,
		CheckoutAttemptCapacityPerInterval: len(checkoutPermits),
	}
	for operation := range eligible {
		load.CheckoutAttemptsForEligibleOperations += attemptsByOperation[operation]
		load.RetryAttempts += retriesByOperation[operation]
	}
	if load.EligibleLogicalOperations != 0 {
		load.RetryAmplification = float64(load.CheckoutAttemptsForEligibleOperations) / float64(load.EligibleLogicalOperations)
	}
	for operation, acceptedAt := range firstAccepted {
		if acceptedAt >= contract.WindowStartMS {
			continue
		}
		for _, event := range events {
			if event.Kind == KindAttemptAccepted && event.OperationID == operation {
				if selected, ok := selectedAt[event.AttemptID]; ok && selected >= contract.WindowStartMS {
					load.PreWindowQueuedAttempts++
				}
			}
		}
	}

	product := ProductLedger{SLOTargetPercent: contract.SLOTargetPercent}
	for operation := range eligible {
		completedAt, ok := firstResult[operation]
		if !ok {
			product.BadLogicalOperations++
			continue
		}
		latency := completedAt - firstAccepted[operation]
		if latency > product.MaxFirstResultLatencyMS {
			product.MaxFirstResultLatencyMS = latency
		}
		if latency <= contract.GoodDeadlineMS {
			product.GoodLogicalOperations++
		} else {
			product.BadLogicalOperations++
		}
	}
	if len(eligible) != 0 {
		product.SLIPercent = float64(product.GoodLogicalOperations) * 100 / float64(len(eligible))
	}
	minimumGood := int(math.Ceil(float64(contract.SLOTargetPercent*len(eligible)) / 100))
	product.BadEventAllowance = len(eligible) - minimumGood
	product.BudgetExceededByOperations = max(0, product.BadLogicalOperations-product.BadEventAllowance)
	product.SLOPass = product.GoodLogicalOperations >= minimumGood

	return DerivedEvidence{
		LoadLedger:       load,
		ProductLedger:    product,
		ResourceRetained: resourceRetained,
		PermitWaiting:    permitWaiting,
		ProducerPaused:   producerPaused,
	}, violations
}

func PredictScenario(contract Contract, allocation Allocation) ScenarioPrediction {
	if allocation.CheckoutPermits <= 0 {
		return ScenarioPrediction{}
	}
	events, _ := simulateCheckout(contract, allocation.CheckoutPermits)
	derived, _ := DeriveScenario(contract, events)
	return ScenarioPrediction{
		CheckoutAttempts:        derived.LoadLedger.CheckoutAttemptsForEligibleOperations,
		RetryAttempts:           derived.LoadLedger.RetryAttempts,
		BrowsePermitsRetained:   min(allocation.BrowsePermits, contract.StartBrowseRetainedPermits),
		BrowseAttemptsWaiting:   1,
		ProducerClaimPaused:     true,
		GoodLogicalOperations:   derived.ProductLedger.GoodLogicalOperations,
		BadLogicalOperations:    derived.ProductLedger.BadLogicalOperations,
		SLIPercent:              derived.ProductLedger.SLIPercent,
		MaxFirstResultLatencyMS: derived.ProductLedger.MaxFirstResultLatencyMS,
		SLOPass:                 derived.ProductLedger.SLOPass,
	}
}

func DeriveScenario(contract Contract, events []Event) (DerivedEvidence, []string) {
	firstAccepted := map[string]int{}
	attemptsByOperation := map[string]int{}
	retriesByOperation := map[string]int{}
	firstResult := map[string]int{}
	for _, event := range events {
		switch event.Kind {
		case KindAttemptAccepted:
			if previous, ok := firstAccepted[event.OperationID]; !ok || event.AtMS < previous {
				firstAccepted[event.OperationID] = event.AtMS
			}
			attemptsByOperation[event.OperationID]++
			if event.AttemptKind == "retry" {
				retriesByOperation[event.OperationID]++
			}
		case KindResultEmitted:
			if !event.ContractValid {
				continue
			}
			if previous, ok := firstResult[event.OperationID]; !ok || event.AtMS < previous {
				firstResult[event.OperationID] = event.AtMS
			}
		}
	}
	eligible := map[string]bool{}
	for operation, at := range firstAccepted {
		if at >= contract.WindowStartMS && at < contract.WindowEndMS {
			eligible[operation] = true
		}
	}
	load := LoadLedger{EligibleLogicalOperations: len(eligible)}
	product := ProductLedger{SLOTargetPercent: contract.SLOTargetPercent}
	for operation := range eligible {
		load.CheckoutAttemptsForEligibleOperations += attemptsByOperation[operation]
		load.RetryAttempts += retriesByOperation[operation]
		completedAt, ok := firstResult[operation]
		if !ok {
			product.BadLogicalOperations++
			continue
		}
		latency := completedAt - firstAccepted[operation]
		if latency > product.MaxFirstResultLatencyMS {
			product.MaxFirstResultLatencyMS = latency
		}
		if latency <= contract.GoodDeadlineMS {
			product.GoodLogicalOperations++
		} else {
			product.BadLogicalOperations++
		}
	}
	if len(eligible) > 0 {
		load.RetryAmplification = float64(load.CheckoutAttemptsForEligibleOperations) / float64(len(eligible))
		product.SLIPercent = float64(product.GoodLogicalOperations) * 100 / float64(len(eligible))
	}
	minimumGood := int(math.Ceil(float64(contract.SLOTargetPercent*len(eligible)) / 100))
	product.BadEventAllowance = len(eligible) - minimumGood
	product.BudgetExceededByOperations = max(0, product.BadLogicalOperations-product.BadEventAllowance)
	product.SLOPass = product.GoodLogicalOperations >= minimumGood
	return DerivedEvidence{LoadLedger: load, ProductLedger: product}, nil
}

func PredictBoundary(contract Contract, allocation Allocation) BoundaryPrediction {
	sharedAdmitted := min(contract.TotalPermits, contract.BoundaryBrowseDemand)
	isolatedAdmitted := min(max(allocation.BrowsePermits, 0), contract.BoundaryBrowseDemand)
	return BoundaryPrediction{
		DependencyXCapable:       true,
		BrowseLatencyMS:          contract.HealthyBrowseLatencyMS,
		SharedPoolBrowseAdmitted: sharedAdmitted,
		IsolatedBrowseAdmitted:   isolatedAdmitted,
		IsolatedBrowseWaiting:    contract.BoundaryBrowseDemand - isolatedAdmitted,
		IdleCheckoutPermits:      max(allocation.CheckoutPermits, 0),
		FungibilityLost:          isolatedAdmitted < sharedAdmitted && allocation.CheckoutPermits > 0,
	}
}

func simulateCheckout(contract Contract, permits int) ([]Event, ScenarioPrediction) {
	var events []Event
	sequence := 0
	appendEvent := func(event Event) {
		sequence++
		event.Sequence = sequence
		events = append(events, event)
	}

	queue := make([]attempt, 0)
	for number := 1; number <= contract.PreWindowAttempts; number++ {
		acceptedAt := contract.WindowStartMS - (contract.PreWindowAttempts-number+1)*contract.ServiceIntervalMS
		candidate := attempt{
			operationID:   fmt.Sprintf("PRE-%03d", number),
			attemptID:     fmt.Sprintf("PRE-%03d-A1", number),
			attemptKind:   "initial",
			firstAccepted: acceptedAt,
		}
		queue = append(queue, candidate)
		appendEvent(Event{AtMS: acceptedAt, Kind: KindAttemptAccepted, OperationID: candidate.operationID,
			AttemptID: candidate.attemptID, AttemptKind: candidate.attemptKind, WorkClass: "checkout"})
	}

	emitted := map[string]int{}
	selected := []attempt{}
	lastRetryAt := contract.WindowEndMS - contract.ArrivalIntervalMS + contract.RetryAfterMS
	for at := contract.WindowStartMS; at <= lastRetryAt+contract.WindowEndMS; at += contract.ServiceIntervalMS {
		for _, candidate := range selected {
			if _, ok := emitted[candidate.operationID]; !ok {
				emitted[candidate.operationID] = at
			}
			appendEvent(Event{AtMS: at, Kind: KindResultEmitted, ContractValid: true, OperationID: candidate.operationID,
				AttemptID: candidate.attemptID, AttemptKind: candidate.attemptKind, WorkClass: "checkout"})
		}
		selected = nil

		if at < contract.WindowEndMS && at%contract.ArrivalIntervalMS == 0 {
			number := at/contract.ArrivalIntervalMS + 1
			candidate := attempt{
				operationID:   fmt.Sprintf("C-%03d", number),
				attemptID:     fmt.Sprintf("C-%03d-A1", number),
				attemptKind:   "initial",
				firstAccepted: at,
			}
			queue = append(queue, candidate)
			appendEvent(Event{AtMS: at, Kind: KindAttemptAccepted, OperationID: candidate.operationID,
				AttemptID: candidate.attemptID, AttemptKind: candidate.attemptKind, WorkClass: "checkout"})
		}

		firstAccepted := at - contract.RetryAfterMS
		if firstAccepted >= contract.WindowStartMS && firstAccepted < contract.WindowEndMS &&
			firstAccepted%contract.ArrivalIntervalMS == 0 {
			number := firstAccepted/contract.ArrivalIntervalMS + 1
			operationID := fmt.Sprintf("C-%03d", number)
			if number%contract.RetryEveryNthOperation == 0 {
				if _, ok := emitted[operationID]; !ok {
					candidate := attempt{
						operationID: operationID, attemptID: fmt.Sprintf("C-%03d-A2", number),
						attemptKind: "retry", firstAccepted: firstAccepted,
					}
					queue = append(queue, candidate)
					appendEvent(Event{AtMS: at, Kind: KindAttemptAccepted, OperationID: candidate.operationID,
						AttemptID: candidate.attemptID, AttemptKind: candidate.attemptKind, WorkClass: "checkout"})
				}
			}
		}

		for permit := 1; permit <= permits && len(queue) > 0; permit++ {
			candidate := queue[0]
			queue = queue[1:]
			selected = append(selected, candidate)
			appendEvent(Event{AtMS: at, Kind: KindAttemptSelected, OperationID: candidate.operationID,
				AttemptID: candidate.attemptID, AttemptKind: candidate.attemptKind, WorkClass: "checkout",
				PermitID: fmt.Sprintf("Y-CHECKOUT-%02d", permit), State: "SELECTED_FOR_INTERVAL"})
		}

		if at > lastRetryAt && len(queue) == 0 && len(selected) == 0 {
			break
		}
	}

	derived, _ := DeriveScenario(contract, events)
	prediction := ScenarioPrediction{
		CheckoutAttempts:        derived.LoadLedger.CheckoutAttemptsForEligibleOperations,
		RetryAttempts:           derived.LoadLedger.RetryAttempts,
		GoodLogicalOperations:   derived.ProductLedger.GoodLogicalOperations,
		BadLogicalOperations:    derived.ProductLedger.BadLogicalOperations,
		SLIPercent:              derived.ProductLedger.SLIPercent,
		MaxFirstResultLatencyMS: derived.ProductLedger.MaxFirstResultLatencyMS,
		SLOPass:                 derived.ProductLedger.SLOPass,
	}
	return events, prediction
}

func validateContract(contract Contract) error {
	if contract.Window == "" || contract.WindowEndMS <= contract.WindowStartMS {
		return fmt.Errorf("contract has an invalid measurement window")
	}
	if contract.TotalPermits <= 0 || contract.StartBrowseRetainedPermits >= contract.TotalPermits {
		return fmt.Errorf("contract has an invalid permit allocation")
	}
	if contract.ServiceIntervalMS <= 0 || contract.ArrivalIntervalMS <= 0 {
		return fmt.Errorf("contract has an invalid service interval")
	}
	if contract.EligibleOperations <= 0 || contract.SLOTargetPercent < 0 || contract.SLOTargetPercent > 100 {
		return fmt.Errorf("contract has an invalid product population")
	}
	if contract.WindowEndMS-contract.WindowStartMS != contract.EligibleOperations*contract.ArrivalIntervalMS {
		return fmt.Errorf("contract window does not match its operation population")
	}
	if contract.RetryEveryNthOperation <= 0 || contract.RetryAfterMS <= 0 {
		return fmt.Errorf("contract has an invalid retry rule")
	}
	if contract.HealthyBrowseLatencyMS <= 0 || contract.BoundaryBrowseDemand <= 0 {
		return fmt.Errorf("contract has an invalid boundary replay")
	}
	return nil
}

func compareLoadLedger(actual, want LoadLedger) []string {
	var violations []string
	if actual.EligibleLogicalOperations != want.EligibleLogicalOperations {
		violations = append(violations, "load_ledger.eligible_logical_operations uses the wrong population")
	}
	if actual.CheckoutAttemptsForEligibleOperations != want.CheckoutAttemptsForEligibleOperations {
		violations = append(violations, "load_ledger.checkout_attempts_for_eligible_operations is inconsistent with the raw attempts")
	}
	if actual.RetryAttempts != want.RetryAttempts {
		violations = append(violations, "load_ledger.retry_attempts is inconsistent with operation identity")
	}
	if !equalFloat(actual.RetryAmplification, want.RetryAmplification) {
		violations = append(violations, "load_ledger.retry_amplification mixes attempt and operation units")
	}
	if actual.PreWindowQueuedAttempts != want.PreWindowQueuedAttempts {
		violations = append(violations, "load_ledger.pre_window_queued_attempts omits work consuming current capacity")
	}
	if actual.BrowsePermitsRetained != want.BrowsePermitsRetained {
		violations = append(violations, "load_ledger.browse_permits_retained is inconsistent with retained-resource events")
	}
	if actual.TotalPermits != want.TotalPermits {
		violations = append(violations, "load_ledger.total_permits differs from the declared resource boundary")
	}
	if actual.CheckoutAttemptCapacityPerInterval != want.CheckoutAttemptCapacityPerInterval {
		violations = append(violations, "load_ledger.checkout_attempt_capacity_per_interval is inconsistent with permit selection")
	}
	return violations
}

func compareProductLedger(actual, want ProductLedger) []string {
	var violations []string
	if actual.GoodLogicalOperations != want.GoodLogicalOperations || actual.BadLogicalOperations != want.BadLogicalOperations {
		violations = append(violations, "product_ledger good/bad counts do not classify logical operations from first acceptance")
	}
	if !equalFloat(actual.SLIPercent, want.SLIPercent) {
		violations = append(violations, "product_ledger.sli_percent uses the wrong units or population")
	}
	if actual.SLOTargetPercent != want.SLOTargetPercent {
		violations = append(violations, "product_ledger.slo_target_percent differs from the declared objective")
	}
	if actual.BadEventAllowance != want.BadEventAllowance || actual.BudgetExceededByOperations != want.BudgetExceededByOperations {
		violations = append(violations, "product_ledger error-budget arithmetic is inconsistent with the SLO")
	}
	if actual.MaxFirstResultLatencyMS != want.MaxFirstResultLatencyMS {
		violations = append(violations, "product_ledger.max_first_result_latency_ms ignores the declared service interval")
	}
	if actual.SLOPass != want.SLOPass {
		violations = append(violations, "product_ledger.slo_pass is inconsistent with the measured window")
	}
	return violations
}

func validateRetention(actual ResourceRetention, derived DerivedEvidence) []string {
	var violations []string
	if actual.StartDesignRetainsYAcrossX != derived.ResourceRetained {
		violations = append(violations, "resource_retention does not recognize the start design's retained permit")
	}
	if actual.RetentionEvidenceKind != KindResourceRetained {
		violations = append(violations, "resource_retention cites no retained-resource event")
	}
	if actual.ProductRequiresLongTransaction {
		violations = append(violations, "the product requires compatible version evidence, not one long transaction")
	}
	if !actual.AlternativeEvidenceSchemeMayReleaseY {
		violations = append(violations, "resource_retention incorrectly excludes another authoritative evidence scheme")
	}
	return violations
}

func validateBackpressure(actual Backpressure, derived DerivedEvidence) []string {
	var violations []string
	if actual.Present != (derived.PermitWaiting && derived.ProducerPaused) {
		violations = append(violations, "backpressure requires both waiting work and producer claim adaptation")
	}
	if actual.WaitingEvidenceKind != KindPermitWaiting || actual.ProducerEvidenceKind != KindProducerClaimPause {
		violations = append(violations, "backpressure cites the wrong raw evidence kinds")
	}
	if actual.FullQueueAloneIsBackpressure {
		violations = append(violations, "a full queue alone does not establish producer adaptation")
	}
	return violations
}

func validateAllocation(actual Allocation, contract Contract, prediction ScenarioPrediction) []string {
	var violations []string
	if actual.BrowsePermits <= 0 || actual.CheckoutPermits <= 0 {
		violations = append(violations, "allocation must assign a positive permit count to both work classes")
	}
	if actual.BrowsePermits+actual.CheckoutPermits != contract.TotalPermits {
		violations = append(violations, "allocation must preserve the twelve-permit total")
	}
	if actual.Borrowing {
		violations = append(violations, "the evaluated allocation cannot borrow permits in this replay")
	}
	if !prediction.SLOPass {
		violations = append(violations, "the proposed allocation does not satisfy the declared W-2 objective")
	}
	return violations
}

func compareScenario(name string, actual, want ScenarioPrediction) []string {
	if reflect.DeepEqual(actual, want) || (actual.CheckoutAttempts == want.CheckoutAttempts &&
		actual.RetryAttempts == want.RetryAttempts &&
		actual.BrowsePermitsRetained == want.BrowsePermitsRetained &&
		actual.BrowseAttemptsWaiting == want.BrowseAttemptsWaiting &&
		actual.ProducerClaimPaused == want.ProducerClaimPaused &&
		actual.GoodLogicalOperations == want.GoodLogicalOperations &&
		actual.BadLogicalOperations == want.BadLogicalOperations &&
		equalFloat(actual.SLIPercent, want.SLIPercent) &&
		actual.MaxFirstResultLatencyMS == want.MaxFirstResultLatencyMS &&
		actual.SLOPass == want.SLOPass) {
		return nil
	}
	return []string{name + " does not match the proposed allocation's deterministic replay"}
}

func compareBoundary(actual, want BoundaryPrediction) []string {
	if reflect.DeepEqual(actual, want) {
		return nil
	}
	return []string{"boundary_prediction does not account for idle protected capacity and waiting browse work"}
}

func validateClaims(actual Claims) []string {
	var violations []string
	if actual.ScopedClaim != ScopedClaim {
		violations = append(violations, "claims.scoped_claim exceeds or misses the evaluated allocation boundary")
	}
	if actual.FutureSLOGuaranteed {
		violations = append(violations, "one passing window does not guarantee future SLO compliance")
	}
	if actual.IsOnlyLegalCorrection {
		violations = append(violations, "the evaluated allocation is not the only legal correction")
	}
	if actual.XFailureProven {
		violations = append(violations, "the incident evidence does not prove that X failed")
	}
	return violations
}

func equalFloat(left, right float64) bool {
	return math.Abs(left-right) < 0.000001
}

func SortedLateOperations(contract Contract, events []Event) []string {
	firstAccepted := map[string]int{}
	firstResult := map[string]int{}
	for _, event := range events {
		if event.Kind == KindAttemptAccepted {
			if previous, ok := firstAccepted[event.OperationID]; !ok || event.AtMS < previous {
				firstAccepted[event.OperationID] = event.AtMS
			}
		}
		if event.Kind == KindResultEmitted && event.ContractValid {
			if previous, ok := firstResult[event.OperationID]; !ok || event.AtMS < previous {
				firstResult[event.OperationID] = event.AtMS
			}
		}
	}
	var late []string
	for operation, acceptedAt := range firstAccepted {
		if acceptedAt < contract.WindowStartMS || acceptedAt >= contract.WindowEndMS {
			continue
		}
		completedAt, ok := firstResult[operation]
		if !ok || completedAt-acceptedAt > contract.GoodDeadlineMS {
			late = append(late, operation)
		}
	}
	sort.Strings(late)
	return late
}
