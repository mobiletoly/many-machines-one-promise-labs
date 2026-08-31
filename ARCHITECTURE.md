# Labs Architecture

This repository contains the runnable code, controlled failures, reader
artifacts, and verification tools for *One Promise Under Test*.

## Repository boundary

The Labs repository owns executable behavior:

- Go source and tests;
- deterministic fault seams;
- `start` and `solution` states;
- Practice evidence and reader artifacts;
- manifests and verification scripts;
- immutable lab releases.

The companion owns the teaching narrative and publication metadata. Published
units link to a tag and exact commit from this repository. Labs build and run
without access to the companion or Book 1 repositories.

The Go module identifies this repository as:

```text
github.com/mobiletoly/many-machines-one-promise-labs
```

## Unit types

### Episode

An Episode uses a small running system to expose one guarantee boundary:

```text
working path
    -> declared property
    -> prediction
    -> controlled legal execution
    -> observed violation
    -> correction
    -> guarantee test
    -> boundary test
```

The code serves that sequence. An Episode adds a process, durable store, or
network path only when its failure requires that boundary.

### Integration Capstone

An Integration Capstone transfers selected guarantees into a
production-shaped system. Capstone 01 runs two independent HTTP service
processes against one PostgreSQL database. Docker Compose provides the default
database bootstrap, and `DATABASE_URL` supports an existing PostgreSQL
instance.

### Practice

A Practice starts from evidence and a declared contract. The reader edits an
engineering artifact, such as a policy, incident analysis, history review,
recovery decision, or product-claim review. A deterministic oracle evaluates
the artifact against several legal cases and stronger-claim boundaries.

## Repository layout

```text
many-machines-one-promise-labs/
├── episodes/
│   ├── 01-concurrent-claim/
│   ├── 02-unknown-create-outcome/
│   ├── 03-g42-partition/
│   ├── 04-stale-completion/
│   ├── 05-graceful-degradation/
│   └── 06-fencing-token/
├── capstones/
│   └── 01-two-servers-one-operation/
├── practices/
│   ├── 01-from-evidence-to-action/
│   ├── 02-one-incident-two-ledgers/
│   ├── 03-which-histories-are-legal/
│   ├── 04-when-is-recovery-complete/
│   └── 05-what-may-the-product-claim/
└── scripts/
```

Each unit owns a public README and a JSON manifest. Episodes and the Capstone
provide `start` and `solution` code. Practices provide a starter artifact, a
reference artifact, and an inspectable replay implementation.

## Start and solution states

The `start` state represents a plausible implementation before the correction.
It must:

- compile without another unit;
- pass its healthy-path tests;
- reproduce the named property violation through a controlled execution;
- fail for the expected assertion instead of a panic, timeout, or missing
  dependency.

The `solution` state contains the correction earned by the declared contract.
It must:

- pass the property test that fails against `start`;
- preserve the unit's earlier properties;
- include a boundary test for a stronger unsupported claim;
- avoid mechanisms that the current property does not require.

Both states live in one immutable release so the reader can inspect and run
them without reconstructing Git history.

## Deterministic failures

Ordering-dependent tests use explicit barriers, controlled clocks, response
drops, process controls, or network proxies. Sleeps and retry loops cannot
establish the required execution.

A verifier distinguishes the expected property assertion from other command
failures. Repetition can expose flaky code, but it cannot replace a selected
execution.

Every standalone Go command uses `GOWORK=off`. A workspace checkout must not
hide a missing dependency.

## Manifests

Episodes use `episode.json`, the Capstone uses `capstone.json`, and Practices
use `practice.json`. A manifest records:

- unit number, stable ID, title, and type;
- related concepts or capabilities;
- artifact paths;
- runnable commands;
- expected exit codes and required output.

Manifest paths stay inside their unit. They contain no paths to private
repositories.

## Verification

### Episodes

An Episode verifier performs this sequence:

1. validate `episode.json`;
2. compile the start state;
3. run the healthy path;
4. reproduce the named failure;
5. match the expected property assertion;
6. run the corrected execution;
7. run the boundary test;
8. test the solution with the race detector.

### Integration Capstone

The Capstone verifier prepares PostgreSQL, then runs the healthy path, named
two-server failure, correction, process-replacement transfer, and boundary
tests. Its harness creates one isolated schema per test and drops that schema
during cleanup. The Docker path uses PostgreSQL 17.10 on an ephemeral local
port and removes its containers, network, and volumes after the run.

### Practices

A Practice exposes two command surfaces:

- the reader verifier evaluates an artifact supplied by path;
- the repository check validates the manifest, starter failure, bundled
  reference, boundaries, tests, race detector, and static analysis.

The reader verifier must reject a missing artifact path. Its oracle binds each
conclusion to the unit's declared evidence and rules instead of comparing the
whole file with a hidden answer key.

## Shared code

Repository tooling may share manifest parsing, command execution, output
matching, and release packaging. Lesson mechanisms stay inside their unit
unless several units require the same abstraction without hiding the
execution.

Small fault seams often remain local. A barrier or response drop should expose
the incident, not choose the conclusion for the reader.

## Infrastructure

The six Episodes use only the infrastructure required by their failures:

| Unit | Main boundary |
|---|---|
| Episode 01 | one process and controlled goroutine interleaving |
| Episode 02 | HTTP, durable storage, and response loss after commit |
| Episode 03 | two processes and a controlled network partition |
| Episode 04 | responsibility transfer and a paused former owner |
| Episode 05 | a controlled result-selection boundary |
| Episode 06 | separate authority and effect-store generations |

Integration Capstone 01 adds two long-lived HTTP processes and PostgreSQL.
Practices use local evidence files and replay programs.

## Releases

Git tags identify published lab versions. The companion pins both the tag and
the full commit:

```yaml
lab:
  repository: https://github.com/mobiletoly/many-machines-one-promise-labs
  tag: episode-01-v1
  commit: FULL_40_CHARACTER_COMMIT_SHA
  episode: 01-concurrent-claim
```

Published prose links to an immutable tag or commit instead of `main`. A tag
contains every state and artifact required by its unit. Do not move or rewrite
published tags.

A lab correction receives a new tag. The companion then records the new lab
identity before publishing corrected prose.

## Public unit documentation

An Episode or Capstone README states:

- the scenario and declared property;
- prerequisites and supported platforms;
- healthy, failure, solution, and boundary commands;
- the expected assertion from the named failure;
- a link to the published companion unit when available.

A Practice README states the evidence, reader artifact, commands, and claim
boundaries. Unit READMEs do not copy the companion narrative.
