#!/bin/sh

set -eu

repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repository_root"

happy_command="go test -count=1 ./episodes/07-partitioned-authority/start -run '^TestOneBoothConfirmsFromLocalState$'"
failure_command="go test -count=1 -tags=failure ./episodes/07-partitioned-authority/start -run '^TestLocalAuthorityPreservesCapacityDuringPartition$'"
solution_command="go test -count=1 -tags=failure ./episodes/07-partitioned-authority/solution -run '^TestLocalAuthorityPreservesCapacityDuringPartition$'"
boundary_command="go test -count=1 ./episodes/07-partitioned-authority/solution -run '^TestBoundarySafeAllocationCanStrandCapacity$'"
expected_failure="property violated: event E-8 exposure = confirmed 9 + outstanding 2 + reserve 4 = 15, capacity 7"

cmp episodes/07-partitioned-authority/start/property_test.go episodes/07-partitioned-authority/solution/property_test.go

GOWORK=off sh -c "$happy_command"

failure_output_file=$(mktemp "${TMPDIR:-/tmp}/episode-07-failure.XXXXXX")
trap 'rm -f "$failure_output_file"' EXIT HUP INT TERM

if GOWORK=off sh -c "$failure_command" >"$failure_output_file" 2>&1; then
	printf '%s\n' "Episode 07 verifier: start failure command unexpectedly passed" >&2
	exit 1
fi
if ! grep -F "$expected_failure" "$failure_output_file" >/dev/null; then
	printf '%s\n' "Episode 07 verifier: expected property assertion was not observed" >&2
	sed -n '1,180p' "$failure_output_file" >&2
	exit 1
fi

GOWORK=off sh -c "$solution_command"
GOWORK=off sh -c "$boundary_command"
GOWORK=off go test ./episodes/07-partitioned-authority/start
GOWORK=off go test ./episodes/07-partitioned-authority/solution
GOWORK=off go test -race ./episodes/07-partitioned-authority/solution
GOWORK=off go vet ./episodes/07-partitioned-authority/...

printf '%s\n' "Episode 07 verification passed."
