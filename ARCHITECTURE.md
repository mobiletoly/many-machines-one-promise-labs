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
│   ├── 06-fencing-token/
│   ├── 07-partitioned-authority/
│   ├── 08-exclusive-authority-transfer/
│   └── 09-time-bounded-authority/
├── capstones/
│   └── 01-two-servers-one-operation/
├── practices/
│   ├── 01-from-evidence-to-action/
│   ├── 02-one-incident-two-ledgers/
│   ├── 03-which-histories-are-legal/
│   ├── 04-when-is-recovery-complete/
│   ├── 05-what-may-the-product-claim/
│   └── 06-atomic-commit-adjudication/
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

The eight assembled Episodes and the retained Episode 09 workshop artifact use
only the infrastructure required by their failures:

| Unit | Main boundary |
|---|---|
| Episode 01 | one process and controlled goroutine interleaving |
| Episode 02 | HTTP, durable storage, and response loss after commit |
| Episode 03 | two processes and a controlled network partition |
| Episode 04 | responsibility transfer and a paused former owner |
| Episode 05 | a controlled result-selection boundary |
| Episode 06 | separate authority and effect-store generations |
| Episode 07 | two local authority stores and finite-right accounting |
| Episode 08 | two local authority stores and controlled grant delivery |
| Episode 09 rejected workshop artifact | one effect authority and a controlled local clock |

Integration Capstone 01 adds two long-lived HTTP processes and PostgreSQL.
Practices 01-06 use local evidence files and replay programs.

## Post-v1 authority candidates

On 2026-09-01, the human author admitted two linked post-v1 Episode candidates
for Book 1 Chapters 11 and 12. Episode 07 covers partitioned local authority
over one finite invariant. Episode 08 covers transfer of one exclusive right.
The workshop keeps them separate because allocation and transfer require
different failure histories, corrections, oracles, and boundaries.

Episode 07 starts with two healthy booth stores that treat the same observed
remaining count as local permission. The correction allocates non-overlapping
rights before communication loss and consumes one right with each retained
confirmation. Its safety oracle counts confirmed sales, outstanding spendable
rights, and reserve. A separate progress assertion requires both booths to use
their declared local allocations. The boundary strands usable rights at the
booth without current demand.

Episode 07 completed its corrected local loop and received a human freeze. It
remains outside the v1 release.

Episode 08 begins with Booth B owning `R-100` and Booth A owning no local right.
The start state emits an actionable `X-100` grant before Booth B relinquishes
the right. The correction relinquishes and retains the grant in one source
transition before delivery. The shared property test requires destination
progress and rejects two distinct confirmations backed by `R-100`. The boundary
loses the grant after relinquishment and exposes a safe authority gap.

Episode 08 adds no live-process requirement, crash recovery, forged-message
defense, successive transfer, replication, merge, clock, lease, fencing,
quorum, or consensus mechanism.

Independent paired review found that the start source supplied the failure
answer before prediction and that the complete verifier did not validate
`episode.json`. The human author accepted both findings. The integrated
correction removes the evaluative source comment and validates the manifest's
identity, paths, command strings, exit codes, and failure output before the
teaching sequence runs.

The human author confirmed the Episode 08 freeze on 2026-09-01 for
reader-facing Companion commit
`ca4240c275fae6211513a1543f6b962120e43588` and Labs candidate commit
`c274d261087e8f8de21db2860f92035112327079`. The freeze ends this local review
loop. Episode 08 remains outside the v1 release and has no release tag, push,
or publication authority.

The human author admitted Episode 09, `The Desk Read 110`, after a focused
lease workshop on 2026-09-01. Publication desk S establishes L-88 for A and
M-88 on S's manual monotonic clock. The shared property test accepts one
pre-expiry publication while the issuer remains unavailable, then advances S
from tick 109 to tick 110 at the controlled acceptance boundary. The start
state uses an earlier eligibility check and accepts the late effect. The
solution reads S's clock and accepts or rejects the effect in one desk
transition. Boundary tests reject at the exact expiry and preserve an effect
accepted before expiry.

Episode 09 adds no live process, synchronized clock, real-time duration claim,
renewal, recovery, replacement holder, fencing generation, transfer, failure
detection, retry, or idempotency contract. Independent review and a focused
substitution probe found no reader decision beyond Book 1's complete argument.
The human author killed the candidate. Its frozen code remains retained
workshop evidence and does not belong to the complete public cut.

## Complete-public-cut candidate

On 2026-09-02, the human author admitted Episodes 07-08 and Practice 06 to the
complete-public-cut assembly. The snapshot contains Episodes 01-08,
Integration Capstone 01, and Practices 01-06.

The Companion manifest records the exact Labs commit after this documentation
integration and uses `tag: pending` during assembly review. An immutable tag at
the same commit, push, and publication remain separate release gates. Episode
09 stays outside the assembly.

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
