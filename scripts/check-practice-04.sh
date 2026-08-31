#!/bin/sh

set -eu

repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repository_root"

GOWORK=off go run \
	./practices/04-when-is-recovery-complete/cmd/checkmanifest
GOWORK=off go test ./practices/04-when-is-recovery-complete/...
GOWORK=off go test -race ./practices/04-when-is-recovery-complete/...
GOWORK=off go vet ./practices/04-when-is-recovery-complete/...

printf '%s\n' "Practice 04 repository check passed."
