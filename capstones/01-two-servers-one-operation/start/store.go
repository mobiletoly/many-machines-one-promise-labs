package orders

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrOperationConflict = errors.New("operation identity reused with different semantics")

var schemaPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

type CreateRequest struct {
	Wine     string `json:"wine"`
	Quantity int    `json:"quantity"`
}

type Order struct {
	ID          int64  `json:"order_id"`
	OperationID string `json:"operation_id"`
	Wine        string `json:"wine"`
	Quantity    int    `json:"quantity"`
	Status      string `json:"status"`
}

type Store struct {
	mu         sync.Mutex
	pool       *pgxpool.Pool
	orders     string
	gateURL    string
	instanceID string
}

func OpenStore(
	ctx context.Context,
	databaseURL string,
	schema string,
	gateURL string,
	instanceID string,
) (*Store, error) {
	if !schemaPattern.MatchString(schema) {
		return nil, fmt.Errorf("invalid schema %q", schema)
	}
	if gateURL != "" && instanceID == "" {
		return nil, errors.New("LAB_INSTANCE_ID is required when LAB_GATE_URL is set")
	}

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open PostgreSQL pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping PostgreSQL: %w", err)
	}

	return &Store{
		pool:       pool,
		orders:     pgx.Identifier{schema, "orders"}.Sanitize(),
		gateURL:    gateURL,
		instanceID: instanceID,
	}, nil
}

func (s *Store) Close() {
	s.pool.Close()
}

func (s *Store) Create(ctx context.Context, operationID string, request CreateRequest) (Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	retained, err := s.find(ctx, operationID)
	if err == nil {
		return matchingResult(retained, request)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Order{}, err
	}

	if err := s.waitAtGate(ctx); err != nil {
		return Order{}, err
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (operation_id, wine, quantity, status)
		VALUES ($1, $2, $3, 'accepted')
		RETURNING order_id, operation_id, wine, quantity, status
	`, s.orders)

	var created Order
	if err := s.pool.QueryRow(
		ctx, query, operationID, request.Wine, request.Quantity,
	).Scan(
		&created.ID,
		&created.OperationID,
		&created.Wine,
		&created.Quantity,
		&created.Status,
	); err != nil {
		return Order{}, fmt.Errorf("insert canonical order: %w", err)
	}
	return created, nil
}

func (s *Store) find(ctx context.Context, operationID string) (Order, error) {
	query := fmt.Sprintf(`
		SELECT order_id, operation_id, wine, quantity, status
		FROM %s
		WHERE operation_id = $1
		ORDER BY order_id
		LIMIT 1
	`, s.orders)

	var order Order
	if err := s.pool.QueryRow(ctx, query, operationID).Scan(
		&order.ID,
		&order.OperationID,
		&order.Wine,
		&order.Quantity,
		&order.Status,
	); err != nil {
		return Order{}, err
	}
	return order, nil
}

func (s *Store) waitAtGate(ctx context.Context) error {
	if s.gateURL == "" {
		return nil
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, s.gateURL, nil)
	if err != nil {
		return fmt.Errorf("build lab gate request: %w", err)
	}
	request.Header.Set("X-Lab-Instance", s.instanceID)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("wait at lab gate: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("lab gate status = %d, want %d", response.StatusCode, http.StatusNoContent)
	}
	return nil
}

func matchingResult(order Order, request CreateRequest) (Order, error) {
	if order.Wine != request.Wine || order.Quantity != request.Quantity {
		return Order{}, ErrOperationConflict
	}
	return order, nil
}
