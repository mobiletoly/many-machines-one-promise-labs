package g42

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

type replicaProcess struct {
	url string
}

type partitionProxy struct {
	mu          sync.RWMutex
	partitioned bool
	targets     map[string]string
	client      *http.Client
	server      *httptest.Server
}

func TestOneReplicaRedeemsLocally(t *testing.T) {
	binary := buildReplica(t)
	replica := startReplica(t, binary)
	client := &http.Client{Timeout: 2 * time.Second}

	redemption := Redemption{OperationID: "RA-80", Person: "Jessica", Amount: 80}
	first, status := redeem(t, client, replica.url, redemption)
	if status != http.StatusCreated || first.Confirmed != 80 {
		t.Fatalf("first redemption = status %d, snapshot %+v", status, first)
	}
	second, status := redeem(t, client, replica.url, redemption)
	if status != http.StatusCreated || operationIDs(second) != "[RA-80]" {
		t.Fatalf("matching retry = status %d, snapshot %+v", status, second)
	}

	_, status = redeem(t, client, replica.url, Redemption{
		OperationID: "RA-80", Person: "Tom", Amount: 80,
	})
	if status != http.StatusConflict {
		t.Fatalf("conflicting reuse status = %d, want %d", status, http.StatusConflict)
	}
}

func runPartitionScenario(t *testing.T) (Snapshot, Snapshot) {
	t.Helper()

	binary := buildReplica(t)
	replicaA := startReplica(t, binary)
	replicaB := startReplica(t, binary)
	client := &http.Client{Timeout: 2 * time.Second}
	proxy := newPartitionProxy(t, map[string]string{
		"a": replicaA.url,
		"b": replicaB.url,
	})

	if status := syncFrom(t, client, replicaA.url, proxy.urlFor("b")); status != http.StatusBadGateway {
		t.Fatalf("A -> B exchange during partition = %d, want %d", status, http.StatusBadGateway)
	}
	if status := syncFrom(t, client, replicaB.url, proxy.urlFor("a")); status != http.StatusBadGateway {
		t.Fatalf("B -> A exchange during partition = %d, want %d", status, http.StatusBadGateway)
	}

	stateA, status := redeem(t, client, replicaA.url, Redemption{
		OperationID: "RA-80", Person: "Jessica", Amount: 80,
	})
	if status != http.StatusCreated || operationIDs(stateA) != "[RA-80]" {
		t.Fatalf("A local redemption = status %d, snapshot %+v", status, stateA)
	}
	stateB, status := redeem(t, client, replicaB.url, Redemption{
		OperationID: "RB-80", Person: "Tom", Amount: 80,
	})
	if status != http.StatusCreated || operationIDs(stateB) != "[RB-80]" {
		t.Fatalf("B local redemption = status %d, snapshot %+v", status, stateB)
	}

	proxy.heal()
	if status := syncFrom(t, client, replicaA.url, proxy.urlFor("b")); status != http.StatusOK {
		t.Fatalf("A imports B after heal = %d, want %d", status, http.StatusOK)
	}
	if status := syncFrom(t, client, replicaB.url, proxy.urlFor("a")); status != http.StatusOK {
		t.Fatalf("B imports A after heal = %d, want %d", status, http.StatusOK)
	}

	return state(t, client, replicaA.url), state(t, client, replicaB.url)
}

func buildReplica(t *testing.T) string {
	t.Helper()

	binary := filepath.Join(t.TempDir(), "g42-replica")
	command := exec.Command("go", "build", "-o", binary, "./cmd/replica")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("build replica: %v\n%s", err, output)
	}
	return binary
}

func startReplica(t *testing.T, binary string) replicaProcess {
	t.Helper()

	command := exec.Command(binary)
	stdout, err := command.StdoutPipe()
	if err != nil {
		t.Fatalf("open replica stdout: %v", err)
	}
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := command.Start(); err != nil {
		t.Fatalf("start replica: %v", err)
	}

	line, err := bufio.NewReader(stdout).ReadString('\n')
	if err != nil {
		_ = command.Wait()
		t.Fatalf("read replica address: %v\nstderr: %s", err, stderr.String())
	}
	address := strings.TrimSpace(line)
	if !strings.HasPrefix(address, "http://127.0.0.1:") {
		_ = command.Process.Kill()
		_ = command.Wait()
		t.Fatalf("replica address = %q", address)
	}

	t.Cleanup(func() {
		_ = command.Process.Kill()
		_ = command.Wait()
	})
	return replicaProcess{url: address}
}

func newPartitionProxy(t *testing.T, targets map[string]string) *partitionProxy {
	t.Helper()

	proxy := &partitionProxy{
		partitioned: true,
		targets:     targets,
		client:      &http.Client{Timeout: 2 * time.Second},
	}
	proxy.server = httptest.NewServer(http.HandlerFunc(proxy.serveHTTP))
	t.Cleanup(proxy.server.Close)
	return proxy
}

func (p *partitionProxy) urlFor(replica string) string {
	return p.server.URL + "/" + replica
}

func (p *partitionProxy) heal() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.partitioned = false
}

func (p *partitionProxy) serveHTTP(response http.ResponseWriter, request *http.Request) {
	p.mu.RLock()
	partitioned := p.partitioned
	p.mu.RUnlock()
	if partitioned {
		http.Error(response, "controlled partition", http.StatusServiceUnavailable)
		return
	}

	parts := strings.Split(strings.Trim(request.URL.Path, "/"), "/")
	if request.Method != http.MethodGet || len(parts) != 2 || parts[1] != "export" {
		http.NotFound(response, request)
		return
	}
	target, ok := p.targets[parts[0]]
	if !ok {
		http.NotFound(response, request)
		return
	}

	upstream, err := p.client.Get(target + "/export")
	if err != nil {
		http.Error(response, fmt.Sprintf("proxy export: %v", err), http.StatusBadGateway)
		return
	}
	defer upstream.Body.Close()
	response.Header().Set("Content-Type", upstream.Header.Get("Content-Type"))
	response.WriteHeader(upstream.StatusCode)
	_, _ = io.Copy(response, upstream.Body)
}

func redeem(t *testing.T, client *http.Client, replicaURL string, redemption Redemption) (Snapshot, int) {
	t.Helper()
	return postSnapshot(t, client, replicaURL+"/redeem", redemption)
}

func syncFrom(t *testing.T, client *http.Client, replicaURL, peerURL string) int {
	t.Helper()
	_, status := postSnapshot(t, client, replicaURL+"/sync", struct {
		PeerURL string `json:"peer_url"`
	}{PeerURL: peerURL})
	return status
}

func postSnapshot(t *testing.T, client *http.Client, url string, input any) (Snapshot, int) {
	t.Helper()

	body, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("encode request: %v", err)
	}
	response, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, response.Body)
		return Snapshot{}, response.StatusCode
	}

	var snapshot Snapshot
	if err := json.NewDecoder(response.Body).Decode(&snapshot); err != nil {
		t.Fatalf("decode %s response: %v", url, err)
	}
	return snapshot, response.StatusCode
}

func state(t *testing.T, client *http.Client, replicaURL string) Snapshot {
	t.Helper()

	response, err := client.Get(replicaURL + "/state")
	if err != nil {
		t.Fatalf("GET state: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("GET state status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	var snapshot Snapshot
	if err := json.NewDecoder(response.Body).Decode(&snapshot); err != nil {
		t.Fatalf("decode state: %v", err)
	}
	return snapshot
}

func operationIDs(snapshot Snapshot) string {
	identities := make([]string, 0, len(snapshot.Operations))
	for _, operation := range snapshot.Operations {
		identities = append(identities, operation.OperationID)
	}
	sort.Strings(identities)
	return "[" + strings.Join(identities, " ") + "]"
}
