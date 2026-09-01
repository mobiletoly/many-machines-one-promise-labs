#!/bin/sh

set -eu

repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repository_root"

if [ "$#" -ne 1 ]; then
	printf '%s\n' "usage: $0 /path/to/adjudication.json" >&2
	exit 2
fi

review_path=$1
if [ ! -f "$review_path" ]; then
	printf '%s\n' "Practice 06 verifier: review file not found: $review_path" >&2
	exit 2
fi

GOWORK=off go run ./practices/06-atomic-commit-adjudication/cmd/replay \
	-review "$review_path" \
	-show-violations
GOWORK=off go test -count=1 \
	./practices/06-atomic-commit-adjudication/replay \
	-run '^TestBoundary'

printf '%s\n' "Practice 06 reader adjudication satisfies the declared recovery contract."
