#!/bin/sh

set -eu

repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repository_root"

if [ "$#" -ne 1 ]; then
	printf '%s\n' "usage: $0 /path/to/reader-policy.json" >&2
	exit 2
fi

policy_path=$1
if [ ! -f "$policy_path" ]; then
	printf '%s\n' "Practice 01 verifier: policy file not found: $policy_path" >&2
	exit 2
fi

GOWORK=off go run ./practices/01-from-evidence-to-action/cmd/replay \
	-policy "$policy_path"
GOWORK=off go test -count=1 \
	./practices/01-from-evidence-to-action/replay \
	-run '^TestBoundary'

printf '%s\n' "Practice 01 reader policy satisfies the declared contract."
