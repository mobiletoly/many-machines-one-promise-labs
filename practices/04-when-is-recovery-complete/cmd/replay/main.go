package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mobiletoly/many-machines-one-promise-labs/practices/04-when-is-recovery-complete/replay"
)

const (
	defaultContract = "practices/04-when-is-recovery-complete/contract.json"
	defaultHistory  = "practices/04-when-is-recovery-complete/accepted-history.jsonl"
	defaultEvidence = "practices/04-when-is-recovery-complete/recovery-evidence.jsonl"
	defaultEvents   = "practices/04-when-is-recovery-complete/serving-events.jsonl"
)

func main() {
	decisionPath := flag.String("decision", "", "path to the reader-owned recovery decision JSON")
	showViolations := flag.Bool("show-violations", false, "print field-level verifier diagnostics")
	flag.Parse()
	if *decisionPath == "" {
		fatalf("-decision is required")
	}

	contract, err := replay.LoadContract(defaultContract)
	if err != nil {
		fatalf("%v", err)
	}
	facts, err := replay.LoadAcceptedHistory(defaultHistory)
	if err != nil {
		fatalf("%v", err)
	}
	states, err := replay.LoadRecoveryEvidence(defaultEvidence, facts)
	if err != nil {
		fatalf("%v", err)
	}
	events, err := replay.LoadServingEvents(defaultEvents, states)
	if err != nil {
		fatalf("%v", err)
	}
	decision, err := replay.LoadDecision(*decisionPath)
	if err != nil {
		fatalf("%v", err)
	}

	evaluation := replay.Evaluate(decision, contract, facts, states, events)
	if len(evaluation.Violations) != 0 {
		fmt.Println("property violated: recovery decision does not prove the declared completion boundary")
		if *showViolations {
			for _, violation := range evaluation.Violations {
				fmt.Printf("- %s\n", violation)
			}
		}
		os.Exit(1)
	}
	fmt.Printf("decision: %s\n", decision.Name)
	fmt.Printf("RTO end: %s at %s\n", evaluation.RTOTruth.BoundaryEvent.ID, evaluation.RTOTruth.BoundaryEvent.At)
	fmt.Println("Practice 04 recovery decision satisfies the declared evidence.")
}

func fatalf(format string, arguments ...any) {
	fmt.Fprintf(os.Stderr, "replay: "+format+"\n", arguments...)
	os.Exit(2)
}
