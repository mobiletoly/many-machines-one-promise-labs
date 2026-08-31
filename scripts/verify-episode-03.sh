#!/bin/sh

set -eu

repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repository_root"

happy_command="go test -count=1 ./episodes/03-g42-partition/start -run '^TestOneReplicaRedeemsLocally$'"
failure_command="go test -count=1 -tags=failure ./episodes/03-g42-partition/start -run '^TestReplicasPreserveAcceptedHistoryAfterPartition$'"
solution_command="go test -count=1 -tags=failure ./episodes/03-g42-partition/solution -run '^TestReplicasPreserveAcceptedHistoryAfterPartition$'"
boundary_command="go test -count=1 ./episodes/03-g42-partition/solution -run '^TestBoundaryConvergedHistoryStillViolatesFunding$'"
expected_failure="property violated: replica A accounts for [RA-80], want [RA-80 RB-80]"

cmp episodes/03-g42-partition/start/http.go episodes/03-g42-partition/solution/http.go
cmp episodes/03-g42-partition/start/system_test.go episodes/03-g42-partition/solution/system_test.go
cmp episodes/03-g42-partition/start/property_test.go episodes/03-g42-partition/solution/property_test.go

GOWORK=off sh -c "$happy_command"

failure_output_file=$(mktemp "${TMPDIR:-/tmp}/episode-03-failure.XXXXXX")
trap 'rm -f "$failure_output_file"' EXIT HUP INT TERM

if GOWORK=off sh -c "$failure_command" >"$failure_output_file" 2>&1; then
	printf '%s\n' "Episode 03 verifier: start failure command unexpectedly passed" >&2
	exit 1
fi
if ! grep -F "$expected_failure" "$failure_output_file" >/dev/null; then
	printf '%s\n' "Episode 03 verifier: expected property assertion was not observed" >&2
	sed -n '1,200p' "$failure_output_file" >&2
	exit 1
fi

GOWORK=off sh -c "$solution_command"
GOWORK=off sh -c "$boundary_command"
GOWORK=off go test ./episodes/03-g42-partition/solution
GOWORK=off go test -race ./episodes/03-g42-partition/solution
GOWORK=off go vet ./episodes/03-g42-partition/...

printf '%s\n' "Episode 03 verification passed."
