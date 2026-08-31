//go:build integration && failure

package orders

import "testing"

func TestTwoServersCreateOneOperation(t *testing.T) {
	scenario := runMatchingOverlap(t)
	orders := scenario.database.Orders(t, "op-61")
	if len(orders) > 1 {
		t.Fatalf(
			"property violated: canonical_orders(operation=op-61) = %d, want <= 1",
			len(orders),
		)
	}
	if len(orders) == 0 {
		t.Fatal("healthy completion violated: canonical_orders(operation=op-61) = 0, want 1")
	}
	if scenario.responseA.Order != scenario.responseB.Order {
		t.Fatalf(
			"matching results differ: A=%+v B=%+v",
			scenario.responseA.Order,
			scenario.responseB.Order,
		)
	}
}
