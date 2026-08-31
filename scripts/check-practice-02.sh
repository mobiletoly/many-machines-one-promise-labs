#!/bin/sh

set -eu

repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repository_root"

GOWORK=off go run \
	./practices/02-one-incident-two-ledgers/cmd/checkmanifest
GOWORK=off go test ./practices/02-one-incident-two-ledgers/...
GOWORK=off go test -race ./practices/02-one-incident-two-ledgers/...
GOWORK=off go vet ./practices/02-one-incident-two-ledgers/...

printf '%s\n' "Practice 02 repository check passed."
