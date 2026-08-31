# The List Without Advice

Episode 05 tests a browse contract with a required catalog and an enriching
recommendation.

```text
usable catalog + usable recommendation
    -> full success

usable catalog + no usable recommendation at selection
    -> reduced success
```

The start state fails the complete browse operation at the selection boundary.
The solution returns the reduced result that the product contract already
permits. It marks the recommendation unavailable for this response.

## Declared property

At the controlled selection boundary, a usable catalog without a usable
recommendation must produce reduced success with the current catalog and mark
the recommendation unavailable for that response.

## Prerequisites and platforms

Install Go 1.23 or newer. The complete verifier requires a POSIX shell and
supports macOS and Linux. Run all commands from the repository root.

## Commands

Run the full-result happy path:

```sh
GOWORK=off go test -count=1 ./episodes/05-graceful-degradation/start -run '^TestBrowseReturnsFullResultWhenRecommendationArrives$'
```

Reproduce the property violation. This command must exit nonzero and print the
named `browse result = error recommendation unavailable, want reduced success with current catalog`
assertion:

```sh
GOWORK=off go test -count=1 -tags=failure ./episodes/05-graceful-degradation/start -run '^TestBrowsePreservesCatalogAtSelectionBoundary$'
```

Run the corrected execution:

```sh
GOWORK=off go test -count=1 -tags=failure ./episodes/05-graceful-degradation/solution -run '^TestBrowsePreservesCatalogAtSelectionBoundary$'
```

Run the boundary test:

```sh
GOWORK=off go test -count=1 ./episodes/05-graceful-degradation/solution -run '^TestBoundaryRecommendationCannotReplaceMissingCatalog$'
```

Run the complete episode verifier from the repository root:

```sh
./scripts/verify-episode-05.sh
```

The verifier uses channels to hold the recommendation operation past a
controlled selection boundary. It needs no clock sleeps or external service.
It also places a usable recommendation and the boundary in the ready state
before selection, then proves that the recorded recommendation wins.
The boundary test removes the catalog and confirms that recommendation-only
content cannot satisfy the core browse contract.

The catalog fixture remains authority-valid until `Browse` returns. This
episode does not implement remote health detection, retries, circuit breaking,
fallback content, cached recommendations, or a fleet-wide degraded mode.
