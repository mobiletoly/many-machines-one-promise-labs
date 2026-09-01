package transfer

import (
	"errors"
	"sync"
	"testing"
)

func TestRelinquishFirstWorkflowReachesOneFinalHolder(t *testing.T) {
	destination, source := newTransferPair(t)

	grant, err := source.IssueGrant(transferX100())
	if err != nil {
		t.Fatalf("issue X-100 grant: %v", err)
	}
	if source.RightStatus(right100) != RightRelinquished {
		t.Fatalf("booth-b R-100 after grant issue = %s, want %s", source.RightStatus(right100), RightRelinquished)
	}
	if _, err := destination.AcceptGrant(grant); err != nil {
		t.Fatalf("accept X-100 grant: %v", err)
	}
	if err := source.Relinquish(transfer100); err != nil {
		t.Fatalf("repeat retained relinquishment: %v", err)
	}

	if destination.RightStatus(right100) != RightUsable {
		t.Fatalf("booth-a R-100 = %s, want %s", destination.RightStatus(right100), RightUsable)
	}
}

func TestMatchingGrantRetryRetainsOneTransfer(t *testing.T) {
	_, source := newTransferPair(t)
	transfer := transferX100()

	first, err := source.IssueGrant(transfer)
	if err != nil {
		t.Fatalf("issue X-100: %v", err)
	}
	second, err := source.IssueGrant(transfer)
	if err != nil {
		t.Fatalf("retry X-100: %v", err)
	}
	if first != second {
		t.Fatalf("matching X-100 retry = %+v, want retained %+v", second, first)
	}

	conflict := transfer
	conflict.DestinationID = "booth-c"
	if _, err := source.IssueGrant(conflict); !errors.Is(err, ErrOperationConflict) {
		t.Fatalf("conflicting X-100 reuse error = %v, want %v", err, ErrOperationConflict)
	}
	if snapshot := source.Snapshot(); snapshot.RetainedTransfers != 1 {
		t.Fatalf("booth-b retained transfers = %d, want 1", snapshot.RetainedTransfers)
	}
}

func TestDestinationRejectsGrantOutsideConfiguredSource(t *testing.T) {
	destination, _ := newTransferPair(t)
	grant := Grant{
		OperationID:   transfer100,
		RightID:       right100,
		SourceID:      "booth-c",
		DestinationID: boothA,
	}

	if _, err := destination.AcceptGrant(grant); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("untrusted source error = %v, want %v", err, ErrInvalidGrant)
	}
	if destination.RightStatus(right100) != RightAbsent {
		t.Fatalf("booth-a accepted R-100 from untrusted source")
	}
}

func TestDestinationRejectsConflictingTransferReuse(t *testing.T) {
	destination, source := newTransferPair(t)
	grant, err := source.IssueGrant(transferX100())
	if err != nil {
		t.Fatalf("issue X-100: %v", err)
	}
	if _, err := destination.AcceptGrant(grant); err != nil {
		t.Fatalf("accept X-100: %v", err)
	}

	conflict := grant
	conflict.RightID = "R-101"
	if _, err := destination.AcceptGrant(conflict); !errors.Is(err, ErrOperationConflict) {
		t.Fatalf("conflicting destination X-100 reuse error = %v, want %v", err, ErrOperationConflict)
	}
	if snapshot := destination.Snapshot(); snapshot.AcceptedTransfers != 1 {
		t.Fatalf("booth-a accepted transfers = %d, want 1", snapshot.AcceptedTransfers)
	}
}

func TestSourceSaleAndGrantIssuanceShareOneDecision(t *testing.T) {
	for run := 0; run < 100; run++ {
		_, source := newTransferPair(t)
		start := make(chan struct{})
		results := make(chan error, 2)
		var group sync.WaitGroup
		group.Add(2)

		go func() {
			defer group.Done()
			<-start
			_, err := source.IssueGrant(transferX100())
			results <- err
		}()
		go func() {
			defer group.Done()
			<-start
			_, err := source.Confirm(
				right100,
				Sale{OperationID: "B-401", CustomerID: "customer-b1"},
			)
			results <- err
		}()

		close(start)
		group.Wait()
		close(results)

		successes := 0
		noAuthority := 0
		for err := range results {
			switch {
			case err == nil:
				successes++
			case errors.Is(err, ErrNoAuthority):
				noAuthority++
			default:
				t.Fatalf("run %d: competing transition error = %v", run, err)
			}
		}
		if successes != 1 || noAuthority != 1 {
			t.Fatalf(
				"run %d: successes = %d, no-authority results = %d; want one of each",
				run,
				successes,
				noAuthority,
			)
		}
	}
}
