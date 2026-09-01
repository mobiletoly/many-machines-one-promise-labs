//go:build failure

package authority

import (
	"errors"
	"fmt"
	"testing"
)

func TestLocalAuthorityPreservesCapacityDuringPartition(t *testing.T) {
	system := newEvent8(t)
	sales := event8Sales()
	accepted := map[string]int{boothA: 0, boothB: 0}
	type retainedSale struct {
		sale         Sale
		confirmation Confirmation
	}
	retained := map[string][]retainedSale{boothA: {}, boothB: {}}
	authorityOwners := make(map[string]string)
	overlappingAuthority := ""

	for _, boothID := range []string{boothA, boothB} {
		for _, sale := range sales[boothID] {
			confirmation, err := system.ConfirmAt(boothID, sale)
			switch {
			case err == nil:
				accepted[boothID]++
				retained[boothID] = append(
					retained[boothID],
					retainedSale{sale: sale, confirmation: confirmation},
				)
				if owner, exists := authorityOwners[confirmation.AuthorityID]; exists {
					overlappingAuthority = fmt.Sprintf(
						"%s used by %s and %s",
						confirmation.AuthorityID,
						owner,
						boothID,
					)
				} else {
					authorityOwners[confirmation.AuthorityID] = boothID
				}
			case errors.Is(err, ErrNoAuthority):
			default:
				t.Fatalf("confirm %s at %s: %v", sale.OperationID, boothID, err)
			}
		}
	}

	accounting := system.Accounting()
	if accounting.Exposure() > accounting.Capacity {
		t.Fatalf(
			"property violated: event E-8 exposure = confirmed %d + outstanding %d + reserve %d = %d, capacity %d",
			accounting.Confirmed,
			accounting.Outstanding,
			accounting.Reserve,
			accounting.Exposure(),
			accounting.Capacity,
		)
	}
	if overlappingAuthority != "" {
		t.Fatalf("authority overlap violated: %s", overlappingAuthority)
	}
	if accepted[boothA] != 2 || accepted[boothB] != 2 {
		t.Fatalf(
			"healthy local progress violated: accepted = booth-a:%d booth-b:%d, want booth-a:2 booth-b:2",
			accepted[boothA],
			accepted[boothB],
		)
	}

	for _, boothID := range []string{boothA, boothB} {
		booth, ok := system.Booth(boothID)
		if !ok {
			t.Fatalf("%s disappeared", boothID)
		}
		if booth.Confirmed != 2 || booth.Outstanding != 0 {
			t.Fatalf(
				"retained progress violated: %s = %+v, want two confirmations and no outstanding authority",
				boothID,
				booth,
			)
		}
	}
	if accounting.Confirmed != 7 || accounting.Outstanding != 0 ||
		accounting.Reserve != 0 || accounting.Exposure() != 7 {
		t.Fatalf(
			"final accounting = %+v, exposure %d; want confirmed 7, outstanding 0, reserve 0, exposure 7",
			accounting,
			accounting.Exposure(),
		)
	}

	for _, boothID := range []string{boothA, boothB} {
		for _, record := range retained[boothID] {
			replayed, err := system.ConfirmAt(boothID, record.sale)
			if err != nil {
				t.Fatalf("replay %s at %s: %v", record.sale.OperationID, boothID, err)
			}
			if replayed != record.confirmation {
				t.Fatalf(
					"replay %s at %s = %+v, want retained %+v",
					record.sale.OperationID,
					boothID,
					replayed,
					record.confirmation,
				)
			}
		}
	}
	if afterReplay := system.Accounting(); afterReplay != accounting {
		t.Fatalf("matching replays changed accounting from %+v to %+v", accounting, afterReplay)
	}
}
