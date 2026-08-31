package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mobiletoly/many-machines-one-promise-labs/practices/03-which-histories-are-legal/replay"
)

const (
	defaultContract  = "practices/03-which-histories-are-legal/contract.json"
	defaultHistories = "practices/03-which-histories-are-legal/histories.json"
)

func main() {
	reviewPath := flag.String("review", "", "path to the reader-owned review JSON")
	showViolations := flag.Bool("show-violations", false, "print field-level verifier diagnostics")
	flag.Parse()

	if *reviewPath == "" {
		fatalf("-review is required")
	}
	contract, err := replay.LoadContract(defaultContract)
	if err != nil {
		fatalf("%v", err)
	}
	corpus, err := replay.LoadCorpus(defaultHistories)
	if err != nil {
		fatalf("%v", err)
	}
	review, err := replay.LoadReview(*reviewPath)
	if err != nil {
		fatalf("%v", err)
	}

	evaluation := replay.Evaluate(review, contract, corpus)
	printEvaluation(evaluation, *showViolations)
	if len(evaluation.Violations) != 0 {
		os.Exit(1)
	}
	fmt.Println("Practice 03 review satisfies the declared histories.")
}

func printEvaluation(evaluation replay.Evaluation, showViolations bool) {
	fmt.Printf("review: %s\n", evaluation.Review.Name)
	if len(evaluation.Violations) != 0 {
		violations := evaluation.Violations[:1]
		if showViolations {
			violations = evaluation.Violations
		}
		for _, violation := range violations {
			fmt.Printf("property violated: %s\n", violation)
		}
		return
	}
	fmt.Printf("visibility histories: %d satisfy, %d violate\n", evaluation.VisibilitySatisfies, evaluation.VisibilityViolates)
	fmt.Printf("object histories: %d satisfy, %d violate\n", evaluation.ObjectSatisfies, evaluation.ObjectViolates)
}

func fatalf(format string, arguments ...any) {
	fmt.Fprintf(os.Stderr, "replay: "+format+"\n", arguments...)
	os.Exit(2)
}
