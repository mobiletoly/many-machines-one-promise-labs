#!/bin/sh

set -eu

repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repository_root"

happy_command="go test -count=1 ./episodes/02-unknown-create-outcome/start -run ^TestCreateOrderHappyPath$"
failure_command="go test -count=1 -tags=failure ./episodes/02-unknown-create-outcome/start -run ^TestRetryAfterCommittedResponseIsLost$"
solution_command="go test -count=1 -tags=failure ./episodes/02-unknown-create-outcome/solution -run ^TestRetryAfterCommittedResponseIsLost$"
boundary_command="go test -count=1 ./episodes/02-unknown-create-outcome/solution"
expected_failure="property violated: canonical_orders(operation=op-61) = 2, want <= 1"

GOWORK=off sh -c "$happy_command"

failure_output_file=$(mktemp "${TMPDIR:-/tmp}/episode-02-failure.XXXXXX")
trap 'rm -f "$failure_output_file"' EXIT HUP INT TERM

if GOWORK=off sh -c "$failure_command" >"$failure_output_file" 2>&1; then
	printf '%s\n' "Episode 02 verifier: start failure command unexpectedly passed" >&2
	exit 1
fi
if ! grep -F "$expected_failure" "$failure_output_file" >/dev/null; then
	printf '%s\n' "Episode 02 verifier: expected property assertion was not observed" >&2
	sed -n '1,160p' "$failure_output_file" >&2
	exit 1
fi

GOWORK=off sh -c "$solution_command"
GOWORK=off sh -c "$boundary_command"
GOWORK=off go test -race ./episodes/02-unknown-create-outcome/solution
GOWORK=off go vet ./episodes/02-unknown-create-outcome/...

printf '%s\n' "Episode 02 verification passed."
