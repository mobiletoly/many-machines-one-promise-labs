# Practice 02: One Incident, Two Ledgers

One checkout incident produces retry work, consumes a finite transaction
boundary, and misses a product objective. You will reconstruct two accounts of
the same raw event stream.

The load ledger counts work and retained capacity. The product ledger counts
eligible logical operations and classifies them under the declared SLO.

## Declared contract

Window `W-2` contains 100 checkout logical operations. One operation is good
when the storefront emits its first contract-valid result no later than 230 ms
after the first accepted invocation. The objective requires at least 99 good
operations.

Each fifth client may retry once after 50 ms without a result. The retry keeps
the original operation identity.

Each `result_emitted` event records `contract_valid`. Only a result marked
`true` can end the good-event latency interval.

The start implementation holds eleven of twelve transaction permits in browse
operations while dependency X remains slow. Five pre-window checkout attempts
already wait for the remaining permit.

## Prerequisites and platforms

Install Go 1.23 or newer. The complete verifier requires a POSIX shell and
supports macOS and Linux. Run commands from the repository root.

## Inspect the raw input

Read these files before running the starter:

```text
practices/02-one-incident-two-ledgers/contract.json
practices/02-one-incident-two-ledgers/evidence.jsonl
practices/02-one-incident-two-ledgers/starter/analysis.json
```

Do not inspect `solution/analysis.json` yet.

## Commands

Run the starter analysis. It must fail without printing the reference ledger:

```sh
GOWORK=off go run ./practices/02-one-incident-two-ledgers/cmd/replay -analysis practices/02-one-incident-two-ledgers/starter/analysis.json
```

Copy the starter artifact and edit your copy:

```sh
cp practices/02-one-incident-two-ledgers/starter/analysis.json \
  /tmp/practice-02-analysis.json
${EDITOR:-vi} /tmp/practice-02-analysis.json
```

Verify your artifact:

```sh
./scripts/verify-practice-02.sh /tmp/practice-02-analysis.json
```

After it passes, compare it with the bundled reference:

```sh
GOWORK=off go run ./practices/02-one-incident-two-ledgers/cmd/replay -analysis practices/02-one-incident-two-ledgers/solution/analysis.json
```

Run the stronger-claim and lost-fungibility boundaries:

```sh
GOWORK=off go test -count=1 ./practices/02-one-incident-two-ledgers/replay -run '^TestBoundary'
```

Run the repository check:

```sh
./scripts/check-practice-02.sh
```

## Boundary

The reference allocation reserves ten permits for browse and two for checkout.
It protects checkout progress in the declared slowdown, but it strands two
permits when X is healthy, twelve browse operations arrive, and checkout has no
work.

The passing replay does not prove that X failed, guarantee future SLO
compliance, or establish that a fixed allocation is the only legal correction.
