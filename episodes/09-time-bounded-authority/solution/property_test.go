//go:build failure

package lease

import "testing"

func TestLeaseExpiryIsEnforcedAtEffectAcceptance(t *testing.T) {
	var clock *ManualClock
	desk, deskClock, lease := newManifest88Desk(t, func(request PublicationRequest) {
		if request.OperationID != pAfter {
			return
		}
		if err := clock.AdvanceTo(110); err != nil {
			t.Fatalf("advance S clock at acceptance boundary: %v", err)
		}
	})
	clock = deskClock
	if lease.ExpiresAt != 110 {
		t.Fatalf("L-88 expiry = %d, want 110", lease.ExpiresAt)
	}

	if err := clock.AdvanceTo(106); err != nil {
		t.Fatal(err)
	}
	result, err := desk.Publish(publication(pBefore, planV1))
	if err != nil {
		t.Fatalf("publish P-before: %v", err)
	}
	if result != PublicationAccepted {
		t.Fatalf("pre-expiry progress violated: publish P-before = %q, want %q", result, PublicationAccepted)
	}

	if err := clock.AdvanceTo(109); err != nil {
		t.Fatal(err)
	}
	result, err = desk.Publish(publication(pAfter, planV2))
	if err != nil {
		t.Fatalf("publish P-after: %v", err)
	}
	if result != PublicationLeaseExpired {
		t.Fatalf(
			"lease expiry violated: publish(operation=P-after, lease=L-88) decided at S time %d = %s, want %s",
			clock.Now(),
			result,
			PublicationLeaseExpired,
		)
	}

	official, ok := desk.Official(manifest88)
	if !ok || official.OperationID != pBefore || official.Content != planV1 {
		t.Fatalf("official M-88 after rejected P-after = %+v, present=%v; want retained P-before", official, ok)
	}
	if _, accepted := desk.Accepted(pAfter); accepted {
		t.Fatal("P-after appears in accepted history after expiry")
	}
}
