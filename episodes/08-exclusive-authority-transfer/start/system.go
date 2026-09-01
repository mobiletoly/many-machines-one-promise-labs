package transfer

import (
	"errors"
	"fmt"
	"sync"
)

var (
	ErrNoAuthority       = errors.New("no local authority")
	ErrOperationConflict = errors.New("operation identity conflict")
	ErrInvalidGrant      = errors.New("invalid grant")
)

type RightStatus string

const (
	RightAbsent       RightStatus = "absent"
	RightUsable       RightStatus = "usable"
	RightConsumed     RightStatus = "consumed"
	RightRelinquished RightStatus = "relinquished"
)

type Transfer struct {
	OperationID   string
	RightID       string
	SourceID      string
	DestinationID string
}

type Grant struct {
	OperationID   string
	RightID       string
	SourceID      string
	DestinationID string
}

type Acceptance struct {
	OperationID   string
	RightID       string
	SourceID      string
	DestinationID string
}

type Sale struct {
	OperationID string
	CustomerID  string
}

type Confirmation struct {
	OperationID string
	CustomerID  string
	RightID     string
	BoothID     string
}

type BoothSnapshot struct {
	BoothID               string
	RetainedTransfers     int
	AcceptedTransfers     int
	RetainedConfirmations int
}

type BoothConfig struct {
	ID                  string
	UsableRights        []string
	TrustedGrantSources []string
}

type transferRecord struct {
	Request Transfer
	Grant   Grant
}

type acceptanceRecord struct {
	Grant      Grant
	Acceptance Acceptance
}

type Booth struct {
	mu             sync.Mutex
	id             string
	rights         map[string]RightStatus
	trustedSources map[string]bool
	transfers      map[string]transferRecord
	acceptances    map[string]acceptanceRecord
	confirmations  map[string]Confirmation
}

func NewBooth(config BoothConfig) (*Booth, error) {
	if config.ID == "" {
		return nil, errors.New("booth id is required")
	}

	rights := make(map[string]RightStatus, len(config.UsableRights))
	for _, rightID := range config.UsableRights {
		if rightID == "" {
			return nil, fmt.Errorf("booth %s: right id is required", config.ID)
		}
		if _, exists := rights[rightID]; exists {
			return nil, fmt.Errorf("booth %s: duplicate right %s", config.ID, rightID)
		}
		rights[rightID] = RightUsable
	}
	trustedSources := make(map[string]bool, len(config.TrustedGrantSources))
	for _, sourceID := range config.TrustedGrantSources {
		if sourceID == "" || sourceID == config.ID || trustedSources[sourceID] {
			return nil, fmt.Errorf("booth %s: invalid trusted grant source %q", config.ID, sourceID)
		}
		trustedSources[sourceID] = true
	}

	return &Booth{
		id:             config.ID,
		rights:         rights,
		trustedSources: trustedSources,
		transfers:      make(map[string]transferRecord),
		acceptances:    make(map[string]acceptanceRecord),
		confirmations:  make(map[string]Confirmation),
	}, nil
}

func (b *Booth) IssueGrant(transfer Transfer) (Grant, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if retained, ok := b.transfers[transfer.OperationID]; ok {
		if retained.Request != transfer {
			return Grant{}, ErrOperationConflict
		}
		return retained.Grant, nil
	}
	if transfer.OperationID == "" || transfer.RightID == "" ||
		transfer.SourceID != b.id || transfer.DestinationID == "" ||
		transfer.DestinationID == b.id {
		return Grant{}, ErrInvalidGrant
	}
	if b.rights[transfer.RightID] != RightUsable {
		return Grant{}, ErrNoAuthority
	}

	grant := Grant{
		OperationID:   transfer.OperationID,
		RightID:       transfer.RightID,
		SourceID:      transfer.SourceID,
		DestinationID: transfer.DestinationID,
	}

	b.transfers[transfer.OperationID] = transferRecord{
		Request: transfer,
		Grant:   grant,
	}
	return grant, nil
}

func (b *Booth) Relinquish(operationID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	record, ok := b.transfers[operationID]
	if !ok {
		return ErrInvalidGrant
	}

	switch b.rights[record.Request.RightID] {
	case RightUsable:
		b.rights[record.Request.RightID] = RightRelinquished
		return nil
	case RightRelinquished:
		return nil
	default:
		return ErrNoAuthority
	}
}

func (b *Booth) AcceptGrant(grant Grant) (Acceptance, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if retained, ok := b.acceptances[grant.OperationID]; ok {
		if retained.Grant != grant {
			return Acceptance{}, ErrOperationConflict
		}
		return retained.Acceptance, nil
	}
	if grant.OperationID == "" || grant.RightID == "" ||
		grant.DestinationID != b.id || !b.trustedSources[grant.SourceID] {
		return Acceptance{}, ErrInvalidGrant
	}
	if _, exists := b.rights[grant.RightID]; exists {
		return Acceptance{}, ErrOperationConflict
	}

	acceptance := Acceptance{
		OperationID:   grant.OperationID,
		RightID:       grant.RightID,
		SourceID:      grant.SourceID,
		DestinationID: grant.DestinationID,
	}
	b.rights[grant.RightID] = RightUsable
	b.acceptances[grant.OperationID] = acceptanceRecord{
		Grant:      grant,
		Acceptance: acceptance,
	}
	return acceptance, nil
}

func (b *Booth) Confirm(rightID string, sale Sale) (Confirmation, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if retained, ok := b.confirmations[sale.OperationID]; ok {
		if retained.CustomerID != sale.CustomerID || retained.RightID != rightID {
			return Confirmation{}, ErrOperationConflict
		}
		return retained, nil
	}
	if sale.OperationID == "" || sale.CustomerID == "" {
		return Confirmation{}, ErrOperationConflict
	}
	if b.rights[rightID] != RightUsable {
		return Confirmation{}, ErrNoAuthority
	}

	b.rights[rightID] = RightConsumed
	confirmation := Confirmation{
		OperationID: sale.OperationID,
		CustomerID:  sale.CustomerID,
		RightID:     rightID,
		BoothID:     b.id,
	}
	b.confirmations[sale.OperationID] = confirmation
	return confirmation, nil
}

func (b *Booth) RightStatus(rightID string) RightStatus {
	b.mu.Lock()
	defer b.mu.Unlock()

	if status, ok := b.rights[rightID]; ok {
		return status
	}
	return RightAbsent
}

func (b *Booth) Snapshot() BoothSnapshot {
	b.mu.Lock()
	defer b.mu.Unlock()

	return BoothSnapshot{
		BoothID:               b.id,
		RetainedTransfers:     len(b.transfers),
		AcceptedTransfers:     len(b.acceptances),
		RetainedConfirmations: len(b.confirmations),
	}
}
