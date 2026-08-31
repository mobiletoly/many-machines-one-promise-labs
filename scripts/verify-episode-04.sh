#!/bin/sh

set -eu

repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repository_root"

happy_command="go test -count=1 ./episodes/04-stale-completion/start -run '^TestCurrentOwnerCompletesOrder73$'"
failure_command="go test -count=1 -tags=failure ./episodes/04-stale-completion/start -run '^TestFormerOwnerCannotCompleteAfterTransfer$'"
solution_command="go test -count=1 -tags=failure ./episodes/04-stale-completion/solution -run '^TestFormerOwnerCannotCompleteAfterTransfer$'"
boundary_command="go test -count=1 ./episodes/04-stale-completion/solution -run '^TestBoundaryStaleRejectionLeavesPreparedDrink$'"
expected_failure="property violated: completion(order=73, worker=worker-a, assignment=17) = prepared, want stale_assignment"

cmp episodes/04-stale-completion/start/preparation.go episodes/04-stale-completion/solution/preparation.go
cmp episodes/04-stale-completion/start/store_test.go episodes/04-stale-completion/solution/store_test.go
cmp episodes/04-stale-completion/start/property_test.go episodes/04-stale-completion/solution/property_test.go

GOWORK=off sh -c "$happy_command"

failure_output_file=$(mktemp "${TMPDIR:-/tmp}/episode-04-failure.XXXXXX")
trap 'rm -f "$failure_output_file"' EXIT HUP INT TERM

if GOWORK=off sh -c "$failure_command" >"$failure_output_file" 2>&1; then
	printf '%s\n' "Episode 04 verifier: start failure command unexpectedly passed" >&2
	exit 1
fi
if ! grep -F "$expected_failure" "$failure_output_file" >/dev/null; then
	printf '%s\n' "Episode 04 verifier: expected property assertion was not observed" >&2
	sed -n '1,160p' "$failure_output_file" >&2
	exit 1
fi

GOWORK=off sh -c "$solution_command"
GOWORK=off sh -c "$boundary_command"
GOWORK=off go test ./episodes/04-stale-completion/start
GOWORK=off go test ./episodes/04-stale-completion/solution
GOWORK=off go test -race ./episodes/04-stale-completion/solution
GOWORK=off go vet ./episodes/04-stale-completion/...

printf '%s\n' "Episode 04 verification passed."
