package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mobiletoly/many-machines-one-promise-labs/practices/01-from-evidence-to-action/replay"
)

const defaultEvidence = "practices/01-from-evidence-to-action/evidence.jsonl"

func main() {
	policyPath := flag.String("policy", "", "path to the reader-owned policy JSON")
	flag.Parse()

	if *policyPath == "" {
		fmt.Fprintln(os.Stderr, "replay: -policy is required")
		os.Exit(2)
	}

	policy, err := replay.LoadPolicy(*policyPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "replay: %v\n", err)
		os.Exit(2)
	}
	evidence, err := replay.LoadEvidence(defaultEvidence)
	if err != nil {
		fmt.Fprintf(os.Stderr, "replay: %v\n", err)
		os.Exit(2)
	}

	evaluation := replay.Evaluate(policy, evidence)
	printEvaluation(policy, evaluation)
	if len(evaluation.Violations) != 0 {
		os.Exit(1)
	}
	fmt.Println("Practice 01 policy satisfies the declared contract.")
}

func printEvaluation(policy replay.Policy, evaluation replay.Evaluation) {
	fmt.Printf("policy: %s\n", policy.Name)
	fmt.Println("initial retained evidence:")
	for _, event := range evaluation.Evidence {
		identity := event.OperationID
		if identity == "" {
			identity = event.Participant
		}
		fmt.Printf("  %s  %-11s  %-24s  %-12s  %s\n",
			event.At, event.Observer, event.Kind, identity, event.Detail)
	}

	for _, report := range evaluation.Reports {
		fmt.Printf("scenario %s:\n", report.Name)
		for _, event := range report.Events {
			fmt.Printf("  round %d  %-10s  %-32s  %s\n",
				event.Round, event.Actor, event.Action, event.Outcome)
		}
	}

	for _, violation := range evaluation.Violations {
		fmt.Printf("property violated: %s\n", violation)
	}
}
