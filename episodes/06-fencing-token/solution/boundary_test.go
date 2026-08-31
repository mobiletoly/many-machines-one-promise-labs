package fencing

import "testing"

func TestBoundaryIssued18DoesNotFence17BeforeStoreSees18(t *testing.T) {
	controller := newManifest77Controller()
	transferManifest77ToB(t, controller)
	store := newManifest77Store()

	result, err := store.Publish(manifest77, coordinatorA, 17, olderPlan)
	if err != nil {
		t.Fatalf("publish A/17 before S sees 18: %v", err)
	}
	if result != PublicationAccepted {
		t.Fatalf("publish A/17 before S sees 18 = %q, want %q", result, PublicationAccepted)
	}

	manifest := mustManifest77(t, store)
	if manifest.HighestGeneration != 17 || manifest.PublishedGeneration != 17 {
		t.Fatalf("M-77 = %+v, want S to accept A/17 before establishing 18", manifest)
	}
}

func TestBoundaryAnotherStoreMayStillAccept17(t *testing.T) {
	storeS := newManifest77Store()
	storeT := newManifest77Store()

	result, err := storeS.Publish(manifest77, coordinatorB, 18, currentPlan)
	if err != nil || result != PublicationAccepted {
		t.Fatalf("publish B/18 at S = %q, %v, want accepted", result, err)
	}

	resultS, err := storeS.Publish(manifest77, coordinatorA, 17, olderPlan)
	if err != nil {
		t.Fatalf("publish A/17 at S: %v", err)
	}
	resultT, err := storeT.Publish(manifest77, coordinatorA, 17, olderPlan)
	if err != nil {
		t.Fatalf("publish A/17 at T: %v", err)
	}

	if resultS != PublicationStaleGeneration || resultT != PublicationAccepted {
		t.Fatalf(
			"A/17 results = S:%q T:%q, want S:%q T:%q",
			resultS,
			resultT,
			PublicationStaleGeneration,
			PublicationAccepted,
		)
	}
}
