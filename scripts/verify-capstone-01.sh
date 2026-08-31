#!/bin/sh

set -eu

repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repository_root"

managed_database=false
if [ -z "${DATABASE_URL:-}" ]; then
	managed_database=true
	MMOP_CAPSTONE_PROJECT="mmop-capstone-01-verify-$$"
	export MMOP_CAPSTONE_PROJECT
	trap './scripts/capstone-01-postgres.sh down' EXIT HUP INT TERM
	./scripts/capstone-01-postgres.sh up
	DATABASE_URL=$(./scripts/capstone-01-postgres.sh url)
	export DATABASE_URL
fi

happy_command="go test -count=1 -tags=integration ./capstones/01-two-servers-one-operation/start -run '^TestOneServerCreatesOrder$'"
failure_command="go test -count=1 -tags='integration failure' ./capstones/01-two-servers-one-operation/start -run '^TestTwoServersCreateOneOperation$'"
solution_command="go test -count=1 -tags='integration failure' ./capstones/01-two-servers-one-operation/solution -run '^TestTwoServersCreateOneOperation$'"
replacement_command="go test -count=1 -tags=integration ./capstones/01-two-servers-one-operation/solution -run '^TestReplacementReturnsCanonicalResult$'"
boundary_command="go test -count=1 -tags=integration ./capstones/01-two-servers-one-operation/solution -run '^(TestConflictingReuseIsRejectedAcrossServers|TestBoundarySamePayloadDifferentOperationsCreateTwoOrders)$'"
expected_failure="property violated: canonical_orders(operation=op-61) = 2, want <= 1"

GOWORK=off sh -c "$happy_command"

failure_output_file=$(mktemp "${TMPDIR:-/tmp}/capstone-01-failure.XXXXXX")
trap 'rm -f "$failure_output_file"; if [ "$managed_database" = true ]; then ./scripts/capstone-01-postgres.sh down; fi' EXIT HUP INT TERM

if GOWORK=off sh -c "$failure_command" >"$failure_output_file" 2>&1; then
	printf '%s\n' "Capstone 01 verifier: start failure command unexpectedly passed" >&2
	exit 1
fi
if ! grep -F "$expected_failure" "$failure_output_file" >/dev/null; then
	printf '%s\n' "Capstone 01 verifier: expected property assertion was not observed" >&2
	sed -n '1,220p' "$failure_output_file" >&2
	exit 1
fi

GOWORK=off sh -c "$solution_command"
GOWORK=off sh -c "$replacement_command"
GOWORK=off sh -c "$boundary_command"
cmp \
	capstones/01-two-servers-one-operation/start/http.go \
	capstones/01-two-servers-one-operation/solution/http.go
cmp \
	capstones/01-two-servers-one-operation/start/system_test.go \
	capstones/01-two-servers-one-operation/solution/system_test.go
cmp \
	capstones/01-two-servers-one-operation/start/property_test.go \
	capstones/01-two-servers-one-operation/solution/property_test.go
GOWORK=off go test -race -count=1 -tags=integration ./capstones/01-two-servers-one-operation/solution
GOWORK=off go vet -tags=integration ./capstones/01-two-servers-one-operation/...

printf '%s\n' "Integration Capstone 01 verification passed."
