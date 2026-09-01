package lease

import "testing"

func TestEstablishedLeaseSupportsPublicationWithoutIssuer(t *testing.T) {
	desk, clock, lease := newManifest88Desk(t, nil)
	if lease.StartsAt != 100 || lease.ExpiresAt != 110 {
		t.Fatalf("L-88 interval = [%d,%d), want [100,110)", lease.StartsAt, lease.ExpiresAt)
	}
	if err := clock.AdvanceTo(106); err != nil {
		t.Fatal(err)
	}

	result, err := desk.Publish(publication(pBefore, planV1))
	if err != nil {
		t.Fatalf("publish P-before without issuer: %v", err)
	}
	if result != PublicationAccepted {
		t.Fatalf("publish P-before = %q, want %q", result, PublicationAccepted)
	}
	official, ok := desk.Official(manifest88)
	if !ok || official.OperationID != pBefore || official.DecisionAt != 106 {
		t.Fatalf("official M-88 = %+v, present=%v; want P-before accepted at 106", official, ok)
	}
}

func TestManualClockRejectsBackwardMovement(t *testing.T) {
	clock := NewManualClock(100)
	if err := clock.AdvanceTo(99); err == nil {
		t.Fatal("manual clock moved backward")
	}
	if clock.Now() != 100 {
		t.Fatalf("clock after rejected movement = %d, want 100", clock.Now())
	}
}

func TestPublicationMustMatchEstablishedLeaseScope(t *testing.T) {
	desk, _, _ := newManifest88Desk(t, nil)
	request := publication(pBefore, planV1)
	request.ManifestID = "M-89"

	result, err := desk.Publish(request)
	if err != nil {
		t.Fatal(err)
	}
	if result != PublicationInvalidLease {
		t.Fatalf("wrong-scope publication = %q, want %q", result, PublicationInvalidLease)
	}
}
