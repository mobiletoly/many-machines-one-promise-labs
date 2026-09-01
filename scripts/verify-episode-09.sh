#!/bin/sh

set -eu

repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repository_root"

GOWORK=off go run ./episodes/09-time-bounded-authority/cmd/checkmanifest

happy_command="go test -count=1 ./episodes/09-time-bounded-authority/start -run '^TestEstablishedLeaseSupportsPublicationWithoutIssuer$'"
failure_command="go test -count=1 -tags=failure ./episodes/09-time-bounded-authority/start -run '^TestLeaseExpiryIsEnforcedAtEffectAcceptance$'"
solution_command="go test -count=1 -tags=failure ./episodes/09-time-bounded-authority/solution -run '^TestLeaseExpiryIsEnforcedAtEffectAcceptance$'"
boundary_command="go test -count=1 ./episodes/09-time-bounded-authority/solution -run '^TestBoundary'"
expected_failure="lease expiry violated: publish(operation=P-after, lease=L-88) decided at S time 110 = accepted, want lease_expired"

cmp episodes/09-time-bounded-authority/start/fixtures_test.go episodes/09-time-bounded-authority/solution/fixtures_test.go
cmp episodes/09-time-bounded-authority/start/property_test.go episodes/09-time-bounded-authority/solution/property_test.go
cmp episodes/09-time-bounded-authority/start/system_test.go episodes/09-time-bounded-authority/solution/system_test.go

GOWORK=off sh -c "$happy_command"

failure_output_file=$(mktemp "${TMPDIR:-/tmp}/episode-09-failure.XXXXXX")
trap 'rm -f "$failure_output_file"' EXIT HUP INT TERM

if GOWORK=off sh -c "$failure_command" >"$failure_output_file" 2>&1; then
	printf '%s\n' "Episode 09 verifier: start failure command unexpectedly passed" >&2
	exit 1
fi
if ! grep -F "$expected_failure" "$failure_output_file" >/dev/null; then
	printf '%s\n' "Episode 09 verifier: expected property assertion was not observed" >&2
	sed -n '1,180p' "$failure_output_file" >&2
	exit 1
fi

GOWORK=off sh -c "$solution_command"
GOWORK=off sh -c "$boundary_command"
GOWORK=off go test ./episodes/09-time-bounded-authority/start
GOWORK=off go test ./episodes/09-time-bounded-authority/solution
GOWORK=off go test -race ./episodes/09-time-bounded-authority/solution
GOWORK=off go vet ./episodes/09-time-bounded-authority/...

printf '%s\n' "Episode 09 verification passed."
