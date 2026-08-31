package g42

import "testing"

func TestBoundaryConvergedHistoryStillViolatesFunding(t *testing.T) {
	stateA, stateB := runPartitionScenario(t)
	const wanted = "[RA-80 RB-80]"

	if operationIDs(stateA) != wanted || operationIDs(stateB) != wanted {
		t.Fatalf("replicas did not converge: A=%s B=%s", operationIDs(stateA), operationIDs(stateB))
	}
	if stateA.Confirmed != stateB.Confirmed {
		t.Fatalf("confirmed values differ: A=%d B=%d", stateA.Confirmed, stateB.Confirmed)
	}
	if stateA.Confirmed <= stateA.Funded {
		t.Fatalf("boundary not exposed: confirmed=%d funded=%d", stateA.Confirmed, stateA.Funded)
	}
}
