package capstonetest

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CreateRequest struct {
	Wine     string `json:"wine"`
	Quantity int    `json:"quantity"`
}

type Order struct {
	ID          int64  `db:"order_id" json:"order_id"`
	OperationID string `db:"operation_id" json:"operation_id"`
	Wine        string `db:"wine" json:"wine"`
	Quantity    int    `db:"quantity" json:"quantity"`
	Status      string `db:"status" json:"status"`
}

type CreateResponse struct {
	InstanceID string
	Order      Order
	Status     int
	Body       string
	Err        error
}

type Database struct {
	URL    string
	Schema string
	pool   *pgxpool.Pool
	orders string
}

func NewDatabase(t *testing.T, schemaPath string) *Database {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("DATABASE_URL is required; run the capstone verifier for managed Docker Compose bootstrap")
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("open PostgreSQL pool: %v", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		t.Fatalf("ping PostgreSQL: %v", err)
	}

	suffix := make([]byte, 8)
	if _, err := rand.Read(suffix); err != nil {
		pool.Close()
		t.Fatalf("create schema suffix: %v", err)
	}
	schema := "mmop_capstone_" + hex.EncodeToString(suffix)
	identifier := pgx.Identifier{schema}.Sanitize()
	if _, err := pool.Exec(context.Background(), "CREATE SCHEMA "+identifier); err != nil {
		pool.Close()
		t.Fatalf("create schema %s: %v", schema, err)
	}

	database := &Database{
		URL:    databaseURL,
		Schema: schema,
		pool:   pool,
		orders: pgx.Identifier{schema, "orders"}.Sanitize(),
	}
	t.Cleanup(func() {
		_, cleanupErr := pool.Exec(context.Background(), "DROP SCHEMA "+identifier+" CASCADE")
		pool.Close()
		if cleanupErr != nil {
			t.Errorf("drop owned schema %s: %v", schema, cleanupErr)
		}
	})

	schemaSQL, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("read schema SQL: %v", err)
	}
	connection, err := pool.Acquire(context.Background())
	if err != nil {
		t.Fatalf("acquire schema connection: %v", err)
	}
	defer connection.Release()
	if _, err := connection.Exec(context.Background(), "SET search_path TO "+identifier); err != nil {
		t.Fatalf("set schema search path: %v", err)
	}
	if _, err := connection.Exec(context.Background(), string(schemaSQL)); err != nil {
		t.Fatalf("apply schema SQL: %v", err)
	}

	return database
}

func (d *Database) Orders(t *testing.T, operationID string) []Order {
	t.Helper()

	query := fmt.Sprintf(`
		SELECT order_id, operation_id, wine, quantity, status
		FROM %s
		WHERE operation_id = $1
		ORDER BY order_id
	`, d.orders)
	rows, err := d.pool.Query(context.Background(), query, operationID)
	if err != nil {
		t.Fatalf("query canonical orders: %v", err)
	}
	defer rows.Close()

	orders, err := pgx.CollectRows(rows, pgx.RowToStructByName[Order])
	if err != nil {
		t.Fatalf("collect canonical orders: %v", err)
	}
	return orders
}

func BuildService(t *testing.T) string {
	t.Helper()

	binary := filepath.Join(t.TempDir(), "ordersvc")
	command := exec.Command("go", "build", "-o", binary, "./cmd/ordersvc")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("build ordersvc: %v\n%s", err, output)
	}
	return binary
}

type Service struct {
	URL      string
	command  *exec.Cmd
	stderr   *bytes.Buffer
	stopOnce sync.Once
}

func StartService(
	t *testing.T,
	binary string,
	database *Database,
	instanceID string,
	gateURL string,
) *Service {
	t.Helper()

	command := exec.Command(binary)
	command.Env = append(os.Environ(),
		"DATABASE_URL="+database.URL,
		"LAB_SCHEMA="+database.Schema,
		"LAB_INSTANCE_ID="+instanceID,
		"LAB_GATE_URL="+gateURL,
	)
	stdout, err := command.StdoutPipe()
	if err != nil {
		t.Fatalf("open ordersvc stdout: %v", err)
	}
	stderr := &bytes.Buffer{}
	command.Stderr = stderr
	if err := command.Start(); err != nil {
		t.Fatalf("start ordersvc %s: %v", instanceID, err)
	}

	line, err := bufio.NewReader(stdout).ReadString('\n')
	if err != nil {
		_ = command.Process.Kill()
		_ = command.Wait()
		t.Fatalf("read ordersvc %s address: %v\nstderr: %s", instanceID, err, stderr.String())
	}
	address := strings.TrimSpace(line)
	if !strings.HasPrefix(address, "http://127.0.0.1:") {
		_ = command.Process.Kill()
		_ = command.Wait()
		t.Fatalf("ordersvc %s address = %q", instanceID, address)
	}

	service := &Service{
		URL:     address,
		command: command,
		stderr:  stderr,
	}
	t.Cleanup(service.Stop)
	return service
}

func (s *Service) Stop() {
	s.stopOnce.Do(func() {
		_ = s.command.Process.Kill()
		_ = s.command.Wait()
	})
}

func CreateOrder(
	client *http.Client,
	instanceID string,
	serviceURL string,
	operationID string,
	create CreateRequest,
) CreateResponse {
	body, err := json.Marshal(create)
	if err != nil {
		return CreateResponse{InstanceID: instanceID, Err: fmt.Errorf("encode create request: %w", err)}
	}
	request, err := http.NewRequest(
		http.MethodPost,
		serviceURL+"/orders",
		bytes.NewReader(body),
	)
	if err != nil {
		return CreateResponse{InstanceID: instanceID, Err: fmt.Errorf("build create request: %w", err)}
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", operationID)

	response, err := client.Do(request)
	if err != nil {
		return CreateResponse{InstanceID: instanceID, Err: fmt.Errorf("create order: %w", err)}
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return CreateResponse{InstanceID: instanceID, Err: fmt.Errorf("read create response: %w", err)}
	}

	result := CreateResponse{
		InstanceID: instanceID,
		Status:     response.StatusCode,
		Body:       strings.TrimSpace(string(responseBody)),
	}
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		if err := json.Unmarshal(responseBody, &result.Order); err != nil {
			result.Err = fmt.Errorf("decode create response: %w", err)
		}
	}
	return result
}

type Gate struct {
	server   *httptest.Server
	arrived  chan string
	releases map[string]chan struct{}
	mu       sync.Mutex
}

func NewGate(t *testing.T, instanceIDs ...string) *Gate {
	t.Helper()

	gate := &Gate{
		arrived:  make(chan string, len(instanceIDs)),
		releases: make(map[string]chan struct{}, len(instanceIDs)),
	}
	for _, instanceID := range instanceIDs {
		gate.releases[instanceID] = make(chan struct{})
	}
	gate.server = httptest.NewServer(http.HandlerFunc(gate.serveHTTP))
	t.Cleanup(gate.server.Close)
	return gate
}

func (g *Gate) URL() string {
	return g.server.URL
}

func (g *Gate) WaitFor(t *testing.T, wanted ...string) {
	t.Helper()

	arrived := make([]string, 0, len(wanted))
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	for len(arrived) < len(wanted) {
		select {
		case instanceID := <-g.arrived:
			arrived = append(arrived, instanceID)
		case <-deadline.C:
			t.Fatalf("lab gate arrivals = %v, want %v", arrived, wanted)
		}
	}
	sort.Strings(arrived)
	sort.Strings(wanted)
	if strings.Join(arrived, ",") != strings.Join(wanted, ",") {
		t.Fatalf("lab gate arrivals = %v, want %v", arrived, wanted)
	}
}

func (g *Gate) Release(t *testing.T, instanceID string) {
	t.Helper()

	g.mu.Lock()
	defer g.mu.Unlock()
	release, ok := g.releases[instanceID]
	if !ok {
		t.Fatalf("unknown lab gate instance %q", instanceID)
	}
	select {
	case <-release:
		t.Fatalf("lab gate instance %q released twice", instanceID)
	default:
		close(release)
	}
}

func (g *Gate) serveHTTP(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost || request.URL.Path != "/" {
		http.NotFound(response, request)
		return
	}
	instanceID := request.Header.Get("X-Lab-Instance")
	g.mu.Lock()
	release, ok := g.releases[instanceID]
	g.mu.Unlock()
	if !ok {
		http.Error(response, "unknown lab instance", http.StatusBadRequest)
		return
	}
	g.arrived <- instanceID
	<-release
	response.WriteHeader(http.StatusNoContent)
}
