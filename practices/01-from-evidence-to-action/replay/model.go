package replay

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

const (
	InferenceUnknown = "unknown"

	AdmissionReject = "reject"
	AdmissionAdmit  = "admit"

	EvidenceNone             = "none"
	EvidenceHealthCheck      = "health_check"
	EvidenceShippingQuote    = "representative_shipping_quote"
	RestoreNone              = "none"
	RestoreHealthCheck       = "health_check_success"
	RestoreShippingQuote     = "shipping_quote_within_deadline"
	ClaimAdmissionAuthorized = "ordinary_admission_authorized"
)

// Policy is the reader-owned decision record. It states what the initial
// evidence establishes and how caller P changes admission while evidence is
// incomplete.
type Policy struct {
	Name                     string `json:"name"`
	InitialInference         string `json:"initial_inference"`
	OrdinaryWhileSuppressed  string `json:"ordinary_while_suppressed"`
	EvidenceAction           string `json:"evidence_action"`
	MaxEvidenceCallsPerRound int    `json:"max_evidence_calls_per_round"`
	RestoreOn                string `json:"restore_on"`
	PostRestoreClaim         string `json:"post_restore_claim"`
	CallerScope              string `json:"caller_scope"`
	OperationScope           string `json:"operation_scope"`
}

// Evidence is one retained observation. The initial evidence does not include
// execution truth from inside shipping-x.
type Evidence struct {
	At          string `json:"at"`
	Observer    string `json:"observer"`
	Kind        string `json:"kind"`
	OperationID string `json:"operation_id,omitempty"`
	Participant string `json:"participant,omitempty"`
	Detail      string `json:"detail"`
}

// TraceEvent records the evidence stream produced after the policy starts
// making admission decisions.
type TraceEvent struct {
	Round   int
	Actor   string
	Action  string
	Outcome string
}

// ScenarioReport contains one policy-dependent execution.
type ScenarioReport struct {
	Name                     string
	Events                   []TraceEvent
	Restored                 bool
	RestoredBy               string
	OrdinaryCallsSuppressed  int
	MaxEvidenceCallsInARound int
}

// Evaluation contains all deterministic scenario reports and every violated
// property. A policy passes only when Violations is empty.
type Evaluation struct {
	Evidence   []Evidence
	Reports    []ScenarioReport
	Violations []string
}

type scenario struct {
	name                  string
	quoteCapableFromRound int
	healthEndpointOK      bool
}

var scenarios = []scenario{
	{name: "alpha", quoteCapableFromRound: 4, healthEndpointOK: true},
	{name: "beta", quoteCapableFromRound: 1, healthEndpointOK: true},
	{name: "gamma", quoteCapableFromRound: 2, healthEndpointOK: true},
}

func LoadPolicy(path string) (Policy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Policy{}, fmt.Errorf("read policy: %w", err)
	}

	var policy Policy
	if err := json.Unmarshal(data, &policy); err != nil {
		return Policy{}, fmt.Errorf("parse policy: %w", err)
	}
	return policy, nil
}

func LoadEvidence(path string) ([]Evidence, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open evidence: %w", err)
	}
	defer file.Close()

	var evidence []Evidence
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var event Evidence
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return nil, fmt.Errorf("parse evidence line %d: %w", len(evidence)+1, err)
		}
		evidence = append(evidence, event)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read evidence: %w", err)
	}
	if len(evidence) == 0 {
		return nil, fmt.Errorf("evidence is empty")
	}
	return evidence, nil
}

// Evaluate replays three legal histories. The policy controls whether later
// operations reach shipping-x, so it also controls which later evidence can
// exist.
func Evaluate(policy Policy, evidence []Evidence) Evaluation {
	evaluation := Evaluation{Evidence: evidence}
	supportedInference, evidenceViolations := analyzeEvidence(evidence)
	evaluation.Violations = append(evaluation.Violations, evidenceViolations...)
	evaluation.Violations = append(
		evaluation.Violations,
		validateDecisionRecord(policy, supportedInference)...,
	)

	for _, candidate := range scenarios {
		report := runScenario(policy, candidate)
		evaluation.Reports = append(evaluation.Reports, report)
		evaluation.Violations = append(
			evaluation.Violations,
			validateScenario(policy, candidate, report)...,
		)
	}

	return evaluation
}

func PolicyApplies(policy Policy, caller, operation string) bool {
	return policy.CallerScope == caller && policy.OperationScope == operation
}

func analyzeEvidence(evidence []Evidence) (string, []string) {
	var violations []string
	quoteDeadlines := map[string]bool{}
	heartbeatMissing := 0
	participantSuspected := 0

	for index, event := range evidence {
		if event.At == "" {
			violations = append(violations, fmt.Sprintf("evidence record %d has no observation time", index+1))
		}
		switch event.Kind {
		case "shipping_quote_deadline":
			if event.Observer != "checkout-p" || event.OperationID == "" {
				violations = append(violations,
					fmt.Sprintf("evidence record %d is not a scoped checkout-p shipping-quote deadline", index+1))
				continue
			}
			if quoteDeadlines[event.OperationID] {
				violations = append(violations,
					fmt.Sprintf("shipping-quote deadline %s appears more than once", event.OperationID))
			}
			quoteDeadlines[event.OperationID] = true
		case "heartbeat_missing":
			if event.Observer != "monitor-m" || event.Participant != "shipping-x" {
				violations = append(violations,
					fmt.Sprintf("evidence record %d is not the declared monitor-m heartbeat observation", index+1))
				continue
			}
			heartbeatMissing++
		case "participant_suspected":
			if event.Observer != "detector-d" || event.Participant != "shipping-x" {
				violations = append(violations,
					fmt.Sprintf("evidence record %d is not the declared detector-d output", index+1))
				continue
			}
			participantSuspected++
		default:
			violations = append(violations,
				fmt.Sprintf("evidence record %d has unsupported kind %q", index+1, event.Kind))
		}
	}

	if len(quoteDeadlines) != 3 {
		violations = append(violations,
			fmt.Sprintf("retained evidence has %d distinct shipping-quote deadlines, want 3", len(quoteDeadlines)))
	}
	for _, operationID := range []string{"quote-101", "quote-102", "quote-103"} {
		if !quoteDeadlines[operationID] {
			violations = append(violations,
				fmt.Sprintf("retained evidence is missing shipping-quote deadline %s", operationID))
		}
	}
	if heartbeatMissing != 1 {
		violations = append(violations,
			fmt.Sprintf("retained evidence has %d monitor heartbeat observations, want 1", heartbeatMissing))
	}
	if participantSuspected != 1 {
		violations = append(violations,
			fmt.Sprintf("retained evidence has %d detector suspicion outputs, want 1", participantSuspected))
	}

	return InferenceUnknown, violations
}

func validateDecisionRecord(policy Policy, supportedInference string) []string {
	var violations []string
	if policy.InitialInference != supportedInference {
		if supportedInference == InferenceUnknown {
			violations = append(violations,
				fmt.Sprintf("initial evidence also fits a capable X with impaired evidence paths; it does not establish %s", policy.InitialInference))
		} else {
			violations = append(violations,
				fmt.Sprintf("initial inference %q does not match evidence-supported inference %q", policy.InitialInference, supportedInference))
		}
	}
	if policy.OrdinaryWhileSuppressed != AdmissionReject && policy.OrdinaryWhileSuppressed != AdmissionAdmit {
		violations = append(violations,
			fmt.Sprintf("unknown ordinary admission action %q", policy.OrdinaryWhileSuppressed))
	}
	if policy.EvidenceAction != EvidenceNone && policy.EvidenceAction != EvidenceHealthCheck && policy.EvidenceAction != EvidenceShippingQuote {
		violations = append(violations,
			fmt.Sprintf("unknown evidence action %q", policy.EvidenceAction))
	}
	if policy.MaxEvidenceCallsPerRound < 0 {
		violations = append(violations,
			fmt.Sprintf("max evidence calls per round = %d, want a non-negative bound", policy.MaxEvidenceCallsPerRound))
	}
	if policy.RestoreOn != RestoreNone && policy.RestoreOn != RestoreHealthCheck && policy.RestoreOn != RestoreShippingQuote {
		violations = append(violations,
			fmt.Sprintf("unknown restore evidence %q", policy.RestoreOn))
	}
	if policy.PostRestoreClaim != ClaimAdmissionAuthorized {
		violations = append(violations,
			fmt.Sprintf("post-restore claim %q exceeds the declared evidence boundary", policy.PostRestoreClaim))
	}
	if !PolicyApplies(policy, "checkout-p", "shipping_quote") {
		violations = append(violations,
			fmt.Sprintf("policy scope %s/%s leaves the declared checkout-p/shipping_quote boundary", policy.CallerScope, policy.OperationScope))
	}
	return violations
}

func runScenario(policy Policy, candidate scenario) ScenarioReport {
	report := ScenarioReport{Name: candidate.name}
	suppressed := true

	for round := 1; round <= 3; round++ {
		quoteCapable := round >= candidate.quoteCapableFromRound
		report.Events = append(report.Events, TraceEvent{
			Round: round, Actor: "client", Action: "shipping_quote_demand", Outcome: "arrived",
		})

		if !suppressed {
			report.Events = append(report.Events, quoteEvent(round, "ordinary_shipping_quote", quoteCapable))
			continue
		}

		if policy.OrdinaryWhileSuppressed == AdmissionAdmit {
			report.OrdinaryCallsSuppressed++
			report.Events = append(report.Events, quoteEvent(round, "ordinary_shipping_quote", quoteCapable))
		} else {
			report.Events = append(report.Events, TraceEvent{
				Round: round, Actor: "checkout-p", Action: "ordinary_shipping_quote", Outcome: "rejected_without_x_call",
			})
		}

		evidenceCalls := 0
		for call := 0; call < policy.MaxEvidenceCallsPerRound; call++ {
			switch policy.EvidenceAction {
			case EvidenceHealthCheck:
				evidenceCalls++
				outcome := "failed"
				if candidate.healthEndpointOK {
					outcome = "ok"
				}
				report.Events = append(report.Events, TraceEvent{
					Round: round, Actor: "checkout-p", Action: "health_check", Outcome: outcome,
				})
				if outcome == "ok" && policy.RestoreOn == RestoreHealthCheck {
					suppressed = false
					report.Restored = true
					report.RestoredBy = RestoreHealthCheck
				}
			case EvidenceShippingQuote:
				evidenceCalls++
				event := quoteEvent(round, "representative_shipping_quote", quoteCapable)
				report.Events = append(report.Events, event)
				if event.Outcome == "usable_within_200ms" && policy.RestoreOn == RestoreShippingQuote {
					suppressed = false
					report.Restored = true
					report.RestoredBy = RestoreShippingQuote
				}
			}
		}
		if evidenceCalls > report.MaxEvidenceCallsInARound {
			report.MaxEvidenceCallsInARound = evidenceCalls
		}
	}

	return report
}

func quoteEvent(round int, action string, capable bool) TraceEvent {
	outcome := "no_usable_result_by_200ms"
	if capable {
		outcome = "usable_within_200ms"
	}
	return TraceEvent{Round: round, Actor: "shipping-x", Action: action, Outcome: outcome}
}

func validateScenario(policy Policy, candidate scenario, report ScenarioReport) []string {
	var violations []string
	if report.OrdinaryCallsSuppressed > 0 {
		violations = append(violations,
			fmt.Sprintf("%s: %d ordinary call(s) reached X while suppression lacked representative restore evidence", candidate.name, report.OrdinaryCallsSuppressed))
	}
	if report.MaxEvidenceCallsInARound > 1 {
		violations = append(violations,
			fmt.Sprintf("%s: evidence exposure = %d calls in one round, want <= 1", candidate.name, report.MaxEvidenceCallsInARound))
	}
	if candidate.name == "alpha" && report.Restored {
		violations = append(violations,
			fmt.Sprintf("alpha: ordinary admission restored from %s while shipping quotes remained incapable", report.RestoredBy))
	}
	if (candidate.name == "beta" || candidate.name == "gamma") && !report.Restored {
		violations = append(violations,
			fmt.Sprintf("%s: quote capability became available but policy produced no recognized evidence for ordinary admission", candidate.name))
	}
	if report.Restored && report.RestoredBy != RestoreShippingQuote {
		violations = append(violations,
			fmt.Sprintf("%s: restored from %s, want shipping_quote_within_deadline", candidate.name, report.RestoredBy))
	}
	return violations
}
