#!/bin/sh

set -eu

repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repository_root"

GOWORK=off go run \
	./practices/05-what-may-the-product-claim/cmd/checkmanifest
GOWORK=off go test ./practices/05-what-may-the-product-claim/...
GOWORK=off go test -race ./practices/05-what-may-the-product-claim/...
GOWORK=off go vet ./practices/05-what-may-the-product-claim/...

printf '%s\n' "Practice 05 repository check passed."
