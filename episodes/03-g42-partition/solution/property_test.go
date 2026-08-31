//go:build failure

package g42

import "testing"

func TestReplicasPreserveAcceptedHistoryAfterPartition(t *testing.T) {
	stateA, stateB := runPartitionScenario(t)
	const wanted = "[RA-80 RB-80]"

	if got := operationIDs(stateA); got != wanted {
		t.Fatalf("property violated: replica A accounts for %s, want %s", got, wanted)
	}
	if got := operationIDs(stateB); got != wanted {
		t.Fatalf("property violated: replica B accounts for %s, want %s", got, wanted)
	}
	if stateA.Confirmed != 160 || stateB.Confirmed != 160 {
		t.Fatalf("confirmed values = A:%d B:%d, want A:160 B:160", stateA.Confirmed, stateB.Confirmed)
	}
}
