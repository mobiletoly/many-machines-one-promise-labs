package degradation

import (
	"context"
	"errors"
)

type catalogFixture struct {
	catalog Catalog
	err     error
}

func (f catalogFixture) Current(context.Context) (Catalog, error) {
	return f.catalog, f.err
}

type recommendationFixture struct {
	recommendation Recommendation
	err            error
	started        chan struct{}
	release        chan struct{}
}

func (f recommendationFixture) Recommend(ctx context.Context) (Recommendation, error) {
	if f.started != nil {
		close(f.started)
	}
	if f.release != nil {
		select {
		case <-f.release:
		case <-ctx.Done():
			return Recommendation{}, ctx.Err()
		}
	}
	return f.recommendation, f.err
}

type controlledBoundary struct {
	reached chan struct{}
	release chan struct{}
}

func (b controlledBoundary) Wait(ctx context.Context) error {
	close(b.reached)
	select {
	case <-b.release:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type neverBoundary struct{}

func (neverBoundary) Wait(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func currentCatalog() Catalog {
	return Catalog{
		Bottles: []string{"Cabernet", "Riesling"},
		Usable:  true,
	}
}

func recommendationUnavailable(err error) bool {
	return errors.Is(err, ErrRecommendationUnavailable)
}
