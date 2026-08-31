# Integration Capstone 01: Two Servers, One Operation

Two independent Go HTTP service processes accept create attempts through one
shared PostgreSQL database.

## Declared property

Attempts carrying one client-assigned `operation_id` may reach different live
service instances. Concurrent matching attempts must create at most one
canonical durable order and return the same canonical result. Reusing that
identity with different request semantics must return a conflict.

The start implementation uses one mutex per process and checks the database
before inserting. The database schema does not enforce operation identity.
The controlled execution lets both processes observe `op-61` as absent before
either process inserts it.

The solution gives PostgreSQL the shared arbitration boundary:

```sql
CONSTRAINT one_canonical_order_per_operation UNIQUE (operation_id)
```

The winning insert returns the new row. A matching attempt that loses the
uniqueness race issues a new read and returns the committed canonical row.

## Prerequisites and platforms

- Go 1.23 or newer;
- macOS or Linux;
- a POSIX shell for the verifier;
- Docker Engine with Docker Compose for the default database bootstrap, or a
  reachable PostgreSQL database supplied through `DATABASE_URL`;
- permission to create and drop test-owned schemas in that database.

The default Compose path uses PostgreSQL 17.10. Docker starts the database; the
Go harness creates one isolated schema per test, applies `schema.sql`, and
drops only that schema during cleanup. Docker entrypoint initialization does
not create the tables.

## Run the complete capstone

From the Labs repository root:

```sh
./scripts/verify-capstone-01.sh
```

When `DATABASE_URL` is unset, the verifier starts a project-scoped Compose
database on an ephemeral local port and removes its containers, network, and
volumes after the run.

To use an existing PostgreSQL database:

```sh
DATABASE_URL='postgres://user:password@127.0.0.1:5432/database?sslmode=disable' \
  ./scripts/verify-capstone-01.sh
```

## Run each state

These commands require `DATABASE_URL` in the environment. If you want the
repository-managed database, start it first:

```sh
./scripts/capstone-01-postgres.sh up
export DATABASE_URL=$(./scripts/capstone-01-postgres.sh url)
```

Run the working one-server path:

```sh
GOWORK=off go test -count=1 -tags=integration \
  ./capstones/01-two-servers-one-operation/start \
  -run '^TestOneServerCreatesOrder$'
```

Run the named failure:

```sh
GOWORK=off go test -count=1 -tags='integration failure' \
  ./capstones/01-two-servers-one-operation/start \
  -run '^TestTwoServersCreateOneOperation$'
```

It must fail with:

```text
property violated: canonical_orders(operation=op-61) = 2, want <= 1
```

Run the corrected execution:

```sh
GOWORK=off go test -count=1 -tags='integration failure' \
  ./capstones/01-two-servers-one-operation/solution \
  -run '^TestTwoServersCreateOneOperation$'
```

Run the process-replacement transfer:

```sh
GOWORK=off go test -count=1 -tags=integration \
  ./capstones/01-two-servers-one-operation/solution \
  -run '^TestReplacementReturnsCanonicalResult$'
```

Run the conflict and same-payload boundaries:

```sh
GOWORK=off go test -count=1 -tags=integration \
  ./capstones/01-two-servers-one-operation/solution \
  -run '^(TestConflictingReuseIsRejectedAcrossServers|TestBoundarySamePayloadDifferentOperationsCreateTwoOrders)$'
```

When you started the managed database yourself, remove it with:

```sh
./scripts/capstone-01-postgres.sh down
```

## Boundary

The database constraint is scoped to one logical operation identity. Two
intentional operations with identical `Cabernet x1` payloads use `op-61` and
`op-62`; both must create an order.

The capstone does not establish one HTTP attempt, one downstream dispatch,
database availability, regional failover, or exactly-once execution.
