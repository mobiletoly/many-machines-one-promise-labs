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
	response := create(t, service, "A", "op-61", capstonetest.CreateRequest{
		Wine: "Cabernet", Quantity: 1,
	})
	if orders := database.Orders(t, "op-61"); len(orders) != 1 {
		t.Fatalf("canonical orders = %d, want 1", len(orders))
	}
	if response.Order.OperationID != "op-61" {
		t.Fatalf("operation identity = %q, want op-61", response.Order.OperationID)
	}
}

func TestReplacementReturnsCanonicalResult(t *testing.T) {
	scenario := runMatchingOverlap(t)
	scenario.serviceA.Stop()
	scenario.serviceB.Stop()

	serviceC := capstonetest.StartService(t, scenario.binary, scenario.database, "C", "")
	retry := create(t, serviceC, "C", "op-61", capstonetest.CreateRequest{
		Wine: "Cabernet", Quantity: 1,
	})
	if retry.Order != scenario.responseB.Order {
		t.Fatalf("replacement result = %+v, canonical result = %+v", retry.Order, scenario.responseB.Order)
	}
	if orders := scenario.database.Orders(t, "op-61"); len(orders) != 1 {
		t.Fatalf("canonical orders after replacement retry = %d, want 1", len(orders))
	}
}

func TestConflictingReuseIsRejectedAcrossServers(t *testing.T) {
	database := capstonetest.NewDatabase(t, "schema.sql")
	binary := capstonetest.BuildService(t)
	serviceA := capstonetest.StartService(t, binary, database, "A", "")
	serviceB := capstonetest.StartService(t, binary, database, "B", "")

	create(t, serviceA, "A", "op-61", capstonetest.CreateRequest{
		Wine: "Cabernet", Quantity: 1,
	})
	conflict := capstonetest.CreateOrder(
		&http.Client{Timeout: 5 * time.Second},
		"B",
		serviceB.URL,
		"op-61",
		capstonetest.CreateRequest{Wine: "Merlot", Quantity: 2},
	)
	if conflict.Err != nil {
		t.Fatalf("conflicting reuse: %v", conflict.Err)
	}
	if conflict.Status != http.StatusConflict {
		t.Fatalf("conflicting reuse status = %d, body %q, want %d", conflict.Status, conflict.Body, http.StatusConflict)
	}
	if orders := database.Orders(t, "op-61"); len(orders) != 1 {
		t.Fatalf("canonical orders after conflicting reuse = %d, want 1", len(orders))
	}
}

func TestBoundarySamePayloadDifferentOperationsCreateTwoOrders(t *testing.T) {
	database := capstonetest.NewDatabase(t, "schema.sql")
	binary := capstonetest.BuildService(t)
	serviceA := capstonetest.StartService(t, binary, database, "A", "")
	serviceB := capstonetest.StartService(t, binary, database, "B", "")
	request := capstonetest.CreateRequest{Wine: "Cabernet", Quantity: 1}

	first := create(t, serviceA, "A", "op-61", request)
	second := create(t, serviceB, "B", "op-62", request)
	if first.Order.ID == second.Order.ID {
		t.Fatalf("distinct logical operations returned one order: %+v", first.Order)
	}
	if len(database.Orders(t, "op-61")) != 1 || len(database.Orders(t, "op-62")) != 1 {
		t.Fatal("distinct logical operations did not each retain one canonical order")
	}
}

func create(
	t *testing.T,
	service *capstonetest.Service,
	instanceID string,
	operationID string,
	request capstonetest.CreateRequest,
) capstonetest.CreateResponse {
	t.Helper()

	response := capstonetest.CreateOrder(
		&http.Client{Timeout: 5 * time.Second},
		instanceID,
		service.URL,
		operationID,
		request,
	)
	if response.Err != nil {
		t.Fatalf("instance %s create: %v", instanceID, response.Err)
	}
	if response.Status != http.StatusCreated {
		t.Fatalf(
			"instance %s create status = %d, body = %q, want %d",
			instanceID,
			response.Status,
			response.Body,
			http.StatusCreated,
		)
	}
	return response
}
