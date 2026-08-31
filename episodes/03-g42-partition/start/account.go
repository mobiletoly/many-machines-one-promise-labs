package g42

import (
	"errors"
	"sort"
	"sync"
)

const FundedValue = 100

var (
	ErrInsufficientValue = errors.New("redemption exceeds locally available funded value")
	ErrOperationConflict = errors.New("operation identity reused with different semantics")
)

type Redemption struct {
	OperationID string `json:"operation_id"`
	Person      string `json:"person"`
	Amount      int    `json:"amount"`
}

type Snapshot struct {
	Funded     int          `json:"funded"`
	Confirmed  int          `json:"confirmed"`
	Operations []Redemption `json:"operations"`
}

type Account struct {
	mu         sync.Mutex
	operations map[string]Redemption
}

func NewAccount() *Account {
	return &Account{operations: make(map[string]Redemption)}
}

func (a *Account) Redeem(redemption Redemption) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if retained, ok := a.operations[redemption.OperationID]; ok {
		if retained != redemption {
			return ErrOperationConflict
		}
		return nil
	}
	if confirmed(a.operations)+redemption.Amount > FundedValue {
		return ErrInsufficientValue
	}
	a.operations[redemption.OperationID] = redemption
	return nil
}

func (a *Account) Merge(remote Snapshot) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if remote.Confirmed <= confirmed(a.operations) {
		return nil
	}

	replacement := make(map[string]Redemption, len(remote.Operations))
	for _, operation := range remote.Operations {
		if retained, ok := replacement[operation.OperationID]; ok && retained != operation {
			return ErrOperationConflict
		}
		replacement[operation.OperationID] = operation
	}
	a.operations = replacement
	return nil
}

func (a *Account) Snapshot() Snapshot {
	a.mu.Lock()
	defer a.mu.Unlock()

	operations := make([]Redemption, 0, len(a.operations))
	for _, operation := range a.operations {
		operations = append(operations, operation)
	}
	sort.Slice(operations, func(i, j int) bool {
		return operations[i].OperationID < operations[j].OperationID
	})

	return Snapshot{
		Funded:     FundedValue,
		Confirmed:  confirmed(a.operations),
		Operations: operations,
	}
}

func confirmed(operations map[string]Redemption) int {
	total := 0
	for _, operation := range operations {
		total += operation.Amount
	}
	return total
}
