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

type Allocation struct {
	BoothID string
	Count   int
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
	mu            sync.Mutex
	id            string
	permitOrder   []string
	usablePermits map[string]bool
	confirmations map[string]Confirmation
}

func newBooth(id string, permits []string) *Booth {
	usable := make(map[string]bool, len(permits))
	for _, permitID := range permits {
		usable[permitID] = true
	}
	return &Booth{
		id:            id,
		permitOrder:   append([]string(nil), permits...),
		usablePermits: usable,
		confirmations: make(map[string]Confirmation),
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

	permitID := ""
	for _, candidate := range b.permitOrder {
		if b.usablePermits[candidate] {
			permitID = candidate
			break
		}
	}
	if permitID == "" {
		return Confirmation{}, ErrNoAuthority
	}

	b.usablePermits[permitID] = false
	confirmation := Confirmation{
		OperationID: sale.OperationID,
		CustomerID:  sale.CustomerID,
		AuthorityID: permitID,
	}
	b.confirmations[sale.OperationID] = confirmation
	return confirmation, nil
}

func (b *Booth) Snapshot() BoothSnapshot {
	b.mu.Lock()
	defer b.mu.Unlock()

	outstanding := 0
	for _, usable := range b.usablePermits {
		if usable {
			outstanding++
		}
	}
	return BoothSnapshot{
		Confirmed:   len(b.confirmations),
		Outstanding: outstanding,
	}
}

type System struct {
	eventID         string
	capacity        int
	confirmedBefore int
	reserve         int
	booths          map[string]*Booth
}

func NewAllocatedSystem(
	eventID string,
	capacity int,
	confirmedBefore int,
	allocations []Allocation,
) (*System, error) {
	if capacity < 0 || confirmedBefore < 0 || confirmedBefore > capacity {
		return nil, fmt.Errorf(
			"event %s: confirmed %d exceeds capacity %d",
			eventID,
			confirmedBefore,
			capacity,
		)
	}

	reserve := capacity - confirmedBefore
	booths := make(map[string]*Booth, len(allocations))
	nextPermit := 1
	for _, allocation := range allocations {
		if allocation.Count < 0 {
			return nil, fmt.Errorf(
				"event %s: booth %s requested a negative allocation",
				eventID,
				allocation.BoothID,
			)
		}
		if _, exists := booths[allocation.BoothID]; exists {
			return nil, fmt.Errorf(
				"event %s: duplicate booth %s",
				eventID,
				allocation.BoothID,
			)
		}
		if allocation.Count > reserve {
			return nil, fmt.Errorf(
				"event %s: allocation of %d to %s exceeds reserve %d",
				eventID,
				allocation.Count,
				allocation.BoothID,
				reserve,
			)
		}

		permits := make([]string, allocation.Count)
		for i := range allocation.Count {
			permits[i] = fmt.Sprintf("%s-P-%03d", eventID, nextPermit)
			nextPermit++
		}
		reserve -= allocation.Count
		booths[allocation.BoothID] = newBooth(allocation.BoothID, permits)
	}

	return &System{
		eventID:         eventID,
		capacity:        capacity,
		confirmedBefore: confirmedBefore,
		reserve:         reserve,
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
