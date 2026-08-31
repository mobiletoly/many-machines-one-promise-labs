# Two Workers Claim Order 42

Episode 01 uses one Go process, two workers, and one in-memory order store.

## Property

For one order in one store instance:

```text
successful_claims(order_id) <= 1
```

Order 42 starts available and both local calls complete, so one worker must
receive `claimed` in the declared execution.

The `start` state protects individual map operations but separates the
availability check from the claimant update. The failure command places two
workers after the check and releases them together.

The `solution` state performs the check and update under one store lock.

## Requirements

- Go 1.23 or newer
- macOS or Linux
- no external services

## Commands

Run the working one-worker path:

```sh
GOWORK=off go test -count=1 ./episodes/01-concurrent-claim/start -run '^TestOneWorkerClaimsOrder$'
```

Reproduce the expected property failure:

```sh
GOWORK=off go test -count=1 -tags=failure ./episodes/01-concurrent-claim/start -run '^TestTwoWorkersClaimOrder42$'
```

The command exits nonzero because both workers receive `claimed`. The expected
assertion contains:

```text
property violated: successful_claims(order=42) = 2, want <= 1
```

Run the same property test against the correction:

```sh
GOWORK=off go test -count=1 -tags=failure ./episodes/01-concurrent-claim/solution -run '^TestTwoWorkersClaimOrder42$'
```

Attack the remaining boundary:

```sh
GOWORK=off go test -count=1 ./episodes/01-concurrent-claim/solution -run '^TestBoundaryClaimCanStrandOrder$'
```

The boundary test confirms that one worker can claim the order and disappear.
The atomic claim prevents a second claimant; it does not make the first worker
finish or transfer ownership.

Run the complete episode verifier:

```sh
./scripts/verify-episode-01.sh
```

The companion chapter has not been published.
