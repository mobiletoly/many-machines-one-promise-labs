//go:build integration

package orders

import (
	"net/http"
	"testing"
	"time"

	"github.com/mobiletoly/many-machines-one-promise-labs/capstones/01-two-servers-one-operation/internal/capstonetest"
)

func TestOneServerCreatesOrder(t *testing.T) {
	database := capstonetest.NewDatabase(t, "schema.sql")
	binary := capstonetest.BuildService(t)
	service := capstonetest.StartService(t, binary, database, "A", "")
	response := capstonetest.CreateOrder(
		&http.Client{Timeout: 5 * time.Second},
		"A",
		service.URL,
		"op-61",
		capstonetest.CreateRequest{Wine: "Cabernet", Quantity: 1},
	)
	if response.Err != nil || response.Status != http.StatusCreated {
		t.Fatalf("create = status %d, body %q, error %v", response.Status, response.Body, response.Err)
	}
	if orders := database.Orders(t, "op-61"); len(orders) != 1 {
		t.Fatalf("canonical orders = %d, want 1", len(orders))
	}
}
