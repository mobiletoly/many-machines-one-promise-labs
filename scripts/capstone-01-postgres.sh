#!/bin/sh

set -eu

repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
compose_file="$repository_root/capstones/01-two-servers-one-operation/compose.yaml"
capstone_project=${MMOP_CAPSTONE_PROJECT:-mmop-capstone-01-reader}

case ${1:-} in
up)
	docker compose -p "$capstone_project" -f "$compose_file" up -d --wait
	;;
url)
	published=$(docker compose -p "$capstone_project" -f "$compose_file" port postgres 5432)
	case "$published" in
	127.0.0.1:*) ;;
	*)
		printf '%s\n' "Unexpected PostgreSQL port mapping: $published" >&2
		exit 1
		;;
	esac
	printf 'postgres://mmop:mmop@%s/mmop?sslmode=disable\n' "$published"
	;;
down)
	docker compose -p "$capstone_project" -f "$compose_file" down --volumes --remove-orphans
	;;
*)
	printf '%s\n' "Usage: $0 {up|url|down}" >&2
	exit 2
	;;
esac
