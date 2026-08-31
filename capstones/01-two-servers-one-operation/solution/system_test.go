//go:build integration

package orders

import (
	"net/http"
	"testing"
	"time"

	"github.com/mobiletoly/many-machines-one-promise-labs/capstones/01-two-servers-one-operation/internal/capstonetest"
)

type matchingScenario struct {
	database  *capstonetest.Database
	binary    string
	serviceA  *capstonetest.Service
	serviceB  *capstonetest.Service
	responseA capstonetest.CreateResponse
	responseB capstonetest.CreateResponse
}

func runMatchingOverlap(t *testing.T) matchingScenario {
	t.Helper()

	database := capstonetest.NewDatabase(t, "schema.sql")
	binary := capstonetest.BuildService(t)
	gate := capstonetest.NewGate(t, "A", "B")
	serviceA := capstonetest.StartService(t, binary, database, "A", gate.URL())
	serviceB := capstonetest.StartService(t, binary, database, "B", gate.URL())
	client := &http.Client{Timeout: 5 * time.Second}
	create := capstonetest.CreateRequest{Wine: "Cabernet", Quantity: 1}
	responses := make(chan capstonetest.CreateResponse, 2)

	go func() {
		responses <- capstonetest.CreateOrder(client, "A", serviceA.URL, "op-61", create)
	}()
	go func() {
		responses <- capstonetest.CreateOrder(client, "B", serviceB.URL, "op-61", create)
	}()

	gate.WaitFor(t, "A", "B")
	gate.Release(t, "B")
	responseB := receiveResponse(t, responses, "B")
	gate.Release(t, "A")
	responseA := receiveResponse(t, responses, "A")

	return matchingScenario{
		database:  database,
		binary:    binary,
		serviceA:  serviceA,
		serviceB:  serviceB,
		responseA: responseA,
		responseB: responseB,
	}
}

func receiveResponse(
	t *testing.T,
	responses <-chan capstonetest.CreateResponse,
	instanceID string,
) capstonetest.CreateResponse {
	t.Helper()

	select {
	case response := <-responses:
		if response.InstanceID != instanceID {
			t.Fatalf("response came from instance %s, want %s", response.InstanceID, instanceID)
		}
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
	case <-time.After(5 * time.Second):
		t.Fatalf("instance %s did not complete after gate release", instanceID)
		return capstonetest.CreateResponse{}
	}
}
