package authority

import (
	"errors"
	"fmt"
	"sync"
)

var (
	ErrNoAuthority       = errors.New("no local authority")
	ErrOperationConflict = errors.New("operation identity conflict")
)

type Sale struct {
	OperationID string
	CustomerID  string
}

type Confirmation struct {
	OperationID string
	CustomerID  string
	AuthorityID string
}

type Accounting struct {
	Capacity    int
	Confirmed   int
	Outstanding int
	Reserve     int
}

func (a Accounting) Exposure() int {
	return a.Confirmed + a.Outstanding + a.Reserve
}

type BoothSnapshot struct {
	Confirmed         int
	ObservedRemaining int
	Outstanding       int
}

type Booth struct {
	mu                sync.Mutex
	id                string
	observedRemaining int
	confirmations     map[string]Confirmation
}

func newBooth(id string, observedRemaining int) *Booth {
	return &Booth{
		id:                id,
		observedRemaining: observedRemaining,
		confirmations:     make(map[string]Confirmation),
	}
}

func (b *Booth) Confirm(sale Sale) (Confirmation, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if retained, ok := b.confirmations[sale.OperationID]; ok {
		if retained.CustomerID != sale.CustomerID {
			return Confirmation{}, ErrOperationConflict
		}
		return retained, nil
	}
	if b.observedRemaining == 0 {
		return Confirmation{}, ErrNoAuthority
	}

	b.observedRemaining--
	confirmation := Confirmation{
		OperationID: sale.OperationID,
		CustomerID:  sale.CustomerID,
		AuthorityID: "observed-remaining-count",
	}
	b.confirmations[sale.OperationID] = confirmation
	return confirmation, nil
}

func (b *Booth) Snapshot() BoothSnapshot {
	b.mu.Lock()
	defer b.mu.Unlock()

	return BoothSnapshot{
		Confirmed:         len(b.confirmations),
		ObservedRemaining: b.observedRemaining,
		Outstanding:       b.observedRemaining,
	}
}

type System struct {
	eventID         string
	capacity        int
	confirmedBefore int
	reserve         int
	booths          map[string]*Booth
}

func NewObservedSystem(
	eventID string,
	capacity int,
	confirmedBefore int,
	boothIDs ...string,
) (*System, error) {
	if capacity < 0 || confirmedBefore < 0 || confirmedBefore > capacity {
		return nil, fmt.Errorf(
			"event %s: confirmed %d exceeds capacity %d",
			eventID,
			confirmedBefore,
			capacity,
		)
	}

	remaining := capacity - confirmedBefore
	booths := make(map[string]*Booth, len(boothIDs))
	for _, boothID := range boothIDs {
		if _, exists := booths[boothID]; exists {
			return nil, fmt.Errorf("event %s: duplicate booth %s", eventID, boothID)
		}
		booths[boothID] = newBooth(boothID, remaining)
	}

	return &System{
		eventID:         eventID,
		capacity:        capacity,
		confirmedBefore: confirmedBefore,
		reserve:         remaining,
		booths:          booths,
	}, nil
}

func (s *System) ConfirmAt(boothID string, sale Sale) (Confirmation, error) {
	booth, ok := s.booths[boothID]
	if !ok {
		return Confirmation{}, fmt.Errorf("event %s: booth %s not found", s.eventID, boothID)
	}
	return booth.Confirm(sale)
}

func (s *System) Booth(boothID string) (BoothSnapshot, bool) {
	booth, ok := s.booths[boothID]
	if !ok {
		return BoothSnapshot{}, false
	}
	return booth.Snapshot(), true
}

func (s *System) Accounting() Accounting {
	confirmed := s.confirmedBefore
	outstanding := 0
	for _, booth := range s.booths {
		snapshot := booth.Snapshot()
		confirmed += snapshot.Confirmed
		outstanding += snapshot.Outstanding
	}

	return Accounting{
		Capacity:    s.capacity,
		Confirmed:   confirmed,
		Outstanding: outstanding,
		Reserve:     s.reserve,
	}
}
