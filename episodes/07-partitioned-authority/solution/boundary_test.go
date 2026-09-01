package authority

import (
	"errors"
	"testing"
)

func TestBoundarySafeAllocationCanStrandCapacity(t *testing.T) {
	system, err := NewAllocatedSystem(event8, 7, 3, []Allocation{
		{BoothID: boothA, Count: 1},
		{BoothID: boothB, Count: 3},
	})
	if err != nil {
		t.Fatalf("create stranded-capacity system: %v", err)
	}

	_, err = system.ConfirmAt(
		boothA,
		Sale{OperationID: "A-101", CustomerID: "customer-a1"},
	)
	if err != nil {
		t.Fatalf("confirm A-101: %v", err)
	}

	_, err = system.ConfirmAt(
		boothA,
		Sale{OperationID: "A-102", CustomerID: "customer-a2"},
	)
	if !errors.Is(err, ErrNoAuthority) {
		t.Fatalf("second booth-a sale error = %v, want %v", err, ErrNoAuthority)
	}

	boothBState, ok := system.Booth(boothB)
	if !ok {
		t.Fatal("booth-b disappeared")
	}
	if boothBState.Outstanding != 3 {
		t.Fatalf("booth-b outstanding rights = %d, want 3", boothBState.Outstanding)
	}

	accounting := system.Accounting()
	if accounting.Exposure() != accounting.Capacity {
		t.Fatalf(
			"safe accounting = confirmed %d + outstanding %d + reserve %d = %d, want capacity %d",
			accounting.Confirmed,
			accounting.Outstanding,
			accounting.Reserve,
			accounting.Exposure(),
			accounting.Capacity,
		)
	}
}
