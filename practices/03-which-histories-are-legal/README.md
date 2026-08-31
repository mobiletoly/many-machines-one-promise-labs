# Practice 03: Which Histories Are Legal?

Eight operation histories mix session-visible obligations, one
cross-operation dependency, and completed operations on a register-like
object. You will submit one review artifact with evidence citations,
state-coverage claims, forced edges, sequential witnesses, and contradictions.

## Declared property

Each submitted conclusion must follow from the declared contract and retained
history:

- a visibility review identifies the accepted facts that a successful result
  must account for;
- an object review supplies one legal sequential witness or a contradiction
  that the verifier can confirm by exhausting the finite witness space.

A trace path receives dependency credit only when an X-derived fact
participates in Y's declared production or acceptance rule. Earlier timestamps
do not create that relationship.

## Prerequisites and platforms

Install Go 1.23 or newer. The complete verifier requires a POSIX shell and
supports macOS and Linux. Run commands from the repository root.

## Inspect the raw input

Read these files before running the starter:

```text
practices/03-which-histories-are-legal/contract.json
practices/03-which-histories-are-legal/histories.json
practices/03-which-histories-are-legal/starter/review.json
```

Do not inspect `solution/review.json` yet.

For each visibility history, identify the source and scope of every required
fact. For each object history, derive the forced non-overlap edges before you
try a sequential order.

## Commands

Run the starter review. It must fail without printing the reference verdicts:

```sh
GOWORK=off go run ./practices/03-which-histories-are-legal/cmd/replay -review practices/03-which-histories-are-legal/starter/review.json
```

Copy the starter artifact and edit your copy:

```sh
cp practices/03-which-histories-are-legal/starter/review.json \
  /tmp/practice-03-review.json
${EDITOR:-vi} /tmp/practice-03-review.json
```

Verify your review:

```sh
./scripts/verify-practice-03.sh /tmp/practice-03-review.json
```

After it passes, compare it with the bundled reference:

```sh
GOWORK=off go run ./practices/03-which-histories-are-legal/cmd/replay -review practices/03-which-histories-are-legal/solution/review.json
```

Run the scope and stronger-claim boundaries:

```sh
GOWORK=off go test -count=1 ./practices/03-which-histories-are-legal/replay -run '^TestBoundary'
```

Run the repository check:

```sh
./scripts/check-practice-03.sh
```

## Boundary

A passing session-visible proof does not establish object-wide
linearizability. A legal object history does not establish a session guarantee.
One scoped X-to-Y dependency does not establish general causal consistency,
and one legal history promises no future availability.
