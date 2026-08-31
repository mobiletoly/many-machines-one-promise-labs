#!/bin/sh

set -eu

repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repository_root"

GOWORK=off go run \
	./practices/03-which-histories-are-legal/cmd/checkmanifest
GOWORK=off go test ./practices/03-which-histories-are-legal/...
GOWORK=off go test -race ./practices/03-which-histories-are-legal/...
GOWORK=off go vet ./practices/03-which-histories-are-legal/...

printf '%s\n' "Practice 03 repository check passed."
