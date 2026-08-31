//go:build failure

package degradation

import (
	"context"
	"fmt"
	"testing"
)

func TestBrowsePreservesCatalogAtSelectionBoundary(t *testing.T) {
	recommendationStarted := make(chan struct{})
	releaseRecommendation := make(chan struct{})
	boundaryReached := make(chan struct{})
	releaseBoundary := make(chan struct{})

	service := Service{
		Catalog: catalogFixture{catalog: currentCatalog()},
		Recommendations: recommendationFixture{
			started: recommendationStarted,
			release: releaseRecommendation,
		},
		Boundary: controlledBoundary{
			reached: boundaryReached,
			release: releaseBoundary,
		},
	}

	results := make(chan BrowseResult, 1)
	errors := make(chan error, 1)
	go func() {
		result, err := service.Browse(context.Background())
		results <- result
		errors <- err
	}()

	<-recommendationStarted
	<-boundaryReached
	close(releaseBoundary)

	result := <-results
	err := <-errors
	if err != nil || result.Variant != ReducedResult {
		observed := fmt.Sprintf("variant %q", result.Variant)
		if recommendationUnavailable(err) {
			observed = "error recommendation unavailable"
		} else if err != nil {
			observed = "error " + err.Error()
		}
		t.Fatalf(
			"property violated: browse result = %s, want reduced success with current catalog",
			observed,
		)
	}
	if !result.Catalog.Usable || len(result.Catalog.Bottles) != 2 {
		t.Fatalf("reduced catalog = %+v, want current catalog", result.Catalog)
	}
	if result.RecommendationStatus != RecommendationUnavailable || result.Recommendation != nil {
		t.Fatalf(
			"reduced recommendation = %q, %+v, want unavailable with no value",
			result.RecommendationStatus,
			result.Recommendation,
		)
	}
}
