package transfer

import "testing"

const (
	boothA      = "booth-a"
	boothB      = "booth-b"
	right100    = "R-100"
	transfer100 = "X-100"
)

func newTransferPair(t *testing.T) (*Booth, *Booth) {
	t.Helper()

	destination, err := NewBooth(BoothConfig{
		ID:                  boothA,
		TrustedGrantSources: []string{boothB},
	})
	if err != nil {
		t.Fatalf("create booth-a: %v", err)
	}
	source, err := NewBooth(BoothConfig{
		ID:           boothB,
		UsableRights: []string{right100},
	})
	if err != nil {
		t.Fatalf("create booth-b: %v", err)
	}
	return destination, source
}

func transferX100() Transfer {
	return Transfer{
		OperationID:   transfer100,
		RightID:       right100,
		SourceID:      boothB,
		DestinationID: boothA,
	}
}
