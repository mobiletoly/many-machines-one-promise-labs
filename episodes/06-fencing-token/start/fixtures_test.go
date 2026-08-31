package fencing

import "testing"

const (
	manifest77   = "M-77"
	coordinatorA = "coordinator-a"
	coordinatorB = "coordinator-b"
	currentPlan  = "current loading plan"
	olderPlan    = "older loading plan"
)

func newManifest77Controller() *Controller {
	return NewController(map[string]Assignment{
		manifest77: {HolderID: coordinatorA, Generation: 17},
	})
}

func newManifest77Store() *PublicationStore {
	return NewPublicationStore(Manifest{
		ID:                manifest77,
		HighestGeneration: 16,
	})
}

func transferManifest77ToB(t *testing.T, controller *Controller) Assignment {
	t.Helper()

	assignment, transferred, err := controller.Transfer(
		manifest77,
		coordinatorA,
		17,
		coordinatorB,
	)
	if err != nil {
		t.Fatalf("transfer M-77 to coordinator-b: %v", err)
	}
	if !transferred {
		t.Fatal("transfer M-77 to coordinator-b was rejected")
	}
	if assignment != (Assignment{HolderID: coordinatorB, Generation: 18}) {
		t.Fatalf("new assignment = %+v, want coordinator-b/18", assignment)
	}
	current, ok := controller.Current(manifest77)
	if !ok || current != assignment {
		t.Fatalf("current assignment = %+v, %v, want coordinator-b/18", current, ok)
	}
	return assignment
}

func mustManifest77(t *testing.T, store *PublicationStore) Manifest {
	t.Helper()

	manifest, ok := store.Manifest(manifest77)
	if !ok {
		t.Fatal("M-77 disappeared")
	}
	return manifest
}
