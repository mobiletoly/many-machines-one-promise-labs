#!/bin/sh

set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_root=$(CDPATH= cd -- "$script_dir/.." && pwd)
verification_dir=$(mktemp -d "${TMPDIR:-/tmp}/mmop-episode-01.XXXXXX")
trap 'rm -rf "$verification_dir"' EXIT HUP INT TERM

cd "$repo_root"

GOWORK=off go test -count=1 ./episodes/01-concurrent-claim/start \
  -run '^TestOneWorkerClaimsOrder$'

failure_output="$verification_dir/start-failure.txt"
if GOWORK=off go test -count=1 -tags=failure \
  ./episodes/01-concurrent-claim/start \
  -run '^TestTwoWorkersClaimOrder42$' >"$failure_output" 2>&1; then
  echo "start property test passed; expected the declared violation" >&2
  exit 1
fi

expected='property violated: successful_claims(order=42) = 2, want <= 1'
if ! grep -F "$expected" "$failure_output" >/dev/null; then
  echo "start property test failed for an unexpected reason" >&2
  cat "$failure_output" >&2
  exit 1
fi

GOWORK=off go test -count=1 -tags=failure \
  ./episodes/01-concurrent-claim/solution \
  -run '^TestTwoWorkersClaimOrder42$'

GOWORK=off go test -count=1 ./episodes/01-concurrent-claim/solution \
  -run '^TestBoundaryClaimCanStrandOrder$'

GOWORK=off go test -count=1 ./...

echo "Episode 01 verification passed."
