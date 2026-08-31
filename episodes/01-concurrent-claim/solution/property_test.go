//go:build failure

package claim

import (
	"testing"
)

func TestTwoWorkersClaimOrder42(t *testing.T) {
	store := NewStore(Order{ID: 42, Status: Available})
	gate := newCommitGate(2)
	results := make(chan workerResult, 2)

	for _, workerID := range []string{"worker-a", "worker-b"} {
		workerID := workerID
		go func() {
			result, err := store.Claim(42, workerID, ClaimOptions{
				BeforeCommit: gate.Wait,
			})
			results <- workerResult{workerID: workerID, result: result, err: err}
		}()
	}

	gate.ReleaseWhenAllArrive()

	claimedBy := ""
	successfulClaims := 0
	for range 2 {
		result := <-results
		if result.err != nil {
			t.Fatalf("%s claim order 42: %v", result.workerID, result.err)
		}
		if result.result == ClaimSucceeded {
			successfulClaims++
			claimedBy = result.workerID
		}
	}

	if successfulClaims > 1 {
		t.Fatalf(
			"property violated: successful_claims(order=42) = %d, want <= 1",
			successfulClaims,
		)
	}
	if successfulClaims == 0 {
		t.Fatal("healthy completion violated: successful_claims(order=42) = 0, want 1")
	}

	order, ok := store.Order(42)
	if !ok {
		t.Fatal("order 42 disappeared")
	}
	if order.ClaimedBy != claimedBy {
		t.Fatalf("stored claimant = %q, successful worker = %q", order.ClaimedBy, claimedBy)
	}
}

type workerResult struct {
	workerID string
	result   ClaimResult
	err      error
}

type commitGate struct {
	arrived chan struct{}
	release chan struct{}
	count   int
}

func newCommitGate(count int) *commitGate {
	return &commitGate{
		arrived: make(chan struct{}, count),
		release: make(chan struct{}),
		count:   count,
	}
}

func (g *commitGate) Wait() {
	g.arrived <- struct{}{}
	<-g.release
}

func (g *commitGate) ReleaseWhenAllArrive() {
	for range g.count {
		<-g.arrived
	}
	close(g.release)
}
