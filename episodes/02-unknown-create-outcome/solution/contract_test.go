package orders

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"
)

func TestConflictingReuseIsRejected(t *testing.T) {
	store, err := OpenStore(filepath.Join(t.TempDir(), "orders.jsonl"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	server := httptest.NewServer(NewHandler(store))
	defer server.Close()

	_, firstStatus, err := createOrder(server.Client(), server.URL, "op-61", CreateRequest{
		Wine: "Cabernet", Quantity: 1,
	}, false)
	if err != nil || firstStatus != http.StatusCreated {
		t.Fatalf("first create = status %d, error %v", firstStatus, err)
	}

	_, conflictStatus, err := createOrder(server.Client(), server.URL, "op-61", CreateRequest{
		Wine: "Merlot", Quantity: 2,
	}, false)
	if err != nil {
		t.Fatalf("conflicting create: %v", err)
	}
	if conflictStatus != http.StatusConflict {
		t.Fatalf("conflicting status = %d, want %d", conflictStatus, http.StatusConflict)
	}
	if orders := store.Orders(); len(orders) != 1 {
		t.Fatalf("orders after conflict = %d, want 1", len(orders))
	}
}

func TestConcurrentMatchingAttemptsShareOneResult(t *testing.T) {
	store, err := OpenStore(filepath.Join(t.TempDir(), "orders.jsonl"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	create := CreateRequest{Wine: "Cabernet", Quantity: 1}

	start := make(chan struct{})
	results := make(chan Order, 2)
	errors := make(chan error, 2)
	var ready sync.WaitGroup
	ready.Add(2)
	for range 2 {
		go func() {
			ready.Done()
			<-start
			order, err := store.Create("op-61", create)
			results <- order
			errors <- err
		}()
	}
	ready.Wait()
	close(start)

	first, second := <-results, <-results
	for range 2 {
		if err := <-errors; err != nil {
			t.Fatalf("matching create: %v", err)
		}
	}
	if first != second {
		t.Fatalf("matching results differ: %+v and %+v", first, second)
	}
	if orders := store.Orders(); len(orders) != 1 {
		t.Fatalf("canonical orders = %d, want 1", len(orders))
	}
}

func TestBoundaryOneAcceptedOrderDoesNotMeanOneDispatch(t *testing.T) {
	store, err := OpenStore(filepath.Join(t.TempDir(), "orders.jsonl"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	create := CreateRequest{Wine: "Cabernet", Quantity: 1}
	dispatchAttempts := 0

	for range 2 {
		order, err := store.Create("op-61", create)
		if err != nil {
			t.Fatalf("matching create: %v", err)
		}
		if order.Status == "accepted" {
			dispatchAttempts++
		}
	}

	if orders := store.Orders(); len(orders) != 1 {
		t.Fatalf("canonical orders = %d, want 1", len(orders))
	}
	if dispatchAttempts != 2 {
		t.Fatalf("downstream dispatch attempts = %d, want 2", dispatchAttempts)
	}
}
