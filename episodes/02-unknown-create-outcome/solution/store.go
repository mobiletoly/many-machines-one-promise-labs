package orders

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
)

var ErrOperationConflict = errors.New("operation identity reused with different semantics")

type CreateRequest struct {
	Wine     string `json:"wine"`
	Quantity int    `json:"quantity"`
}

type Order struct {
	ID       int    `json:"order_id"`
	Wine     string `json:"wine"`
	Quantity int    `json:"quantity"`
	Status   string `json:"status"`
}

type acceptedOperation struct {
	OperationID string        `json:"operation_id"`
	Request     CreateRequest `json:"request"`
	Result      Order         `json:"result"`
}

type Store struct {
	mu         sync.Mutex
	path       string
	operations map[string]acceptedOperation
	orders     []Order
}

func OpenStore(path string) (*Store, error) {
	store := &Store{
		path:       path,
		operations: make(map[string]acceptedOperation),
	}

	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return store, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open operation log: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var operation acceptedOperation
		if err := json.Unmarshal(scanner.Bytes(), &operation); err != nil {
			return nil, fmt.Errorf("decode operation log: %w", err)
		}
		if _, exists := store.operations[operation.OperationID]; exists {
			return nil, fmt.Errorf("duplicate retained operation %q", operation.OperationID)
		}
		store.operations[operation.OperationID] = operation
		store.orders = append(store.orders, operation.Result)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read operation log: %w", err)
	}

	return store, nil
}

func (s *Store) Create(operationID string, request CreateRequest) (Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if retained, ok := s.operations[operationID]; ok {
		if retained.Request != request {
			return Order{}, ErrOperationConflict
		}
		return retained.Result, nil
	}

	order := Order{
		ID:       len(s.orders) + 1,
		Wine:     request.Wine,
		Quantity: request.Quantity,
		Status:   "accepted",
	}
	operation := acceptedOperation{
		OperationID: operationID,
		Request:     request,
		Result:      order,
	}
	if err := appendAndSync(s.path, operation); err != nil {
		return Order{}, err
	}
	s.operations[operationID] = operation
	s.orders = append(s.orders, order)
	return order, nil
}

func (s *Store) Orders() []Order {
	s.mu.Lock()
	defer s.mu.Unlock()

	return append([]Order(nil), s.orders...)
}

func appendAndSync(path string, value any) error {
	record, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode durable record: %w", err)
	}
	record = append(record, '\n')

	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open durable record: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(record); err != nil {
		return fmt.Errorf("append durable record: %w", err)
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync durable record: %w", err)
	}
	return nil
}
