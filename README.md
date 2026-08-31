# Many Machines, One Promise Labs

Runnable Go labs for *One Promise Under Test*, the companion to *Many
Machines, One Promise*.

Each lab starts with a scoped system property. You predict a legal execution,
run it, inspect the failure, apply the correction, and test the boundary of the
resulting guarantee.

## Requirements

- Go 1.23 or newer;
- macOS or Linux;
- a POSIX shell;
- Docker with Docker Compose, or access to PostgreSQL, for Integration
  Capstone 01.

## Get the code

```sh
git clone https://github.com/mobiletoly/many-machines-one-promise-labs.git
cd many-machines-one-promise-labs
```

The companion links each published unit to an immutable release. Check out the
tag named by the unit before running its commands.

## Episodes

| # | Lab | Complete verifier |
|---|---|---|
| 01 | [Two Workers Claim Order 42](episodes/01-concurrent-claim/) | `./scripts/verify-episode-01.sh` |
| 02 | [The Commit Happened, the Response Did Not](episodes/02-unknown-create-outcome/) | `./scripts/verify-episode-02.sh` |
| 03 | [G-42 Splits and Reunites](episodes/03-g42-partition/) | `./scripts/verify-episode-03.sh` |
| 04 | [Worker A Comes Back with Order 73](episodes/04-stale-completion/) | `./scripts/verify-episode-04.sh` |
| 05 | [The List Without Advice](episodes/05-graceful-degradation/) | `./scripts/verify-episode-05.sh` |
| 06 | [The Publication Desk Has Seen 18](episodes/06-fencing-token/) | `./scripts/verify-episode-06.sh` |

Every Episode contains independently runnable `start` and `solution` states.
Its README lists the happy path, controlled failure, corrected execution, and
boundary command.

## Integration Capstone

[Two Servers, One Operation](capstones/01-two-servers-one-operation/) runs two
Go HTTP service processes against PostgreSQL. The verifier starts PostgreSQL
through Docker Compose when `DATABASE_URL` is unset.

```sh
./scripts/verify-capstone-01.sh
```

## Practices

| # | Practice | Repository check |
|---|---|---|
| 01 | [From Evidence to Action](practices/01-from-evidence-to-action/) | `./scripts/check-practice-01.sh` |
| 02 | [One Incident, Two Ledgers](practices/02-one-incident-two-ledgers/) | `./scripts/check-practice-02.sh` |
| 03 | [Which Histories Are Legal?](practices/03-which-histories-are-legal/) | `./scripts/check-practice-03.sh` |
| 04 | [When Is Recovery Complete?](practices/04-when-is-recovery-complete/) | `./scripts/check-practice-04.sh` |
| 05 | [What May the Product Claim?](practices/05-what-may-the-product-claim/) | `./scripts/check-practice-05.sh` |

A Practice gives you evidence and a starter artifact. Its replay command
rejects unsupported conclusions. After editing the artifact, pass its path to
the reader verifier shown in the Practice README.

## Repository checks

Run the Go tests and static analysis from the repository root:

```sh
GOWORK=off go test -count=1 ./...
GOWORK=off go vet ./...
```

Run each unit verifier for the full teaching checks:

```sh
./scripts/verify-episode-01.sh
./scripts/verify-episode-02.sh
./scripts/verify-episode-03.sh
./scripts/verify-episode-04.sh
./scripts/verify-episode-05.sh
./scripts/verify-episode-06.sh
./scripts/verify-capstone-01.sh
./scripts/check-practice-01.sh
./scripts/check-practice-02.sh
./scripts/check-practice-03.sh
./scripts/check-practice-04.sh
./scripts/check-practice-05.sh
```

## Repository layout

```text
episodes/    controlled running-system failures and corrections
capstones/   production-shaped integration work
practices/   evidence, reader artifacts, and deterministic oracles
scripts/     complete unit verifiers and PostgreSQL bootstrap
```

[ARCHITECTURE.md](ARCHITECTURE.md) describes the unit, determinism, manifest,
and release contracts.

## License

Copyright 2026 Toly Pochkin.

Licensed under the [Apache License, Version 2.0](LICENSE).
