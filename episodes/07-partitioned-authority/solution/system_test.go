package authority

import (
	"errors"
	"testing"
)

func TestOneBoothConfirmsFromLocalState(t *testing.T) {
	system := newEvent8(t)
	sale := Sale{OperationID: "A-101", CustomerID: "customer-a1"}

	first, err := system.ConfirmAt(boothA, sale)
	if err != nil {
		t.Fatalf("confirm A-101: %v", err)
	}
	second, err := system.ConfirmAt(boothA, sale)
	if err != nil {
		t.Fatalf("retry A-101: %v", err)
	}
	if first != second {
		t.Fatalf("matching retry = %+v, want retained %+v", second, first)
	}

	_, err = system.ConfirmAt(
		boothA,
		Sale{OperationID: "A-101", CustomerID: "different-customer"},
	)
	if !errors.Is(err, ErrOperationConflict) {
		t.Fatalf("conflicting A-101 reuse error = %v, want %v", err, ErrOperationConflict)
	}

	booth, ok := system.Booth(boothA)
	if !ok {
		t.Fatal("booth-a disappeared")
	}
	if booth.Confirmed != 1 || booth.Outstanding != 1 {
		t.Fatalf("booth-a = %+v, want one confirmation and one outstanding permit", booth)
	}
}

func TestAllocatorRejectsAuthorityBeyondCapacity(t *testing.T) {
	_, err := NewAllocatedSystem(event8, 7, 3, []Allocation{
		{BoothID: boothA, Count: 3},
		{BoothID: boothB, Count: 2},
	})
	if err == nil {
		t.Fatal("allocation of five rights from reserve four succeeded")
	}
}
