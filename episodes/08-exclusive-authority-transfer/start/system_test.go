package transfer

import (
	"errors"
	"testing"
)

func TestGrantFirstWorkflowCanReachOneFinalHolder(t *testing.T) {
	destination, source := newTransferPair(t)

	grant, err := source.IssueGrant(transferX100())
	if err != nil {
		t.Fatalf("issue X-100 grant: %v", err)
	}
	if _, err := destination.AcceptGrant(grant); err != nil {
		t.Fatalf("accept X-100 grant: %v", err)
	}
	if err := source.Relinquish(transfer100); err != nil {
		t.Fatalf("relinquish R-100: %v", err)
	}

	if source.RightStatus(right100) != RightRelinquished {
		t.Fatalf("booth-b R-100 = %s, want %s", source.RightStatus(right100), RightRelinquished)
	}
	if destination.RightStatus(right100) != RightUsable {
		t.Fatalf("booth-a R-100 = %s, want %s", destination.RightStatus(right100), RightUsable)
	}

	confirmation, err := destination.Confirm(
		right100,
		Sale{OperationID: "A-301", CustomerID: "customer-a1"},
	)
	if err != nil {
		t.Fatalf("confirm A-301 at booth-a: %v", err)
	}
	if confirmation.RightID != right100 || confirmation.BoothID != boothA {
		t.Fatalf("booth-a confirmation = %+v, want R-100 at booth-a", confirmation)
	}

	_, err = source.Confirm(
		right100,
		Sale{OperationID: "B-401", CustomerID: "customer-b1"},
	)
	if !errors.Is(err, ErrNoAuthority) {
		t.Fatalf("booth-b sale after final relinquishment error = %v, want %v", err, ErrNoAuthority)
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
