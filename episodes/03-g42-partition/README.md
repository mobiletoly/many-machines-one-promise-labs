# G-42 Splits and Reunites

Episode 03 runs two independent replica processes. A controlled HTTP proxy
blocks state exchange while each process accepts a different `$80` redemption
against the same `$100` gift balance. The proxy then heals and the replicas
exchange their retained state.

## Property

After G-42 activity stops and both sync calls succeed:

```text
Replica A accounts for RA-80 and RB-80
Replica B accounts for RA-80 and RB-80
```

Each operation appears once by identity, and both replicas derive the same
confirmed value from the same accepted facts.

The `start` merge compares only the derived confirmed-value summaries. Both
replicas show `$80`, so each keeps its different local operation. The
`solution` merge preserves the union of accepted operations by identity.

## Requirements

- Go 1.23 or newer
- macOS or Linux
- no external services or containers

The parent test builds and starts two live child processes. Process readiness
comes from the address reported after each listener binds. A test-owned proxy
selects the partition and later forwards state exchange; no sleep selects the
execution.

## Commands

Run the one-replica path:

```sh
GOWORK=off go test -count=1 ./episodes/03-g42-partition/start -run '^TestOneReplicaRedeemsLocally$'
```

Reproduce the expected missing-history failure:

```sh
GOWORK=off go test -count=1 -tags=failure ./episodes/03-g42-partition/start -run '^TestReplicasPreserveAcceptedHistoryAfterPartition$'
```

The expected assertion contains:

```text
property violated: replica A accounts for [RA-80], want [RA-80 RB-80]
```

Run the same execution against the correction:

```sh
GOWORK=off go test -count=1 -tags=failure ./episodes/03-g42-partition/solution -run '^TestReplicasPreserveAcceptedHistoryAfterPartition$'
```

Attack the stronger business claim:

```sh
GOWORK=off go test -count=1 ./episodes/03-g42-partition/solution -run '^TestBoundaryConvergedHistoryStillViolatesFunding$'
```

The boundary test passes when the replicas agree on `$160 confirmed` against
`$100 funded`. Convergence preserves and aligns accepted history; it does not
repair the funded-value invariant.

Run the complete episode verifier:

```sh
./scripts/verify-episode-03.sh
```
