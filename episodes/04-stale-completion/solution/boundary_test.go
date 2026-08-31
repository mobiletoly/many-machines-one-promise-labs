package ownership

import "testing"

func TestBoundaryStaleRejectionLeavesPreparedDrink(t *testing.T) {
	store := newOrder73Store()
	preparation := &PreparationSink{}

	prepared := preparation.Prepare(73, "worker-a", "Cabernet")
	if prepared.WorkerID != "worker-a" || prepared.OrderID != 73 {
		t.Fatalf("prepared drink = %+v, want worker-a work for Order 73", prepared)
	}

	assignment, transferred, err := store.Transfer(73, "worker-a", 17, "worker-b")
	if err != nil {
		t.Fatalf("transfer Order 73 to worker-b: %v", err)
	}
	if !transferred || assignment.Version != 18 {
		t.Fatalf("transfer result = %+v, %v, want worker-b/18", assignment, transferred)
	}

	completion, err := store.Complete(73, "worker-a", 17)
	if err != nil {
		t.Fatalf("complete Order 73 under A/17: %v", err)
	}
	if completion != CompletionStaleAssignment {
		t.Fatalf("completion result = %q, want %q", completion, CompletionStaleAssignment)
	}

	drinks := preparation.Drinks()
	if len(drinks) != 1 || drinks[0] != prepared {
		t.Fatalf("prepared drinks = %+v, want stale worker's Cabernet to remain", drinks)
	}
}
