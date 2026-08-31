package claim

import "testing"

func TestBoundaryClaimCanStrandOrder(t *testing.T) {
	store := NewStore(Order{ID: 42, Status: Available})

	first, err := store.Claim(42, "worker-a", ClaimOptions{})
	if err != nil {
		t.Fatalf("worker-a claim order 42: %v", err)
	}
	if first != ClaimSucceeded {
		t.Fatalf("worker-a result = %q, want %q", first, ClaimSucceeded)
	}

	// Worker A disappears here. This episode has no completion or transfer rule.
	second, err := store.Claim(42, "worker-b", ClaimOptions{})
	if err != nil {
		t.Fatalf("worker-b claim order 42: %v", err)
	}
	if second != ClaimRejected {
		t.Fatalf("worker-b result = %q, want %q", second, ClaimRejected)
	}

	order, ok := store.Order(42)
	if !ok {
		t.Fatal("order 42 disappeared")
	}
	if order.Status != Claimed || order.ClaimedBy != "worker-a" {
		t.Fatalf("order 42 = %+v, want stranded claim owned by worker-a", order)
	}
}
