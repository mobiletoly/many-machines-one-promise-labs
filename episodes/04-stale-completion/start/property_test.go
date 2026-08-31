//go:build failure

package ownership

import "testing"

func TestFormerOwnerCannotCompleteAfterTransfer(t *testing.T) {
	store := newOrder73Store()
	beforeCompletion := make(chan struct{})
	releaseCompletion := make(chan struct{})
	result := make(chan CompleteResult, 1)
	errors := make(chan error, 1)

	go func() {
		close(beforeCompletion)
		<-releaseCompletion
		completion, err := store.Complete(73, "worker-a", 17)
		result <- completion
		errors <- err
	}()

	<-beforeCompletion
	assignment, transferred, err := store.Transfer(73, "worker-a", 17, "worker-b")
	if err != nil {
		t.Fatalf("transfer Order 73 to worker-b: %v", err)
	}
	if !transferred {
		t.Fatal("transfer Order 73 to worker-b was rejected")
	}
	if assignment != (Assignment{WorkerID: "worker-b", Version: 18}) {
		t.Fatalf("new assignment = %+v, want worker-b/18", assignment)
	}

	current := mustOrder73(t, store)
	if current.AssignedTo != "worker-b" || current.AssignmentVersion != 18 {
		t.Fatalf("Order 73 after transfer = %+v, want worker-b/18", current)
	}

	close(releaseCompletion)
	completion := <-result
	if err := <-errors; err != nil {
		t.Fatalf("complete Order 73 under A/17: %v", err)
	}
	if completion != CompletionStaleAssignment {
		t.Fatalf(
			"property violated: completion(order=73, worker=worker-a, assignment=17) = %s, want %s",
			completion,
			CompletionStaleAssignment,
		)
	}

	final := mustOrder73(t, store)
	if final.Status != Claimed || final.AssignedTo != "worker-b" || final.AssignmentVersion != 18 {
		t.Fatalf("Order 73 after stale completion = %+v, want claimed by worker-b/18", final)
	}
}
