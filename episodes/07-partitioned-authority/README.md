# Both Booths See Four Left

Episode 07 gives two ticket booths the same last observed count for event E-8.
The booths remain healthy while communication with the event authority is
unavailable. Each booth can still handle local requests.

The start system treats `remaining = 4` as permission at both booths. Each
accepts three distinct sales. The correction allocates two non-overlapping
selling rights to each booth before communication stops.

## Property

After the controlled local activity stops, the event must satisfy:

```text
confirmed sales
    + outstanding spendable rights
    + unallocated reserve
    <= event capacity
```

The property also requires both healthy booths to accept their two locally
authorized sales without consulting the other booth or the event authority.
The progress check prevents a reject-all implementation from passing the
safety assertion.

## Requirements

- Go 1.23 or newer
- macOS or Linux
- no external services or containers

The lab keeps two independently locked booth stores in one process. The test
harness inspects their combined state after activity stops. Process boundaries
would add machinery without changing the authority decision exercised here.

## Commands

Run one booth's existing local path:

```sh
GOWORK=off go test -count=1 ./episodes/07-partitioned-authority/start -run '^TestOneBoothConfirmsFromLocalState$'
```

Reproduce the capacity failure:

```sh
GOWORK=off go test -count=1 -tags=failure ./episodes/07-partitioned-authority/start -run '^TestLocalAuthorityPreservesCapacityDuringPartition$'
```

The expected assertion contains:

```text
property violated: event E-8 exposure = confirmed 9 + outstanding 2 + reserve 4 = 15, capacity 7
```

Each booth still holds one observed count that its local code can spend, so the
accounting includes two outstanding permissions as well as the four units that
the event authority still calls reserve.

Run the same local requests against the allocated-rights correction:

```sh
GOWORK=off go test -count=1 -tags=failure ./episodes/07-partitioned-authority/solution -run '^TestLocalAuthorityPreservesCapacityDuringPartition$'
```

Attack the stronger utilization claim:

```sh
GOWORK=off go test -count=1 ./episodes/07-partitioned-authority/solution -run '^TestBoundarySafeAllocationCanStrandCapacity$'
```

The boundary test gives Booth A one right and Booth B three. After Booth A
uses its right, another local request receives `no local authority` while
Booth B still holds three usable rights. Capacity remains safe and locally
unavailable where demand arrived.

Run the complete verifier:

```sh
./scripts/verify-episode-07.sh
```
