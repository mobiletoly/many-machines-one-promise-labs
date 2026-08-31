#!/bin/sh

set -eu

repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repository_root"

if [ "$#" -ne 1 ]; then
	printf '%s\n' "usage: $0 /path/to/recovery-decision.json" >&2
	exit 2
fi

decision_path=$1
if [ ! -f "$decision_path" ]; then
	printf '%s\n' "Practice 04 verifier: decision file not found: $decision_path" >&2
	exit 2
fi

GOWORK=off go run ./practices/04-when-is-recovery-complete/cmd/replay \
	-decision "$decision_path" \
	-show-violations
GOWORK=off go test -count=1 \
	./practices/04-when-is-recovery-complete/replay \
	-run '^TestBoundary'

printf '%s\n' "Practice 04 reader decision satisfies the declared recovery contract."
