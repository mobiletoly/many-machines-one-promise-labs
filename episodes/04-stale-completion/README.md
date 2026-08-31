# Worker A Comes Back with Order 73

Episode 04 tests explicit responsibility transfer while a former owner may
still be executing.

```text
worker-a owns assignment 17
    -> Order 73 transfers to worker-b under assignment 18
    -> worker-a submits completion under assignment 17
```

The start state accepts the stale canonical completion. The solution compares
the submitted owner and assignment version with current order state inside the
completion transition.

## Declared property

After responsibility for Order 73 transfers from worker-a under assignment 17
to worker-b under assignment 18, the order service must reject worker-a's
canonical completion under assignment 17 and retain worker-b under assignment
18 as the current owner.

## Prerequisites and platforms

Install Go 1.23 or newer. The complete verifier requires a POSIX shell and
supports macOS and Linux. Run all commands from the repository root.

## Commands

Run the current-owner happy path:

```sh
GOWORK=off go test -count=1 ./episodes/04-stale-completion/start -run '^TestCurrentOwnerCompletesOrder73$'
```

Reproduce the property violation. This command must exit nonzero and print the
named `completion(order=73, worker=worker-a, assignment=17) = prepared, want stale_assignment`
assertion:

```sh
GOWORK=off go test -count=1 -tags=failure ./episodes/04-stale-completion/start -run '^TestFormerOwnerCannotCompleteAfterTransfer$'
```

Run the corrected execution:

```sh
GOWORK=off go test -count=1 -tags=failure ./episodes/04-stale-completion/solution -run '^TestFormerOwnerCannotCompleteAfterTransfer$'
```

Run the boundary test:

```sh
GOWORK=off go test -count=1 ./episodes/04-stale-completion/solution -run '^TestBoundaryStaleRejectionLeavesPreparedDrink$'
```

Run the complete episode verifier from the repository root:

```sh
./scripts/verify-episode-04.sh
```

The verifier expects the start failure, proves the solution guarantee, and
runs a boundary test showing that canonical rejection does not remove a
Cabernet prepared outside the order service.

This episode uses one Go process and in-memory state. It does not implement
expiry, cancellation, leases, downstream fencing, durable recovery, or
exactly-once execution.
