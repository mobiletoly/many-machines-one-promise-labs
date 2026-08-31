package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mobiletoly/many-machines-one-promise-labs/practices/05-what-may-the-product-claim/replay"
)

const (
	defaultContract   = "practices/05-what-may-the-product-claim/contract.json"
	defaultDesign     = "practices/05-what-may-the-product-claim/workflow-design.json"
	defaultEvidence   = "practices/05-what-may-the-product-claim/authority-evidence.jsonl"
	defaultCandidates = "practices/05-what-may-the-product-claim/claim-candidates.json"
)

func main() {
	reviewPath := flag.String("review", "", "path to the reader-owned product claim review JSON")
	showViolations := flag.Bool("show-violations", false, "print field-level verifier diagnostics")
	flag.Parse()
	if *reviewPath == "" {
		fatalf("-review is required")
	}

	contract, err := replay.LoadContract(defaultContract)
	if err != nil {
		fatalf("%v", err)
	}
	design, err := replay.LoadWorkflowDesign(defaultDesign)
	if err != nil {
		fatalf("%v", err)
	}
	evidence, err := replay.LoadEvidence(defaultEvidence, design)
	if err != nil {
		fatalf("%v", err)
	}
	candidates, err := replay.LoadClaimCandidates(defaultCandidates, contract, design)
	if err != nil {
		fatalf("%v", err)
	}
	review, err := replay.LoadReview(*reviewPath)
	if err != nil {
		fatalf("%v", err)
	}

	evaluation := replay.Evaluate(review, contract, design, evidence, candidates)
	if len(evaluation.Violations) != 0 {
		fmt.Println("property violated: product claim review exceeds or omits the declared evidence")
		if *showViolations {
			for _, violation := range evaluation.Violations {
				fmt.Printf("- %s\n", violation)
			}
		}
		os.Exit(1)
	}

	fmt.Printf("review: %s\n", review.Name)
	fmt.Printf("evidence cuts: %d\n", len(evaluation.CutTruths))
	fmt.Printf("architecture boundaries: %d\n", len(evaluation.ArchitectureTruths))
	fmt.Println("Practice 05 product claim review satisfies the declared evidence.")
}

func fatalf(format string, arguments ...any) {
	fmt.Fprintf(os.Stderr, "replay: "+format+"\n", arguments...)
	os.Exit(2)
}
