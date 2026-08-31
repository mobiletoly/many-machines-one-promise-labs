#!/bin/sh

set -eu

repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repository_root"

GOWORK=off go run \
	./practices/01-from-evidence-to-action/cmd/checkmanifest
GOWORK=off go test ./practices/01-from-evidence-to-action/...
GOWORK=off go test -race ./practices/01-from-evidence-to-action/...
GOWORK=off go vet ./practices/01-from-evidence-to-action/...

printf '%s\n' "Practice 01 repository check passed."
