package g42

import (
	"errors"
	"testing"
)

func TestMergeRejectsConflictingOperationReuse(t *testing.T) {
	account := NewAccount()
	if err := account.Redeem(Redemption{
		OperationID: "RA-80", Person: "Jessica", Amount: 80,
	}); err != nil {
		t.Fatalf("redeem RA-80: %v", err)
	}

	err := account.Merge(Snapshot{
		Funded:    FundedValue,
		Confirmed: 40,
		Operations: []Redemption{{
			OperationID: "RA-80", Person: "Jessica", Amount: 40,
		}},
	})
	if !errors.Is(err, ErrOperationConflict) {
		t.Fatalf("conflicting merge error = %v, want %v", err, ErrOperationConflict)
	}
	if got := account.Snapshot(); got.Confirmed != 80 || operationIDs(got) != "[RA-80]" {
		t.Fatalf("state changed after conflict: %+v", got)
	}
}
