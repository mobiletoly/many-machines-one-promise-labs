#!/bin/sh

set -eu

repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repository_root"

GOWORK=off go run ./episodes/08-exclusive-authority-transfer/cmd/checkmanifest

happy_command="go test -count=1 ./episodes/08-exclusive-authority-transfer/start -run '^TestGrantFirstWorkflowCanReachOneFinalHolder$'"
failure_command="go test -count=1 -tags=failure ./episodes/08-exclusive-authority-transfer/start -run '^TestExclusiveAuthoritySurvivesGrantDelivery$'"
solution_command="go test -count=1 -tags=failure ./episodes/08-exclusive-authority-transfer/solution -run '^TestExclusiveAuthoritySurvivesGrantDelivery$'"
boundary_command="go test -count=1 ./episodes/08-exclusive-authority-transfer/solution -run '^TestBoundaryLostGrantLeavesAuthorityGap$'"
expected_failure="exclusive authority violated: right R-100 confirmed by A-301 at booth-a and B-401 at booth-b during X-100"

cmp episodes/08-exclusive-authority-transfer/start/property_test.go episodes/08-exclusive-authority-transfer/solution/property_test.go

GOWORK=off sh -c "$happy_command"

failure_output_file=$(mktemp "${TMPDIR:-/tmp}/episode-08-failure.XXXXXX")
trap 'rm -f "$failure_output_file"' EXIT HUP INT TERM

if GOWORK=off sh -c "$failure_command" >"$failure_output_file" 2>&1; then
	printf '%s\n' "Episode 08 verifier: start failure command unexpectedly passed" >&2
	exit 1
fi
if ! grep -F "$expected_failure" "$failure_output_file" >/dev/null; then
	printf '%s\n' "Episode 08 verifier: expected property assertion was not observed" >&2
	sed -n '1,180p' "$failure_output_file" >&2
	exit 1
fi

GOWORK=off sh -c "$solution_command"
GOWORK=off sh -c "$boundary_command"
GOWORK=off go test ./episodes/08-exclusive-authority-transfer/start
GOWORK=off go test ./episodes/08-exclusive-authority-transfer/solution
GOWORK=off go test -race ./episodes/08-exclusive-authority-transfer/solution
GOWORK=off go vet ./episodes/08-exclusive-authority-transfer/...

printf '%s\n' "Episode 08 verification passed."
