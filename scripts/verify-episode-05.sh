#!/bin/sh

set -eu

repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repository_root"

happy_command="go test -count=1 ./episodes/05-graceful-degradation/start -run '^TestBrowseReturnsFullResultWhenRecommendationArrives$'"
failure_command="go test -count=1 -tags=failure ./episodes/05-graceful-degradation/start -run '^TestBrowsePreservesCatalogAtSelectionBoundary$'"
solution_command="go test -count=1 -tags=failure ./episodes/05-graceful-degradation/solution -run '^TestBrowsePreservesCatalogAtSelectionBoundary$'"
boundary_command="go test -count=1 ./episodes/05-graceful-degradation/solution -run '^TestBoundaryRecommendationCannotReplaceMissingCatalog$'"
expected_failure="property violated: browse result = error recommendation unavailable, want reduced success with current catalog"

cmp episodes/05-graceful-degradation/start/fixtures_test.go episodes/05-graceful-degradation/solution/fixtures_test.go
cmp episodes/05-graceful-degradation/start/service_test.go episodes/05-graceful-degradation/solution/service_test.go
cmp episodes/05-graceful-degradation/start/property_test.go episodes/05-graceful-degradation/solution/property_test.go

GOWORK=off sh -c "$happy_command"

failure_output_file=$(mktemp "${TMPDIR:-/tmp}/episode-05-failure.XXXXXX")
trap 'rm -f "$failure_output_file"' EXIT HUP INT TERM

if GOWORK=off sh -c "$failure_command" >"$failure_output_file" 2>&1; then
	printf '%s\n' "Episode 05 verifier: start failure command unexpectedly passed" >&2
	exit 1
fi
if ! grep -F "$expected_failure" "$failure_output_file" >/dev/null; then
	printf '%s\n' "Episode 05 verifier: expected property assertion was not observed" >&2
	sed -n '1,160p' "$failure_output_file" >&2
	exit 1
fi

GOWORK=off sh -c "$solution_command"
GOWORK=off sh -c "$boundary_command"
GOWORK=off go test ./episodes/05-graceful-degradation/start
GOWORK=off go test ./episodes/05-graceful-degradation/solution
GOWORK=off go test -race ./episodes/05-graceful-degradation/solution
GOWORK=off go vet ./episodes/05-graceful-degradation/...

printf '%s\n' "Episode 05 verification passed."
