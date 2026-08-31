package ownership

import "testing"

func TestCurrentOwnerCompletesOrder73(t *testing.T) {
	store := newOrder73Store()

	result, err := store.Complete(73, "worker-a", 17)
	if err != nil {
		t.Fatalf("complete Order 73 under A/17: %v", err)
	}
	if result != CompletionPrepared {
		t.Fatalf("completion result = %q, want %q", result, CompletionPrepared)
	}

	order := mustOrder73(t, store)
	if order.Status != Prepared || order.PreparedBy != "worker-a" {
		t.Fatalf("Order 73 = %+v, want prepared by worker-a", order)
	}
}

func TestTransferRejectsPreparedOrder(t *testing.T) {
	store := newOrder73Store()

	result, err := store.Complete(73, "worker-a", 17)
	if err != nil {
		t.Fatalf("complete Order 73 under A/17: %v", err)
	}
	if result != CompletionPrepared {
		t.Fatalf("completion result = %q, want %q", result, CompletionPrepared)
	}

	_, transferred, err := store.Transfer(73, "worker-a", 17, "worker-b")
	if err != nil {
		t.Fatalf("transfer prepared Order 73: %v", err)
	}
	if transferred {
		t.Fatal("prepared Order 73 transferred, want transfer rejection")
	}
}

func newOrder73Store() *Store {
	return NewStore(Order{
		ID:                73,
		Drink:             "Cabernet",
		Status:            Claimed,
		AssignedTo:        "worker-a",
		AssignmentVersion: 17,
	})
}

func mustOrder73(t *testing.T, store *Store) Order {
	t.Helper()

	order, ok := store.Order(73)
	if !ok {
		t.Fatal("Order 73 disappeared")
	}
	return order
}
