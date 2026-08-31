package orders

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestCreateOrderHappyPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "orders.jsonl")
	store, err := OpenStore(path)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	server := httptest.NewServer(NewHandler(store))

	order, status, err := createOrder(server.Client(), server.URL, "op-61", CreateRequest{
		Wine: "Cabernet", Quantity: 1,
	}, false)
	server.Close()
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if status != http.StatusCreated || order.ID != 1 || order.Status != "accepted" {
		t.Fatalf("response = status %d, order %+v", status, order)
	}

	reopened, err := OpenStore(path)
	if err != nil {
		t.Fatalf("reopen store: %v", err)
	}
	orders := reopened.Orders()
	if len(orders) != 1 || orders[0] != order {
		t.Fatalf("reopened orders = %+v, want [%+v]", orders, order)
	}
}

func createOrder(
	client *http.Client,
	baseURL string,
	operationID string,
	create CreateRequest,
	dropResponse bool,
) (Order, int, error) {
	body, err := json.Marshal(create)
	if err != nil {
		return Order{}, 0, fmt.Errorf("encode request: %w", err)
	}
	request, err := http.NewRequest(http.MethodPost, baseURL+"/orders", bytes.NewReader(body))
	if err != nil {
		return Order{}, 0, fmt.Errorf("build request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", operationID)
	if dropResponse {
		request.Header.Set(dropResponseHeader, "true")
	}

	response, err := client.Do(request)
	if err != nil {
		return Order{}, 0, err
	}
	defer response.Body.Close()

	var order Order
	if response.StatusCode == http.StatusCreated {
		if err := json.NewDecoder(response.Body).Decode(&order); err != nil {
			return Order{}, response.StatusCode, fmt.Errorf("decode response: %w", err)
		}
	}
	return order, response.StatusCode, nil
}
