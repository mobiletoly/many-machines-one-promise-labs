# The Publication Desk Has Seen 18

Episode 06 separates the controller that assigns publication authority from
the store that accepts the official manifest.

```text
C: A/17 -> B/18

B/18 -> S : accepted, establish 18
A/17 -> S : stale_generation
```

## Declared property

After S establishes generation 18 for M-77, S must reject a later publication
with a generation lower than 18 and keep the generation-18 content official.

The start store retains the largest generation it has seen but still accepts a
later lower generation. The solution compares and advances the per-manifest
high-water mark inside the same critical section that accepts official
content.

## Prerequisites and platforms

Install Go 1.23 or newer. The complete verifier requires a POSIX shell and
supports macOS and Linux. Run all commands from the repository root.

## Commands

Run the current-holder happy path:

```sh
GOWORK=off go test -count=1 ./episodes/06-fencing-token/start -run '^TestCurrentHolderPublishesManifest77$'
```

Reproduce the property violation. This command must exit nonzero and print the
named `accepted, want stale_generation` assertion:

```sh
GOWORK=off go test -count=1 -tags=failure ./episodes/06-fencing-token/start -run '^TestStoreRejectsOlderGenerationAfterEstablishing18$'
```

Run the corrected execution:

```sh
GOWORK=off go test -count=1 -tags=failure ./episodes/06-fencing-token/solution -run '^TestStoreRejectsOlderGenerationAfterEstablishing18$'
```

Run both boundary tests:

```sh
GOWORK=off go test -count=1 ./episodes/06-fencing-token/solution -run '^TestBoundary'
```

Run the complete episode verifier from the repository root:

```sh
./scripts/verify-episode-06.sh
```

The boundary tests show that C issuing 18 does not establish 18 at S and that
a second effect authority T remains unprotected until it sees and enforces the
newer generation.

This in-memory episode makes no crash-recovery, authentication,
same-generation, expiry, retry, or multi-boundary activation claim.
