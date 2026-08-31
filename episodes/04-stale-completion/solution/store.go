package ownership

import (
	"fmt"
	"sync"
)

type Status string

const (
	Claimed  Status = "claimed"
	Prepared Status = "prepared"
)

type CompleteResult string

const (
	CompletionPrepared        CompleteResult = "prepared"
	CompletionStaleAssignment CompleteResult = "stale_assignment"
	CompletionRejected        CompleteResult = "not_completed"
)

type Assignment struct {
	WorkerID string
	Version  int
}

type Order struct {
	ID                int
	Drink             string
	Status            Status
	AssignedTo        string
	AssignmentVersion int
	PreparedBy        string
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

func (s *Store) Transfer(
	id int,
	expectedOwner string,
	expectedVersion int,
	newOwner string,
) (Assignment, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.orders[id]
	if !ok {
		return Assignment{}, false, fmt.Errorf("order %d: not found", id)
	}
	if order.Status != Claimed ||
		order.AssignedTo != expectedOwner ||
		order.AssignmentVersion != expectedVersion {
		return Assignment{}, false, nil
	}

	order.AssignedTo = newOwner
	order.AssignmentVersion++
	s.orders[id] = order

	return Assignment{WorkerID: newOwner, Version: order.AssignmentVersion}, true, nil
}

func (s *Store) Complete(
	id int,
	workerID string,
	assignmentVersion int,
) (CompleteResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.orders[id]
	if !ok {
		return "", fmt.Errorf("order %d: not found", id)
	}
	if order.AssignedTo != workerID || order.AssignmentVersion != assignmentVersion {
		return CompletionStaleAssignment, nil
	}
	if order.Status != Claimed {
		return CompletionRejected, nil
	}

	order.Status = Prepared
	order.PreparedBy = workerID
	s.orders[id] = order

	return CompletionPrepared, nil
}
