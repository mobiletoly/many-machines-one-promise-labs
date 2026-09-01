//go:build failure

package transfer

import (
	"errors"
	"testing"
)

func TestExclusiveAuthoritySurvivesGrantDelivery(t *testing.T) {
	destination, source := newTransferPair(t)
	transfer := transferX100()

	firstAttempt, err := source.IssueGrant(transfer)
	if err != nil {
		t.Fatalf("issue first X-100 grant attempt: %v", err)
	}
	// The first attempt is lost. The source must retain enough state to retry
	// the same logical transfer without restoring or creating authority.
	retryAttempt, err := source.IssueGrant(transfer)
	if err != nil {
		t.Fatalf("retry X-100 grant: %v", err)
	}
	if retryAttempt != firstAttempt {
		t.Fatalf("retry grant = %+v, want retained %+v", retryAttempt, firstAttempt)
	}

	accepted, err := destination.AcceptGrant(retryAttempt)
	if err != nil {
		t.Fatalf("deliver X-100 grant to healthy booth-a: %v", err)
	}
	replayed, err := destination.AcceptGrant(retryAttempt)
	if err != nil {
		t.Fatalf("redeliver X-100 grant: %v", err)
	}
	if replayed != accepted {
		t.Fatalf("duplicate acceptance = %+v, want retained %+v", replayed, accepted)
	}

	destinationConfirmation, err := destination.Confirm(
		right100,
		Sale{OperationID: "A-301", CustomerID: "customer-a1"},
	)
	if err != nil {
		t.Fatalf("healthy destination progress violated: confirm A-301 at booth-a: %v", err)
	}
	if destinationConfirmation.RightID != right100 {
		t.Fatalf("destination confirmation = %+v, want right R-100", destinationConfirmation)
	}

	_, sourceErr := source.Confirm(
		right100,
		Sale{OperationID: "B-401", CustomerID: "customer-b1"},
	)
	if sourceErr == nil {
		t.Fatalf(
			"exclusive authority violated: right R-100 confirmed by A-301 at booth-a and B-401 at booth-b during X-100",
		)
	}
	if !errors.Is(sourceErr, ErrNoAuthority) {
		t.Fatalf("source sale error = %v, want %v", sourceErr, ErrNoAuthority)
	}

	if snapshot := source.Snapshot(); snapshot.RetainedTransfers != 1 {
		t.Fatalf("booth-b retained transfers = %d, want 1", snapshot.RetainedTransfers)
	}
	if snapshot := destination.Snapshot(); snapshot.AcceptedTransfers != 1 || snapshot.RetainedConfirmations != 1 {
		t.Fatalf("booth-a retained state = %+v, want one acceptance and one confirmation", snapshot)
	}
}
