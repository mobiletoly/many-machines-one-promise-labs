package claim

import (
	"fmt"
	"sync"
)

type Status string

const (
	Available Status = "available"
	Claimed   Status = "claimed"
)

type ClaimResult string

const (
	ClaimSucceeded ClaimResult = "claimed"
	ClaimRejected  ClaimResult = "not_claimed"
)

type Order struct {
	ID        int
	Status    Status
	ClaimedBy string
}

type ClaimOptions struct {
	BeforeCommit func()
}

type Store struct {
	mu     sync.Mutex
	orders map[int]Order
}

func NewStore(orders ...Order) *Store {
	stored := make(map[int]Order, len(orders))
	for _, order := range orders {
		stored[order.ID] = order
	}
	return &Store{orders: stored}
}

func (s *Store) Order(id int) (Order, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.orders[id]
	return order, ok
}

func (s *Store) Claim(id int, workerID string, options ClaimOptions) (ClaimResult, error) {
	if options.BeforeCommit != nil {
		options.BeforeCommit()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.orders[id]
	if !ok {
		return "", fmt.Errorf("order %d: not found", id)
	}
	if order.Status != Available {
		return ClaimRejected, nil
	}

	order.Status = Claimed
	order.ClaimedBy = workerID
	s.orders[id] = order
	return ClaimSucceeded, nil
}
