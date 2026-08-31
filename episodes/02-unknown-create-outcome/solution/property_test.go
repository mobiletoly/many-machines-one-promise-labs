//go:build failure

package orders

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestRetryAfterCommittedResponseIsLost(t *testing.T) {
	path := filepath.Join(t.TempDir(), "orders.jsonl")
	create := CreateRequest{Wine: "Cabernet", Quantity: 1}

	firstStore, err := OpenStore(path)
	if err != nil {
		t.Fatalf("open first store: %v", err)
	}
	firstServer := httptest.NewServer(NewHandler(firstStore))
	_, _, firstErr := createOrder(firstServer.Client(), firstServer.URL, "op-61", create, true)
	firstServer.Close()
	if firstErr == nil {
		t.Fatal("fault seam returned a response, want transport error after commit")
	}
	firstOrders := firstStore.Orders()
	if len(firstOrders) != 1 {
		t.Fatalf("post-fault durable orders = %d, want 1", len(firstOrders))
	}

	replacementStore, err := OpenStore(path)
	if err != nil {
		t.Fatalf("open replacement store: %v", err)
	}
	replacementServer := httptest.NewServer(NewHandler(replacementStore))
	retryResult, retryStatus, err := createOrder(
		replacementServer.Client(), replacementServer.URL, "op-61", create, false,
	)
	replacementServer.Close()
	if err != nil {
		t.Fatalf("retry op-61: %v", err)
	}
	if retryStatus != http.StatusCreated {
		t.Fatalf("retry status = %d, want %d", retryStatus, http.StatusCreated)
	}

	orders := replacementStore.Orders()
	if len(orders) > 1 {
		t.Fatalf(
			"property violated: canonical_orders(operation=op-61) = %d, want <= 1",
			len(orders),
		)
	}
	if len(orders) == 0 {
		t.Fatal("healthy completion violated: canonical_orders(operation=op-61) = 0, want 1")
	}
	if retryResult != firstOrders[0] {
		t.Fatalf("retry result = %+v, retained first result = %+v", retryResult, firstOrders[0])
	}
}
