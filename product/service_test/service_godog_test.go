package service_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"testing"

	"github.com/cucumber/godog"
	"github.com/siderustler/go-ecommerce/product"
)

type productRepositoryMock struct {
	productsFn      func(ctx context.Context, offset int, limit int, filter product.Filter) ([]product.Product, int, error)
	productsByIDsFn func(ctx context.Context, ids []string) (map[string]product.Product, error)
	promotionsFn    func(ctx context.Context, offset int, limit int) ([]product.Product, int, error)
}

func (m productRepositoryMock) Products(ctx context.Context, offset int, limit int, filter product.Filter) ([]product.Product, int, error) {
	if m.productsFn == nil {
		return nil, 0, fmt.Errorf("Products not configured")
	}
	return m.productsFn(ctx, offset, limit, filter)
}

func (m productRepositoryMock) ProductsByIDs(ctx context.Context, ids []string) (map[string]product.Product, error) {
	if m.productsByIDsFn == nil {
		return nil, fmt.Errorf("ProductsByIDs not configured")
	}
	return m.productsByIDsFn(ctx, ids)
}

func (m productRepositoryMock) Promotions(ctx context.Context, offset int, limit int) ([]product.Product, int, error) {
	if m.promotionsFn == nil {
		return nil, 0, fmt.Errorf("Promotions not configured")
	}
	return m.promotionsFn(ctx, offset, limit)
}

type paginationFeatureState struct {
	services          *product.Services
	productsQuery     product.ProductsQuery
	promotionsQuery   product.PromotionsQuery
	recordedOffset    int
	recordedLimit     int
	normalizedPage    int
	recordedQueryName string
}

func newPaginationFeatureState() *paginationFeatureState {
	state := &paginationFeatureState{}
	state.reset()
	return state
}

func (s *paginationFeatureState) reset() {
	s.recordedOffset = 0
	s.recordedLimit = 0
	s.normalizedPage = 0
	s.recordedQueryName = ""
	s.productsQuery = product.ProductsQuery{}
	s.promotionsQuery = product.PromotionsQuery{}

	repo := productRepositoryMock{
		productsFn: func(ctx context.Context, offset int, limit int, filter product.Filter) ([]product.Product, int, error) {
			s.recordedQueryName = "products"
			s.recordedOffset = offset
			s.recordedLimit = limit
			s.normalizedPage = s.productsQuery.Page
			return nil, 0, nil
		},
		productsByIDsFn: func(ctx context.Context, ids []string) (map[string]product.Product, error) {
			return map[string]product.Product{}, nil
		},
		promotionsFn: func(ctx context.Context, offset int, limit int) ([]product.Product, int, error) {
			s.recordedQueryName = "promotions"
			s.recordedOffset = offset
			s.recordedLimit = limit
			s.normalizedPage = 1 + (offset / limit)
			return nil, 0, nil
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	s.services = product.NewServices(repo, logger)
}

func (s *paginationFeatureState) givenProductsQuery(page int, limit int) error {
	s.productsQuery = product.ProductsQuery{
		Page:  page,
		Limit: limit,
	}
	return nil
}

func (s *paginationFeatureState) givenPromotionsQuery(page int, pageSize int) error {
	s.promotionsQuery = product.PromotionsQuery{
		Page:     page,
		PageSize: pageSize,
	}
	return nil
}

func (s *paginationFeatureState) whenProductServicePreparesTheRepositoryRequest() error {
	switch {
	case s.productsQuery.Limit != 0:
		_, err := s.services.Query.Products.Handle(context.Background(), s.productsQuery)
		return err
	case s.promotionsQuery.PageSize != 0:
		_, err := s.services.Query.Promotions.Handle(context.Background(), s.promotionsQuery)
		return err
	default:
		return fmt.Errorf("no query configured")
	}
}

func (s *paginationFeatureState) thenRepositoryOffsetShouldBe(expectedOffset int) error {
	if s.recordedOffset != expectedOffset {
		return fmt.Errorf("expected repository offset %d, got %d", expectedOffset, s.recordedOffset)
	}
	return nil
}

func (s *paginationFeatureState) thenPageShouldBeAssignedTo(expectedPage int) error {
	if s.recordedQueryName != "promotions" {
		return fmt.Errorf("page normalization assertions are only supported for promotions queries")
	}
	if s.normalizedPage != expectedPage {
		return fmt.Errorf("expected normalized page %d, got %d", expectedPage, s.normalizedPage)
	}
	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	state := newPaginationFeatureState()

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		state.reset()
		return ctx, nil
	})

	ctx.Step(`^a products query with page (-?\d+) and limit (\d+)$`, state.givenProductsQuery)
	ctx.Step(`^a promotions query with page (-?\d+) and page size (\d+)$`, state.givenPromotionsQuery)
	ctx.Step(`^the product service prepares the repository request$`, state.whenProductServicePreparesTheRepositoryRequest)
	ctx.Step(`^the repository offset should be (\d+)$`, state.thenRepositoryOffsetShouldBe)
	ctx.Step(`^the page should be assigned to (\d+)$`, state.thenPageShouldBeAssignedTo)
}

func TestProductServiceFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run product service feature tests")
	}
}
