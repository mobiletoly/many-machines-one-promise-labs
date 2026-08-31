//go:build failure

package fencing

import "testing"

func TestStoreRejectsOlderGenerationAfterEstablishing18(t *testing.T) {
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
		t.Fatalf("publish B/18 = %q, want %q", result, PublicationAccepted)
	}

	established := mustManifest77(t, store)
	if established.HighestGeneration != 18 || established.OfficialContent != currentPlan {
		t.Fatalf("M-77 after B/18 = %+v, want current plan at high-water 18", established)
	}

	result, err = store.Publish(manifest77, coordinatorA, 17, olderPlan)
	if err != nil {
		t.Fatalf("publish M-77 under A/17: %v", err)
	}
	if result != PublicationStaleGeneration {
		t.Fatalf(
			"property violated: publish(manifest=M-77, holder=coordinator-a, generation=17) = %s, want %s",
			result,
			PublicationStaleGeneration,
		)
	}

	final := mustManifest77(t, store)
	if final != established {
		t.Fatalf("M-77 after A/17 = %+v, want unchanged %+v", final, established)
	}
}
