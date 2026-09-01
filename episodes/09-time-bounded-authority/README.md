# The Desk Read 110

Episode 09 gives publication desk S one established lease for coordinator A
and manifest M-88:

```text
L-88 = [100, 110) on S's clock
```

The issuer becomes unavailable after S establishes the lease. S must accept a
valid publication before tick 110 and reject a new publication at or after
tick 110. S evaluates the lease at the same boundary that accepts the official
manifest effect.

The start state checks L-88 before a controlled acceptance boundary. The test
advances S's manual clock from 109 to 110 at that boundary. The start state
then accepts the late effect. The solution reads S's clock and accepts or
rejects the effect in one desk transition.

## Prerequisites and platforms

Install Go 1.23 or newer. The complete verifier requires a POSIX shell and
supports macOS and Linux. Run all commands from the repository root.

## Commands

Run the pre-expiry path while the issuer remains unavailable:

```sh
GOWORK=off go test -count=1 ./episodes/09-time-bounded-authority/start -run '^TestEstablishedLeaseSupportsPublicationWithoutIssuer$'
```

Reproduce the property violation. This command must exit nonzero and print the
named `accepted, want lease_expired` assertion:

```sh
GOWORK=off go test -count=1 -tags=failure ./episodes/09-time-bounded-authority/start -run '^TestLeaseExpiryIsEnforcedAtEffectAcceptance$'
```

Run the same execution against the solution:

```sh
GOWORK=off go test -count=1 -tags=failure ./episodes/09-time-bounded-authority/solution -run '^TestLeaseExpiryIsEnforcedAtEffectAcceptance$'
```

Run the exact-expiry and accepted-history boundaries:

```sh
GOWORK=off go test -count=1 ./episodes/09-time-bounded-authority/solution -run '^TestBoundary'
```

Run the complete verifier:

```sh
./scripts/verify-episode-09.sh
```

The manual clock chooses S-local ordering. The lab makes no claim that ten
ticks equal ten seconds of physical time. It covers no renewal, recovery,
holder replacement, fencing, transfer, failure detection, clock
synchronization, retry, or idempotency contract.
