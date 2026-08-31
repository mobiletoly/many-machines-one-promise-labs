package claim

import "testing"

func TestOneWorkerClaimsOrder(t *testing.T) {
	store := NewStore(Order{ID: 42, Status: Available})

	result, err := store.Claim(42, "worker-a", ClaimOptions{})
	if err != nil {
		t.Fatalf("claim order 42: %v", err)
	}
	if result != ClaimSucceeded {
		t.Fatalf("claim result = %q, want %q", result, ClaimSucceeded)
	}

	order, ok := store.Order(42)
	if !ok {
		t.Fatal("order 42 disappeared")
	}
	if order.Status != Claimed || order.ClaimedBy != "worker-a" {
		t.Fatalf("order 42 = %+v, want claimed by worker-a", order)
	}
}
