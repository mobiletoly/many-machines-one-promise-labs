package degradation

import (
	"context"
	"errors"
	"testing"
)

func TestBoundaryRecommendationCannotReplaceMissingCatalog(t *testing.T) {
	service := Service{
		Catalog: catalogFixture{
			catalog: Catalog{Usable: false},
		},
		Recommendations: recommendationFixture{
			recommendation: Recommendation{Text: "Cabernet with braised beef"},
		},
		Boundary: neverBoundary{},
	}

	result, err := service.Browse(context.Background())
	if !errors.Is(err, ErrCatalogUnavailable) {
		t.Fatalf("browse error = %v, want %v", err, ErrCatalogUnavailable)
	}
	if result.Variant != "" || result.Catalog.Usable || len(result.Catalog.Bottles) != 0 ||
		result.RecommendationStatus != "" || result.Recommendation != nil {
		t.Fatalf("browse result = %+v, want no successful result", result)
	}
}
