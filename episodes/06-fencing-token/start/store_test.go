package fencing

import "testing"

func TestCurrentHolderPublishesManifest77(t *testing.T) {
	controller := newManifest77Controller()
	assignment := transferManifest77ToB(t, controller)
	store := newManifest77Store()

	result, err := store.Publish(
		manifest77,
		assignment.HolderID,
		assignment.Generation,
		currentPlan,
	)
	if err != nil {
		t.Fatalf("publish M-77 under B/18: %v", err)
	}
	if result != PublicationAccepted {
		t.Fatalf("publish result = %q, want %q", result, PublicationAccepted)
	}

	manifest := mustManifest77(t, store)
	if manifest.HighestGeneration != 18 ||
		manifest.PublishedGeneration != 18 ||
		manifest.PublishedBy != coordinatorB ||
		manifest.OfficialContent != currentPlan {
		t.Fatalf("M-77 = %+v, want current plan from coordinator-b/18", manifest)
	}
}
