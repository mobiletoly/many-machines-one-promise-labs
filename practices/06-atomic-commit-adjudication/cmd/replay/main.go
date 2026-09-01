package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mobiletoly/many-machines-one-promise-labs/practices/06-atomic-commit-adjudication/replay"
)

const practiceDirectory = "practices/06-atomic-commit-adjudication"

func main() {
	reviewPath := flag.String("review", "", "path to the reader adjudication")
	showViolations := flag.Bool("show-violations", false, "show detailed evidence mismatches")
	flag.Parse()

	if *reviewPath == "" {
		fmt.Fprintln(os.Stderr, "-review is required")
		os.Exit(2)
	}

	corpus, err := replay.LoadCorpus(practiceDirectory)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load evidence corpus: %v\n", err)
		os.Exit(2)
	}
	review, err := replay.LoadReview(*reviewPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load review: %v\n", err)
		os.Exit(2)
	}
	evaluation, err := replay.Evaluate(corpus, review)
	if err != nil {
		fmt.Fprintf(os.Stderr, "evaluate review: %v\n", err)
		os.Exit(2)
	}

	fmt.Printf("reviewed %d named evidence cuts\n", len(evaluation.Truths))
	if len(evaluation.Violations) != 0 {
		if *showViolations {
			for _, violation := range evaluation.Violations {
				fmt.Printf("violation: %s\n", violation)
			}
		}
		fmt.Println("property violated: adjudication does not certify the supplied evidence cuts")
		os.Exit(1)
	}

	fmt.Println("Practice 06 adjudication satisfies the declared atomic-commit evidence.")
}
