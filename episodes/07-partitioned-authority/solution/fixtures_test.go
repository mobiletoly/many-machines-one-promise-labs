package authority

import "testing"

const (
	event8 = "E-8"
	boothA = "booth-a"
	boothB = "booth-b"
)

func newEvent8(t *testing.T) *System {
	t.Helper()

	system, err := NewAllocatedSystem(event8, 7, 3, []Allocation{
		{BoothID: boothA, Count: 2},
		{BoothID: boothB, Count: 2},
	})
	if err != nil {
		t.Fatalf("create E-8 allocated-authority system: %v", err)
	}
	return system
}

func event8Sales() map[string][]Sale {
	return map[string][]Sale{
		boothA: {
			{OperationID: "A-101", CustomerID: "customer-a1"},
			{OperationID: "A-102", CustomerID: "customer-a2"},
			{OperationID: "A-103", CustomerID: "customer-a3"},
		},
		boothB: {
			{OperationID: "B-201", CustomerID: "customer-b1"},
			{OperationID: "B-202", CustomerID: "customer-b2"},
			{OperationID: "B-203", CustomerID: "customer-b3"},
		},
	}
}
