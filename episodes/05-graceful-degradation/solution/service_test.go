package degradation

import (
	"context"
	"testing"
)

func TestBrowseReturnsFullResultWhenRecommendationArrives(t *testing.T) {
	wantRecommendation := Recommendation{Text: "Cabernet with braised beef"}
	service := Service{
		Catalog: catalogFixture{catalog: currentCatalog()},
		Recommendations: recommendationFixture{
			recommendation: wantRecommendation,
		},
		Boundary: neverBoundary{},
	}

	result, err := service.Browse(context.Background())
	if err != nil {
		t.Fatalf("browse with recommendation: %v", err)
	}
	if result.Variant != FullResult {
		t.Fatalf("result variant = %q, want %q", result.Variant, FullResult)
	}
	if result.RecommendationStatus != RecommendationAvailable {
		t.Fatalf("recommendation status = %q, want %q", result.RecommendationStatus, RecommendationAvailable)
	}
	if result.Recommendation == nil || *result.Recommendation != wantRecommendation {
		t.Fatalf("recommendation = %+v, want %+v", result.Recommendation, wantRecommendation)
	}
}

func TestRecordedRecommendationWinsWhenBoundaryIsAlsoReady(t *testing.T) {
	wantRecommendation := Recommendation{Text: "Cabernet with braised beef"}
	recommendationResults := make(chan recommendationOutcome, 1)
	selection := make(chan error, 1)

	recommendationResults <- recommendationOutcome{recommendation: wantRecommendation}
	selection <- nil

	result, err := chooseBrowseResult(
		context.Background(),
		currentCatalog(),
		recommendationResults,
		selection,
	)
	if err != nil {
		t.Fatalf("select recorded recommendation: %v", err)
	}
	if result.Variant != FullResult {
		t.Fatalf("result variant = %q, want %q", result.Variant, FullResult)
	}
	if result.Recommendation == nil || *result.Recommendation != wantRecommendation {
		t.Fatalf("recommendation = %+v, want %+v", result.Recommendation, wantRecommendation)
	}
}
