package degradation

import (
	"context"
	"errors"
)

var (
	ErrCatalogUnavailable        = errors.New("catalog unavailable")
	ErrRecommendationUnavailable = errors.New("recommendation unavailable")
)

type ResultVariant string

const (
	FullResult    ResultVariant = "full"
	ReducedResult ResultVariant = "reduced"
)

type RecommendationStatus string

const (
	RecommendationAvailable   RecommendationStatus = "available"
	RecommendationUnavailable RecommendationStatus = "unavailable_for_response"
)

type Catalog struct {
	Bottles []string
	Usable  bool
}

type Recommendation struct {
	Text string
}

type BrowseResult struct {
	Variant              ResultVariant
	Catalog              Catalog
	RecommendationStatus RecommendationStatus
	Recommendation       *Recommendation
}

type CatalogSource interface {
	Current(context.Context) (Catalog, error)
}

type RecommendationSource interface {
	Recommend(context.Context) (Recommendation, error)
}

type SelectionBoundary interface {
	Wait(context.Context) error
}

type Service struct {
	Catalog         CatalogSource
	Recommendations RecommendationSource
	Boundary        SelectionBoundary
}

type recommendationOutcome struct {
	recommendation Recommendation
	err            error
}

func (s Service) Browse(ctx context.Context) (BrowseResult, error) {
	operationContext, cancel := context.WithCancel(ctx)
	defer cancel()

	recommendationResults := make(chan recommendationOutcome, 1)
	go func() {
		recommendation, err := s.Recommendations.Recommend(operationContext)
		recommendationResults <- recommendationOutcome{
			recommendation: recommendation,
			err:            err,
		}
	}()

	catalog, err := s.Catalog.Current(operationContext)
	if err != nil || !catalog.Usable {
		return BrowseResult{}, ErrCatalogUnavailable
	}

	selection := make(chan error, 1)
	go func() {
		selection <- s.Boundary.Wait(operationContext)
	}()

	return chooseBrowseResult(ctx, catalog, recommendationResults, selection)
}

func chooseBrowseResult(
	ctx context.Context,
	catalog Catalog,
	recommendationResults <-chan recommendationOutcome,
	selection <-chan error,
) (BrowseResult, error) {
	for {
		select {
		case outcome := <-recommendationResults:
			if outcome.err != nil {
				recommendationResults = nil
				continue
			}
			return fullBrowseResult(catalog, outcome.recommendation), nil
		case err := <-selection:
			if err != nil {
				return BrowseResult{}, err
			}
			select {
			case outcome := <-recommendationResults:
				if outcome.err == nil {
					return fullBrowseResult(catalog, outcome.recommendation), nil
				}
			default:
			}
			return BrowseResult{}, ErrRecommendationUnavailable
		case <-ctx.Done():
			return BrowseResult{}, ctx.Err()
		}
	}
}

func fullBrowseResult(catalog Catalog, recommendation Recommendation) BrowseResult {
	return BrowseResult{
		Variant:              FullResult,
		Catalog:              catalog,
		RecommendationStatus: RecommendationAvailable,
		Recommendation:       &recommendation,
	}
}
