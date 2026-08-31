# Practice 04: When Is Recovery Complete?

One recovery dossier combines an authoritative accepted history, complete
membership evidence for three reconstructed states, and a serving timeline.
You will derive the latest complete prefix, preserve one independently covered
acknowledgement, and identify the first evidence-supported boundary at which
the named capability can serve a contract-valid recovered result.

## Declared property

The recovered state is admissible only when:

- its proved complete prefix satisfies the declared RPO;
- it preserves every operation in the separately protected acknowledgement
  class.

RTO ends at the earliest evidence-supported boundary where admissible state,
valid interpretation, and an eligible path let the named capability serve its
declared result. A later customer request does not move that boundary.

## Prerequisites and platforms

Install Go 1.23 or newer. The complete verifier requires a POSIX shell and
supports macOS and Linux. Run commands from the repository root.

## Inspect the dossier

Read:

```text
practices/04-when-is-recovery-complete/contract.json
practices/04-when-is-recovery-complete/accepted-history.jsonl
practices/04-when-is-recovery-complete/recovery-evidence.jsonl
practices/04-when-is-recovery-complete/serving-events.jsonl
practices/04-when-is-recovery-complete/starter/recovery-decision.json
```

Do not inspect `solution/recovery-decision.json` yet.

## Commands

Run the starter decision. It must fail without printing the reference cut or
RTO boundary:

```sh
GOWORK=off go run ./practices/04-when-is-recovery-complete/cmd/replay -decision practices/04-when-is-recovery-complete/starter/recovery-decision.json
```

Copy the starter and edit your review:

```sh
cp practices/04-when-is-recovery-complete/starter/recovery-decision.json \
  /tmp/practice-04-recovery-decision.json
${EDITOR:-vi} /tmp/practice-04-recovery-decision.json
```

Verify your artifact:

```sh
./scripts/verify-practice-04.sh /tmp/practice-04-recovery-decision.json
```

After it passes, compare it with the bundled reference:

```sh
GOWORK=off go run ./practices/04-when-is-recovery-complete/cmd/replay -decision practices/04-when-is-recovery-complete/solution/recovery-decision.json
```

Run the scope boundaries:

```sh
GOWORK=off go test -count=1 ./practices/04-when-is-recovery-complete/replay -run '^TestBoundary'
```

Run the repository check:

```sh
./scripts/check-practice-04.sh
```

## Boundary

A passing dossier establishes one named capability under one declared
disruption and evidence set. It does not establish recovery for another
capability, prove future disruptions meet the objectives, select a recovery
mechanism, or make the first customer request the universal RTO boundary.
