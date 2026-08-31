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

type Store struct {
	mu     sync.Mutex
	path   string
	orders []Order
}

func OpenStore(path string) (*Store, error) {
	store := &Store{path: path}

	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return store, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open order log: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var order Order
		if err := json.Unmarshal(scanner.Bytes(), &order); err != nil {
			return nil, fmt.Errorf("decode order log: %w", err)
		}
		store.orders = append(store.orders, order)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read order log: %w", err)
	}

	return store, nil
}

func (s *Store) Create(_ string, request CreateRequest) (Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order := Order{
		ID:       len(s.orders) + 1,
		Wine:     request.Wine,
		Quantity: request.Quantity,
		Status:   "accepted",
	}
	if err := appendAndSync(s.path, order); err != nil {
		return Order{}, err
	}
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
