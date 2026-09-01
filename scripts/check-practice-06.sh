#!/bin/sh

set -eu

repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repository_root"

GOWORK=off go run \
	./practices/06-atomic-commit-adjudication/cmd/checkmanifest
GOWORK=off go test ./practices/06-atomic-commit-adjudication/...
GOWORK=off go test -race ./practices/06-atomic-commit-adjudication/...
GOWORK=off go vet ./practices/06-atomic-commit-adjudication/...

printf '%s\n' "Practice 06 repository check passed."
