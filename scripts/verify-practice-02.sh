#!/bin/sh

set -eu

repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repository_root"

if [ "$#" -ne 1 ]; then
	printf '%s\n' "usage: $0 /path/to/reader-analysis.json" >&2
	exit 2
fi

analysis_path=$1
if [ ! -f "$analysis_path" ]; then
	printf '%s\n' "Practice 02 verifier: analysis file not found: $analysis_path" >&2
	exit 2
fi

GOWORK=off go run ./practices/02-one-incident-two-ledgers/cmd/replay \
	-analysis "$analysis_path" \
	-show-violations
GOWORK=off go test -count=1 \
	./practices/02-one-incident-two-ledgers/replay \
	-run '^TestBoundary'

printf '%s\n' "Practice 02 reader analysis satisfies the declared contract."
