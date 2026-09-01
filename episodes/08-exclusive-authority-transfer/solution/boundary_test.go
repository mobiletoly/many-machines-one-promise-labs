package transfer

import (
	"errors"
	"testing"
)

func TestBoundaryLostGrantLeavesAuthorityGap(t *testing.T) {
	destination, source := newTransferPair(t)

	grant, err := source.IssueGrant(transferX100())
	if err != nil {
		t.Fatalf("issue X-100 before loss: %v", err)
	}
	if grant.RightID != right100 {
		t.Fatalf("grant = %+v, want right R-100", grant)
	}

	_, sourceErr := source.Confirm(
		right100,
		Sale{OperationID: "B-401", CustomerID: "customer-b1"},
	)
	if !errors.Is(sourceErr, ErrNoAuthority) {
		t.Fatalf("booth-b sale during gap error = %v, want %v", sourceErr, ErrNoAuthority)
	}
	_, destinationErr := destination.Confirm(
		right100,
		Sale{OperationID: "A-301", CustomerID: "customer-a1"},
	)
	if !errors.Is(destinationErr, ErrNoAuthority) {
		t.Fatalf("booth-a sale before delivery error = %v, want %v", destinationErr, ErrNoAuthority)
	}

	if source.RightStatus(right100) != RightRelinquished {
		t.Fatalf("booth-b R-100 = %s, want %s", source.RightStatus(right100), RightRelinquished)
	}
	if destination.RightStatus(right100) != RightAbsent {
		t.Fatalf("booth-a R-100 = %s, want %s", destination.RightStatus(right100), RightAbsent)
	}
}

func TestBoundaryDelayedDuplicateDoesNotRestoreConsumedRight(t *testing.T) {
	destination, source := newTransferPair(t)

	grant, err := source.IssueGrant(transferX100())
	if err != nil {
		t.Fatalf("issue X-100: %v", err)
	}
	firstAcceptance, err := destination.AcceptGrant(grant)
	if err != nil {
		t.Fatalf("accept X-100: %v", err)
	}
	if _, err := destination.Confirm(
		right100,
		Sale{OperationID: "A-301", CustomerID: "customer-a1"},
	); err != nil {
		t.Fatalf("consume transferred R-100: %v", err)
	}

	lateAcceptance, err := destination.AcceptGrant(grant)
	if err != nil {
		t.Fatalf("accept delayed duplicate X-100: %v", err)
	}
	if lateAcceptance != firstAcceptance {
		t.Fatalf("late acceptance = %+v, want retained %+v", lateAcceptance, firstAcceptance)
	}
	if destination.RightStatus(right100) != RightConsumed {
		t.Fatalf("booth-a R-100 after duplicate = %s, want %s", destination.RightStatus(right100), RightConsumed)
	}

	_, err = destination.Confirm(
		right100,
		Sale{OperationID: "A-302", CustomerID: "customer-a2"},
	)
	if !errors.Is(err, ErrNoAuthority) {
		t.Fatalf("second sale after delayed duplicate error = %v, want %v", err, ErrNoAuthority)
	}
}
