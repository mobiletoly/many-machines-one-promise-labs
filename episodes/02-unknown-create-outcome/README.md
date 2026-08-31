# The Commit Happened, the Response Did Not

Episode 02 sends one create operation across an HTTP boundary. The service
commits an accepted order, then a deterministic fault closes the connection
before the client receives the response.

## Property

For one running service process at a time and one retained store:

```text
matching attempts carrying one operation identity
    -> at most one canonical accepted order
    -> one retained acceptance result
```

The `start` state commits every HTTP attempt as a new order. The `solution`
state retains the operation identity, accepted semantics, and canonical result
in one synced record.

## Requirements

- Go 1.23 or newer
- macOS or Linux
- no external services

The episode uses an append-only file and scopes durability to service-process
restart after a completed `Sync`. It does not simulate a torn write, storage
failure, machine loss, or two service processes writing concurrently.

## Commands

Run the working one-attempt path:

```sh
GOWORK=off go test -count=1 ./episodes/02-unknown-create-outcome/start -run '^TestCreateOrderHappyPath$'
```

Reproduce the expected duplicate-order property failure:

```sh
GOWORK=off go test -count=1 -tags=failure ./episodes/02-unknown-create-outcome/start -run '^TestRetryAfterCommittedResponseIsLost$'
```

The expected assertion contains:

```text
property violated: canonical_orders(operation=op-61) = 2, want <= 1
```

Run the same execution against the correction:

```sh
GOWORK=off go test -count=1 -tags=failure ./episodes/02-unknown-create-outcome/solution -run '^TestRetryAfterCommittedResponseIsLost$'
```

Check conflicting reuse, concurrent matching attempts, and the downstream
dispatch boundary:

```sh
GOWORK=off go test -count=1 ./episodes/02-unknown-create-outcome/solution
```

Run the complete episode verifier:

```sh
./scripts/verify-episode-02.sh
```
