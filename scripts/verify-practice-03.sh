#!/bin/sh

set -eu

repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repository_root"

if [ "$#" -ne 1 ]; then
	printf '%s\n' "usage: $0 /path/to/reader-review.json" >&2
	exit 2
fi

review_path=$1
if [ ! -f "$review_path" ]; then
	printf '%s\n' "Practice 03 verifier: review file not found: $review_path" >&2
	exit 2
fi

GOWORK=off go run ./practices/03-which-histories-are-legal/cmd/replay \
	-review "$review_path" \
	-show-violations
GOWORK=off go test -count=1 \
	./practices/03-which-histories-are-legal/replay \
	-run '^TestBoundary'

printf '%s\n' "Practice 03 reader review satisfies the declared contract."
