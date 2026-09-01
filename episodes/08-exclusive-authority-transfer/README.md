# Booth A Has It, Booth B Still Does

Episode 08 starts after Episode 07 strands unused selling rights at Booth B
while customers wait at Booth A. Booth B owns `R-100`. The booths can exchange
intermittent messages, but they cannot change both local authority stores in
one atomic transition.

The start system issues an actionable `X-100` grant before Booth B relinquishes
`R-100`. Booth A accepts the grant and confirms one sale. Booth B can still
confirm a different sale with the same right.

The correction makes grant issuance and source relinquishment one local atomic
transition. Booth A accepts the retained grant once. A matching delivery after
acknowledgement loss returns the retained acceptance and cannot restore a
consumed right.

## Property

One exclusive right must never authorize retained effects at both booths:

```text
usable_by_A(R-100) + usable_by_B(R-100) <= 1
```

The executable oracle materializes the failure as two distinct confirmations
that both cite `R-100`.

The property also requires scoped progress. Once Booth B has retained the
relinquished `X-100` transfer and one valid grant reaches healthy Booth A,
processing that grant must retain one acceptance and let Booth A confirm one
sale with `R-100`.

## Requirements

- Go 1.23 or newer
- macOS or Linux
- no external services or containers

The lab keeps two separately locked booth stores in one process. The verifier
selects the message schedule. Process boundaries would add transport machinery
without changing the two local authority decisions exercised here.

## Commands

Run the plausible grant-first path that reaches one final holder:

```sh
GOWORK=off go test -count=1 ./episodes/08-exclusive-authority-transfer/start -run '^TestGrantFirstWorkflowCanReachOneFinalHolder$'
```

Reproduce the authority overlap before Booth B relinquishes the right:

```sh
GOWORK=off go test -count=1 -tags=failure ./episodes/08-exclusive-authority-transfer/start -run '^TestExclusiveAuthoritySurvivesGrantDelivery$'
```

The expected assertion contains:

```text
exclusive authority violated: right R-100 confirmed by A-301 at booth-a and B-401 at booth-b during X-100
```

Run the same lost-attempt, retry, duplicate-delivery, and sale history against
the relinquish-first correction:

```sh
GOWORK=off go test -count=1 -tags=failure ./episodes/08-exclusive-authority-transfer/solution -run '^TestExclusiveAuthoritySurvivesGrantDelivery$'
```

Attack the stronger availability claim by losing the grant after Booth B has
relinquished `R-100`:

```sh
GOWORK=off go test -count=1 ./episodes/08-exclusive-authority-transfer/solution -run '^TestBoundaryLostGrantLeavesAuthorityGap$'
```

Both booths reject a sale with `R-100` until a valid matching grant reaches
Booth A. The right still exists. Neither booth can use it.

Run the complete verifier:

```sh
./scripts/verify-episode-08.sh
```
