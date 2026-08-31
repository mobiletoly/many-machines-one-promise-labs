# Practice 01: From Evidence to Action

Checkout caller P requires one read-only shipping quote from dependency X.
Monitor M observes X through a separate heartbeat path. The retained evidence
contains three quote deadlines, three missing heartbeats, and detector output
that reports X as suspected.

The practice asks which execution claims those retained observations support
and which later shipping-quote interactions caller P should permit.

## Declared contract

After the policy suppresses ordinary shipping-quote calls:

- ordinary calls must fail without reaching X;
- each logical trial round may expose X to at most one read-only representative
  shipping-quote operation;
- a successful representative result within the 200 ms operation boundary may
  restore ordinary admission;
- if X becomes and remains capable while demand continues, P must restore
  ordinary admission by the end of the third replay round;
- the policy must not claim that X crashed, recovered, or can serve full load.

A successful `/healthz` call does not establish shipping-quote capability.
The policy controls only `checkout-p` and the `shipping_quote` operation.

## Prerequisites and platforms

Install Go 1.23 or newer. The complete verifier requires a POSIX shell and
supports macOS and Linux. Run every command from the repository root.

## Inspect the evidence

Read:

```text
practices/01-from-evidence-to-action/evidence.jsonl
```

Then inspect the reader-owned decision record:

```text
practices/01-from-evidence-to-action/starter/policy.json
```

The record separates an inference about X from the action P takes while
shipping-quote evidence remains incomplete.

## Commands

Replay the starter policy. The command must exit nonzero and report that the
initial evidence does not establish `crashed`:

```sh
GOWORK=off go run ./practices/01-from-evidence-to-action/cmd/replay \
  -policy practices/01-from-evidence-to-action/starter/policy.json
```

Create a reader-owned record outside the repository and edit it:

```sh
cp practices/01-from-evidence-to-action/starter/policy.json \
  /tmp/practice-01-policy.json
${EDITOR:-vi} /tmp/practice-01-policy.json
```

Verify that exact reader-owned record:

```sh
./scripts/verify-practice-01.sh /tmp/practice-01-policy.json
```

After your policy passes, compare it with the bundled reference record:

```sh
GOWORK=off go run ./practices/01-from-evidence-to-action/cmd/replay \
  -policy practices/01-from-evidence-to-action/solution/policy.json
```

Run the evidence and scope boundaries:

```sh
GOWORK=off go test -count=1 \
  ./practices/01-from-evidence-to-action/replay \
  -run '^TestBoundary'
```

Run the repository check for the bundled artifacts, manifest commands, tests,
race detector, and static analysis:

```sh
./scripts/check-practice-01.sh
```

## Boundary

The corrected policy admits one representative shipping quote while ordinary
calls remain suppressed. The admitted interaction can create evidence that a
fully suppressing policy would remove.

The result authorizes one caller-side policy transition under the declared
contract. It does not prove objective dependency health, future success, full
capacity, or a fleet-wide admission limit.
