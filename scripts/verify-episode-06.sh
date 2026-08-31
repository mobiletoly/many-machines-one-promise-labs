#!/bin/sh

set -eu

repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repository_root"

happy_command="go test -count=1 ./episodes/06-fencing-token/start -run '^TestCurrentHolderPublishesManifest77$'"
failure_command="go test -count=1 -tags=failure ./episodes/06-fencing-token/start -run '^TestStoreRejectsOlderGenerationAfterEstablishing18$'"
solution_command="go test -count=1 -tags=failure ./episodes/06-fencing-token/solution -run '^TestStoreRejectsOlderGenerationAfterEstablishing18$'"
boundary_command="go test -count=1 ./episodes/06-fencing-token/solution -run '^TestBoundary'"
expected_failure="property violated: publish(manifest=M-77, holder=coordinator-a, generation=17) = accepted, want stale_generation"

cmp episodes/06-fencing-token/start/controller.go episodes/06-fencing-token/solution/controller.go
cmp episodes/06-fencing-token/start/fixtures_test.go episodes/06-fencing-token/solution/fixtures_test.go
cmp episodes/06-fencing-token/start/store_test.go episodes/06-fencing-token/solution/store_test.go
cmp episodes/06-fencing-token/start/property_test.go episodes/06-fencing-token/solution/property_test.go

GOWORK=off sh -c "$happy_command"

failure_output_file=$(mktemp "${TMPDIR:-/tmp}/episode-06-failure.XXXXXX")
trap 'rm -f "$failure_output_file"' EXIT HUP INT TERM

if GOWORK=off sh -c "$failure_command" >"$failure_output_file" 2>&1; then
	printf '%s\n' "Episode 06 verifier: start failure command unexpectedly passed" >&2
	exit 1
fi
if ! grep -F "$expected_failure" "$failure_output_file" >/dev/null; then
	printf '%s\n' "Episode 06 verifier: expected property assertion was not observed" >&2
	sed -n '1,180p' "$failure_output_file" >&2
	exit 1
fi

GOWORK=off sh -c "$solution_command"
GOWORK=off sh -c "$boundary_command"
GOWORK=off go test ./episodes/06-fencing-token/start
GOWORK=off go test ./episodes/06-fencing-token/solution
GOWORK=off go test -race ./episodes/06-fencing-token/solution
GOWORK=off go vet ./episodes/06-fencing-token/...

printf '%s\n' "Episode 06 verification passed."
