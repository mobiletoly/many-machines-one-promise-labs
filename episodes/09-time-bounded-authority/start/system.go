package lease

import (
	"errors"
	"fmt"
	"sync"
)

var (
	ErrInvalidLease       = errors.New("invalid lease")
	ErrLeaseAlreadyExists = errors.New("lease already exists")
)

type Tick int64

type Clock interface {
	Now() Tick
}

type ManualClock struct {
	mu  sync.Mutex
	now Tick
}

func NewManualClock(start Tick) *ManualClock {
	return &ManualClock{now: start}
}

func (c *ManualClock) Now() Tick {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *ManualClock) AdvanceTo(next Tick) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if next < c.now {
		return fmt.Errorf("manual clock cannot move backward from %d to %d", c.now, next)
	}
	c.now = next
	return nil
}

type LeaseRequest struct {
	LeaseID    string
	HolderID   string
	ManifestID string
	Duration   Tick
}

type Lease struct {
	LeaseID    string
	HolderID   string
	ManifestID string
	StartsAt   Tick
	ExpiresAt  Tick
}

type PublicationRequest struct {
	OperationID string
	LeaseID     string
	HolderID    string
	ManifestID  string
	Content     string
}

type PublicationResult string

const (
	PublicationAccepted     PublicationResult = "accepted"
	PublicationLeaseExpired PublicationResult = "lease_expired"
	PublicationInvalidLease PublicationResult = "invalid_lease"
)

type Publication struct {
	OperationID string
	LeaseID     string
	HolderID    string
	ManifestID  string
	Content     string
	DecisionAt  Tick
}

type DeskConfig struct {
	Clock            Clock
	BeforeAcceptance func(PublicationRequest)
}

type Desk struct {
	mu               sync.Mutex
	clock            Clock
	beforeAcceptance func(PublicationRequest)
	leases           map[string]Lease
	accepted         map[string]Publication
	official         map[string]Publication
}

func NewDesk(config DeskConfig) (*Desk, error) {
	if config.Clock == nil {
		return nil, errors.New("desk clock is required")
	}
	return &Desk{
		clock:            config.Clock,
		beforeAcceptance: config.BeforeAcceptance,
		leases:           make(map[string]Lease),
		accepted:         make(map[string]Publication),
		official:         make(map[string]Publication),
	}, nil
}

func (d *Desk) EstablishLease(request LeaseRequest) (Lease, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if request.LeaseID == "" || request.HolderID == "" ||
		request.ManifestID == "" || request.Duration <= 0 {
		return Lease{}, ErrInvalidLease
	}
	if _, exists := d.leases[request.LeaseID]; exists {
		return Lease{}, ErrLeaseAlreadyExists
	}

	startsAt := d.clock.Now()
	lease := Lease{
		LeaseID:    request.LeaseID,
		HolderID:   request.HolderID,
		ManifestID: request.ManifestID,
		StartsAt:   startsAt,
		ExpiresAt:  startsAt + request.Duration,
	}
	d.leases[lease.LeaseID] = lease
	return lease, nil
}

func (d *Desk) Publish(request PublicationRequest) (PublicationResult, error) {
	d.mu.Lock()
	lease, ok := d.leases[request.LeaseID]
	if !ok || !matchesLease(lease, request) {
		d.mu.Unlock()
		return PublicationInvalidLease, nil
	}
	if d.clock.Now() >= lease.ExpiresAt {
		d.mu.Unlock()
		return PublicationLeaseExpired, nil
	}
	d.mu.Unlock()

	if d.beforeAcceptance != nil {
		d.beforeAcceptance(request)
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	publication := Publication{
		OperationID: request.OperationID,
		LeaseID:     request.LeaseID,
		HolderID:    request.HolderID,
		ManifestID:  request.ManifestID,
		Content:     request.Content,
		DecisionAt:  d.clock.Now(),
	}
	d.accepted[publication.OperationID] = publication
	d.official[publication.ManifestID] = publication
	return PublicationAccepted, nil
}

func (d *Desk) Official(manifestID string) (Publication, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	publication, ok := d.official[manifestID]
	return publication, ok
}

func (d *Desk) Accepted(operationID string) (Publication, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	publication, ok := d.accepted[operationID]
	return publication, ok
}

func matchesLease(lease Lease, request PublicationRequest) bool {
	return request.OperationID != "" && request.Content != "" &&
		request.HolderID == lease.HolderID && request.ManifestID == lease.ManifestID
}
