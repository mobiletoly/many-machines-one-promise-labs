package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mobiletoly/many-machines-one-promise-labs/practices/02-one-incident-two-ledgers/replay"
)

const (
	defaultContract = "practices/02-one-incident-two-ledgers/contract.json"
	defaultEvidence = "practices/02-one-incident-two-ledgers/evidence.jsonl"
)

func main() {
	analysisPath := flag.String("analysis", "", "path to the reader-owned analysis JSON")
	writeEvidencePath := flag.String("write-evidence", "", "write the canonical raw incident and exit")
	showViolations := flag.Bool("show-violations", false, "print field-level verifier diagnostics")
	flag.Parse()

	contract, err := replay.LoadContract(defaultContract)
	if err != nil {
		fatalf("%v", err)
	}
	if *writeEvidencePath != "" {
		if err := replay.WriteEvidence(*writeEvidencePath, replay.GenerateStartEvidence(contract)); err != nil {
			fatalf("%v", err)
		}
		fmt.Printf("wrote Practice 02 raw evidence to %s\n", *writeEvidencePath)
		return
	}
	if *analysisPath == "" {
		fatalf("-analysis is required")
	}

	analysis, err := replay.LoadAnalysis(*analysisPath)
	if err != nil {
		fatalf("%v", err)
	}
	events, err := replay.LoadEvidence(defaultEvidence)
	if err != nil {
		fatalf("%v", err)
	}

	evaluation := replay.Evaluate(analysis, contract, events)
	printEvaluation(evaluation, *showViolations)
	if len(evaluation.Violations) != 0 {
		os.Exit(1)
	}
	fmt.Println("Practice 02 analysis satisfies the declared contract.")
}

func printEvaluation(evaluation replay.Evaluation, showViolations bool) {
	fmt.Printf("analysis: %s\n", evaluation.Analysis.Name)
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

	fmt.Printf("start load ledger: %d operations, %d attempts, %.2f amplification\n",
		evaluation.Derived.LoadLedger.EligibleLogicalOperations,
		evaluation.Derived.LoadLedger.CheckoutAttemptsForEligibleOperations,
		evaluation.Derived.LoadLedger.RetryAmplification)
	fmt.Printf("start product ledger: %d good, %d bad, %.0f%% SLI\n",
		evaluation.Derived.ProductLedger.GoodLogicalOperations,
		evaluation.Derived.ProductLedger.BadLogicalOperations,
		evaluation.Derived.ProductLedger.SLIPercent)
	fmt.Printf("proposed allocation: browse=%d checkout=%d borrowing=%t\n",
		evaluation.Analysis.Allocation.BrowsePermits,
		evaluation.Analysis.Allocation.CheckoutPermits,
		evaluation.Analysis.Allocation.Borrowing)
	fmt.Printf("corrected replay: %d attempts, %d good, %d bad, max latency %d ms\n",
		evaluation.CorrectedPrediction.CheckoutAttempts,
		evaluation.CorrectedPrediction.GoodLogicalOperations,
		evaluation.CorrectedPrediction.BadLogicalOperations,
		evaluation.CorrectedPrediction.MaxFirstResultLatencyMS)
	fmt.Printf("boundary replay: %d browse admitted, %d waiting, %d checkout permits idle\n",
		evaluation.BoundaryPrediction.IsolatedBrowseAdmitted,
		evaluation.BoundaryPrediction.IsolatedBrowseWaiting,
		evaluation.BoundaryPrediction.IdleCheckoutPermits)
}

func fatalf(format string, arguments ...any) {
	fmt.Fprintf(os.Stderr, "replay: "+format+"\n", arguments...)
	os.Exit(2)
}
