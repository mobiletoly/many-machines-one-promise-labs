package lease

import "testing"

const (
	manifest88  = "M-88"
	lease88     = "L-88"
	coordinator = "coordinator-a"
	pBefore     = "P-before"
	pAfter      = "P-after"
	planV1      = "loading plan v1"
	planV2      = "loading plan v2"
)

func newManifest88Desk(
	t *testing.T,
	beforeAcceptance func(PublicationRequest),
) (*Desk, *ManualClock, Lease) {
	t.Helper()
	clock := NewManualClock(100)
	desk, err := NewDesk(DeskConfig{
		Clock:            clock,
		BeforeAcceptance: beforeAcceptance,
	})
	if err != nil {
		t.Fatalf("create publication desk: %v", err)
	}
	lease, err := desk.EstablishLease(LeaseRequest{
		LeaseID:    lease88,
		HolderID:   coordinator,
		ManifestID: manifest88,
		Duration:   10,
	})
	if err != nil {
		t.Fatalf("establish L-88: %v", err)
	}
	return desk, clock, lease
}

func publication(operationID, content string) PublicationRequest {
	return PublicationRequest{
		OperationID: operationID,
		LeaseID:     lease88,
		HolderID:    coordinator,
		ManifestID:  manifest88,
		Content:     content,
	}
}
