package lease

import "testing"

func TestBoundaryAtExpiryRejectsPublication(t *testing.T) {
	desk, clock, _ := newManifest88Desk(t, nil)
	if err := clock.AdvanceTo(110); err != nil {
		t.Fatal(err)
	}

	result, err := desk.Publish(publication(pAfter, planV2))
	if err != nil {
		t.Fatal(err)
	}
	if result != PublicationLeaseExpired {
		t.Fatalf("publish at expiry = %q, want %q", result, PublicationLeaseExpired)
	}
}

func TestBoundaryAcceptedPublicationSurvivesExpiry(t *testing.T) {
	desk, clock, _ := newManifest88Desk(t, nil)
	if err := clock.AdvanceTo(109); err != nil {
		t.Fatal(err)
	}
	result, err := desk.Publish(publication(pBefore, planV1))
	if err != nil {
		t.Fatal(err)
	}
	if result != PublicationAccepted {
		t.Fatalf("publish before expiry = %q, want %q", result, PublicationAccepted)
	}

	if err := clock.AdvanceTo(111); err != nil {
		t.Fatal(err)
	}
	official, ok := desk.Official(manifest88)
	if !ok || official.OperationID != pBefore || official.DecisionAt != 109 {
		t.Fatalf("official M-88 after expiry = %+v, present=%v; want retained P-before at 109", official, ok)
	}
	accepted, ok := desk.Accepted(pBefore)
	if !ok || accepted != official {
		t.Fatalf("accepted P-before after expiry = %+v, present=%v; want retained official publication", accepted, ok)
	}
}
