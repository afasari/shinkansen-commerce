package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	productpb "github.com/afasari/shinkansen-commerce/gen/proto/go/product"
	sharedpb "github.com/afasari/shinkansen-commerce/gen/proto/go/shared"
	"github.com/afasari/shinkansen-commerce/services/product-service/internal/cache"
	"github.com/afasari/shinkansen-commerce/services/product-service/internal/db"
	"github.com/afasari/shinkansen-commerce/services/product-service/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func setupIntegrationTest(t *testing.T) (*service.ProductService, func()) {
	logger := zap.NewNop()
	defer func() { _ = logger.Sync() }()

	dbpool, err := db.New("postgres://shinkansen:shinkansen_dev_password@localhost:5432/shinkansen?sslmode=disable")
	require.NoError(t, err, "failed to connect to database")

	queries := db.NewQueries(dbpool)

	redisClient := cache.NewRedisClient("localhost:6379")
	cacheClient := cache.NewRedisCache(redisClient)

	productService := service.NewProductService(queries, cacheClient, logger)

	cleanup := func() {
		dbpool.Close()
		_ = redisClient.Close()
	}

	return productService, cleanup
}

func TestIntegration_CreateAndGetProduct(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	service, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	createReq := &productpb.CreateProductRequest{
		Name:        "Integration Test Product",
		Description: "A product created by integration tests",
		Sku:         "INT-TEST-001",
		Price: &sharedpb.Money{
			Units:    1999,
			Currency: "JPY",
		},
		StockQuantity: 100,
	}

	createResp, err := service.CreateProduct(ctx, createReq)
	require.NoError(t, err)
	assert.NotEmpty(t, createResp.ProductId)

	getReq := &productpb.GetProductRequest{
		ProductId: createResp.ProductId,
	}

	getResp, err := service.GetProduct(ctx, getReq)
	require.NoError(t, err)
	assert.NotNil(t, getResp.Product)
	assert.Equal(t, createReq.Name, getResp.Product.Name)
	assert.Equal(t, createReq.Description, getResp.Product.Description)
	assert.Equal(t, createReq.Sku, getResp.Product.Sku)
	assert.Equal(t, createReq.Price.Units, getResp.Product.Price.Units)
	assert.Equal(t, createReq.Price.Currency, getResp.Product.Price.Currency)
	assert.Equal(t, createReq.StockQuantity, getResp.Product.StockQuantity)
	assert.Equal(t, createResp.ProductId, getResp.Product.Id)
}

func TestIntegration_UpdateProduct(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	service, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	createReq := &productpb.CreateProductRequest{
		Name:        "Product to Update",
		Description: "Original description",
		Sku:         "UPDATE-TEST-001",
		Price: &sharedpb.Money{
			Units:    1000,
			Currency: "JPY",
		},
		StockQuantity: 50,
	}

	createResp, err := service.CreateProduct(ctx, createReq)
	require.NoError(t, err)

	updateReq := &productpb.UpdateProductRequest{
		ProductId: createResp.ProductId,
		Name:      &wrapperspb.StringValue{Value: "Updated Product Name"},
		Price: &sharedpb.Money{
			Units:    1500,
			Currency: "JPY",
		},
		Active: &wrapperspb.BoolValue{Value: false},
	}

	updateResp, err := service.UpdateProduct(ctx, updateReq)
	require.NoError(t, err)
	assert.NotNil(t, updateResp.Product)
	assert.Equal(t, "Updated Product Name", updateResp.Product.Name)
	assert.Equal(t, int64(1500), updateResp.Product.Price.Units)
	assert.Equal(t, false, updateResp.Product.Active)

	getReq := &productpb.GetProductRequest{
		ProductId: createResp.ProductId,
	}

	getResp, err := service.GetProduct(ctx, getReq)
	require.NoError(t, err)
	assert.Equal(t, "Updated Product Name", getResp.Product.Name)
	assert.Equal(t, int64(1500), getResp.Product.Price.Units)
	assert.Equal(t, false, getResp.Product.Active)
}

func TestIntegration_DeleteProduct(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	service, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	createReq := &productpb.CreateProductRequest{
		Name:        "Product to Delete",
		Description: "This product will be deleted",
		Sku:         "DELETE-TEST-001",
		Price: &sharedpb.Money{
			Units:    500,
			Currency: "JPY",
		},
		StockQuantity: 25,
	}

	createResp, err := service.CreateProduct(ctx, createReq)
	require.NoError(t, err)

	deleteReq := &productpb.DeleteProductRequest{
		ProductId: createResp.ProductId,
	}

	_, err = service.DeleteProduct(ctx, deleteReq)
	require.NoError(t, err)

	getReq := &productpb.GetProductRequest{
		ProductId: createResp.ProductId,
	}

	_, err = service.GetProduct(ctx, getReq)
	assert.Error(t, err)
}

func TestIntegration_ListProducts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	service, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	productIDs := []string{}
	for i := 0; i < 5; i++ {
		createReq := &productpb.CreateProductRequest{
			Name:        fmt.Sprintf("List Test Product %d", i),
			Description: "Product for list test",
			Sku:         fmt.Sprintf("LIST-TEST-%03d", i),
			Price: &sharedpb.Money{
				Units:    int64(1000 + i*100),
				Currency: "JPY",
			},
			StockQuantity: int32(10 + i),
		}

		resp, err := service.CreateProduct(ctx, createReq)
		require.NoError(t, err)
		_ = append(productIDs, resp.ProductId)
	}

	listReq := &productpb.ListProductsRequest{
		ActiveOnly: true,
		Pagination: &sharedpb.Pagination{
			Page:  1,
			Limit: 10,
		},
	}

	listResp, err := service.ListProducts(ctx, listReq)
	require.NoError(t, err)
	assert.NotNil(t, listResp.Products)
	assert.GreaterOrEqual(t, len(listResp.Products), 5)
}

func TestIntegration_SearchProducts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	service, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	createReq := &productpb.CreateProductRequest{
		Name:        "Special Coffee Maker",
		Description: "Premium coffee machine for home brewing",
		Sku:         "COFFEE-001",
		Price: &sharedpb.Money{
			Units:    5000,
			Currency: "JPY",
		},
		StockQuantity: 30,
	}

	resp, err := service.CreateProduct(ctx, createReq)
	require.NoError(t, err)

	searchReq := &productpb.SearchProductsRequest{
		Query: "coffee",
		Pagination: &sharedpb.Pagination{
			Page:  1,
			Limit: 10,
		},
	}

	searchResp, err := service.SearchProducts(ctx, searchReq)
	require.NoError(t, err)
	assert.NotNil(t, searchResp.Products)
	found := false
	for _, p := range searchResp.Products {
		if p.Id == resp.ProductId {
			found = true
			break
		}
	}
	assert.True(t, found, "search result should contain the created product")
}

func TestIntegration_CacheBehavior(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	service, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	createReq := &productpb.CreateProductRequest{
		Name:        "Cache Test Product",
		Description: "Testing cache behavior",
		Sku:         "CACHE-TEST-001",
		Price: &sharedpb.Money{
			Units:    999,
			Currency: "JPY",
		},
		StockQuantity: 20,
	}

	createResp, err := service.CreateProduct(ctx, createReq)
	require.NoError(t, err)

	getReq := &productpb.GetProductRequest{
		ProductId: createResp.ProductId,
	}

	start := time.Now()
	_, err = service.GetProduct(ctx, getReq)
	require.NoError(t, err)
	firstCallDuration := time.Since(start)

	start = time.Now()
	_, err = service.GetProduct(ctx, getReq)
	require.NoError(t, err)
	secondCallDuration := time.Since(start)

	assert.Less(t, secondCallDuration, firstCallDuration, "cached call should be faster")
}

func TestIntegration_Pagination(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	service, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	productIDs := []string{}
	for i := 0; i < 25; i++ {
		createReq := &productpb.CreateProductRequest{
			Name:        fmt.Sprintf("Pagination Test Product %d", i),
			Description: "Product for pagination test",
			Sku:         fmt.Sprintf("PAGE-TEST-%03d", i),
			Price: &sharedpb.Money{
				Units:    1000,
				Currency: "JPY",
			},
			StockQuantity: 10,
		}

		resp, err := service.CreateProduct(ctx, createReq)
		require.NoError(t, err)
		_ = append(productIDs, resp.ProductId)
	}

	page1Req := &productpb.ListProductsRequest{
		ActiveOnly: true,
		Pagination: &sharedpb.Pagination{
			Page:  1,
			Limit: 10,
		},
	}

	page1Resp, err := service.ListProducts(ctx, page1Req)
	require.NoError(t, err)
	assert.Len(t, page1Resp.Products, 10)

	page2Req := &productpb.ListProductsRequest{
		ActiveOnly: true,
		Pagination: &sharedpb.Pagination{
			Page:  2,
			Limit: 10,
		},
	}

	page2Resp, err := service.ListProducts(ctx, page2Req)
	require.NoError(t, err)
	assert.Len(t, page2Resp.Products, 10)

	page3Req := &productpb.ListProductsRequest{
		ActiveOnly: true,
		Pagination: &sharedpb.Pagination{
			Page:  3,
			Limit: 10,
		},
	}

	page3Resp, err := service.ListProducts(ctx, page3Req)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(page3Resp.Products), 5)

	allProductIDs := []string{}
	allProductIDs = append(allProductIDs, getProductIDs(page1Resp.Products)...)
	allProductIDs = append(allProductIDs, getProductIDs(page2Resp.Products)...)
	allProductIDs = append(allProductIDs, getProductIDs(page3Resp.Products)...)

	for _, id := range productIDs {
		assert.Contains(t, allProductIDs, id)
	}
}

func getProductIDs(products []*productpb.Product) []string {
	ids := make([]string, len(products))
	for i, p := range products {
		ids[i] = p.Id
	}
	return ids
}
