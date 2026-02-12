package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/wrapperspb"

	productpb "github.com/afasari/shinkansen-commerce/gen/proto/go/product"
	sharedpb "github.com/afasari/shinkansen-commerce/gen/proto/go/shared"
	"github.com/afasari/shinkansen-commerce/services/product-service/internal/cache"
	"github.com/afasari/shinkansen-commerce/services/product-service/internal/db"
)

func setupTestService(t *testing.T) (*ProductService, *db.MockQuerier, *cache.MockCache) {
	mockQueries := new(db.MockQuerier)
	mockCache := new(cache.MockCache)
	logger := zap.NewNop()

	service := NewProductService(mockQueries, mockCache, logger)
	return service, mockQueries, mockCache
}

func TestProductService_CreateProduct(t *testing.T) {
	tests := []struct {
		name    string
		req     *productpb.CreateProductRequest
		setup   func(*db.MockQuerier, *cache.MockCache)
		want    *productpb.CreateProductResponse
		wantErr bool
		errMsg  string
	}{
		{
			name: "success",
			req: &productpb.CreateProductRequest{
				Name:        "Test Product",
				Description: "Test Description",
				Sku:         "TEST001",
				Price: &sharedpb.Money{
					Units:    1000,
					Currency: "JPY",
				},
				StockQuantity: 10,
			},
			setup: func(mq *db.MockQuerier, mc *cache.MockCache) {
				productID := uuid.New()
				mq.On("CreateProduct", mock.Anything, mock.Anything).
					Return(productID, nil)
			},
			want: &productpb.CreateProductResponse{
				ProductId: "test-id",
			},
			wantErr: false,
		},
		{
			name: "success with category",
			req: &productpb.CreateProductRequest{
				Name:        "Test Product",
				Description: "Test Description",
				CategoryId:  uuid.New().String(),
				Sku:         "TEST002",
				Price: &sharedpb.Money{
					Units:    2000,
					Currency: "JPY",
				},
				StockQuantity: 20,
			},
			setup: func(mq *db.MockQuerier, mc *cache.MockCache) {
				productID := uuid.New()
				mq.On("CreateProduct", mock.Anything, mock.Anything).
					Return(productID, nil)
			},
			wantErr: false,
		},
		{
			name: "database error",
			req: &productpb.CreateProductRequest{
				Name: "Test Product",
				Sku:  "TEST003",
				Price: &sharedpb.Money{
					Units:    1000,
					Currency: "JPY",
				},
				StockQuantity: 10,
			},
			setup: func(mq *db.MockQuerier, mc *cache.MockCache) {
				mq.On("CreateProduct", mock.Anything, mock.Anything).
					Return(uuid.Nil, errors.New("database error"))
			},
			wantErr: true,
			errMsg:  "failed to create product",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockQueries, mockCache := setupTestService(t)
			tt.setup(mockQueries, mockCache)

			got, err := service.CreateProduct(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.NotEmpty(t, got.ProductId)
			}

			mockQueries.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

func TestProductService_GetProduct(t *testing.T) {
	tests := []struct {
		name    string
		req     *productpb.GetProductRequest
		setup   func(*db.MockQuerier, *cache.MockCache)
		want    *productpb.GetProductResponse
		wantErr bool
		errMsg  string
	}{
		{
			name: "cache hit",
			req: &productpb.GetProductRequest{
				ProductId: uuid.New().String(),
			},
			setup: func(mq *db.MockQuerier, mc *cache.MockCache) {
				mc.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mq.AssertNotCalled(t, "GetProduct", mock.Anything, mock.Anything)
			},
			wantErr: false,
		},
		{
			name: "cache miss - success",
			req: &productpb.GetProductRequest{
				ProductId: uuid.New().String(),
			},
			setup: func(mq *db.MockQuerier, mc *cache.MockCache) {
				productID := uuid.New()
				product := db.Product{
					ID:            productID,
					Name:          "Test Product",
					Description:   strPtr("Test Description"),
					PriceUnits:    1000,
					PriceCurrency: "JPY",
					Sku:           "TEST001",
					Active:        true,
					StockQuantity: 10,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				}
				mc.On("Get", mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("cache miss"))
				mq.On("GetProduct", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(product, nil)
				mc.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "product not found",
			req: &productpb.GetProductRequest{
				ProductId: uuid.New().String(),
			},
			setup: func(mq *db.MockQuerier, mc *cache.MockCache) {
				mc.On("Get", mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("cache miss"))
				mq.On("GetProduct", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(db.Product{}, errors.New("not found"))
			},
			wantErr: true,
			errMsg:  "failed to get product",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockQueries, mockCache := setupTestService(t)
			tt.setup(mockQueries, mockCache)

			got, err := service.GetProduct(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.NotNil(t, got.Product)
			}

			mockQueries.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

func TestProductService_ListProducts(t *testing.T) {
	tests := []struct {
		name    string
		req     *productpb.ListProductsRequest
		setup   func(*db.MockQuerier, *cache.MockCache)
		wantErr bool
		errMsg  string
	}{
		{
			name: "success with products",
			req: &productpb.ListProductsRequest{
				ActiveOnly: true,
				Pagination: &sharedpb.Pagination{
					Page:  1,
					Limit: 10,
				},
			},
			setup: func(mq *db.MockQuerier, mc *cache.MockCache) {
				products := []db.Product{
					{
						ID:            uuid.New(),
						Name:          "Product 1",
						PriceUnits:    1000,
						PriceCurrency: "JPY",
						Sku:           "SKU001",
						Active:        true,
						StockQuantity: 10,
						CreatedAt:     time.Now(),
						UpdatedAt:     time.Now(),
					},
				}
				mq.On("ListProducts", mock.Anything, mock.Anything).
					Return(products, nil)
			},
			wantErr: false,
		},
		{
			name: "success with empty list",
			req: &productpb.ListProductsRequest{
				ActiveOnly: true,
				Pagination: &sharedpb.Pagination{
					Page:  1,
					Limit: 10,
				},
			},
			setup: func(mq *db.MockQuerier, mc *cache.MockCache) {
				mq.On("ListProducts", mock.Anything, mock.Anything).
					Return([]db.Product{}, nil)
			},
			wantErr: false,
		},
		{
			name: "database error",
			req: &productpb.ListProductsRequest{
				ActiveOnly: true,
				Pagination: &sharedpb.Pagination{
					Page:  1,
					Limit: 10,
				},
			},
			setup: func(mq *db.MockQuerier, mc *cache.MockCache) {
				mq.On("ListProducts", mock.Anything, mock.Anything).
					Return([]db.Product(nil), errors.New("database error"))
			},
			wantErr: true,
			errMsg:  "failed to list products",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockQueries, mockCache := setupTestService(t)
			tt.setup(mockQueries, mockCache)

			got, err := service.ListProducts(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
			}

			mockQueries.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

func TestProductService_UpdateProduct(t *testing.T) {
	tests := []struct {
		name    string
		req     *productpb.UpdateProductRequest
		setup   func(*db.MockQuerier, *cache.MockCache)
		wantErr bool
		errMsg  string
	}{
		{
			name: "success - partial update",
			req: &productpb.UpdateProductRequest{
				ProductId: uuid.New().String(),
				Name:      &wrapperspb.StringValue{Value: "Updated Name"},
				Active:    &wrapperspb.BoolValue{Value: true},
			},
			setup: func(mq *db.MockQuerier, mc *cache.MockCache) {
				product := db.Product{
					ID:            uuid.New(),
					Name:          "Updated Name",
					PriceUnits:    1000,
					PriceCurrency: "JPY",
					Sku:           "SKU001",
					Active:        true,
					StockQuantity: 10,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				}
				mq.On("UpdateProduct", mock.Anything, mock.Anything).
					Return(product, nil)
				mc.On("Delete", mock.Anything, mock.Anything).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "product not found",
			req: &productpb.UpdateProductRequest{
				ProductId: uuid.New().String(),
				Name:      &wrapperspb.StringValue{Value: "Updated Name"},
			},
			setup: func(mq *db.MockQuerier, mc *cache.MockCache) {
				mq.On("UpdateProduct", mock.Anything, mock.Anything).
					Return(db.Product{}, errors.New("not found"))
			},
			wantErr: true,
			errMsg:  "failed to update product",
		},
		{
			name: "cache delete failure - still updates product",
			req: &productpb.UpdateProductRequest{
				ProductId: uuid.New().String(),
				Name:      &wrapperspb.StringValue{Value: "Updated Name"},
			},
			setup: func(mq *db.MockQuerier, mc *cache.MockCache) {
				product := db.Product{
					ID:            uuid.New(),
					Name:          "Updated Name",
					PriceUnits:    1000,
					PriceCurrency: "JPY",
					Sku:           "SKU001",
					Active:        true,
					StockQuantity: 10,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				}
				mq.On("UpdateProduct", mock.Anything, mock.Anything).
					Return(product, nil)
				mc.On("Delete", mock.Anything, mock.Anything).
					Return(errors.New("cache error"))
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockQueries, mockCache := setupTestService(t)
			tt.setup(mockQueries, mockCache)

			got, err := service.UpdateProduct(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.NotNil(t, got.Product)
			}

			mockQueries.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

func TestProductService_DeleteProduct(t *testing.T) {
	tests := []struct {
		name    string
		req     *productpb.DeleteProductRequest
		setup   func(*db.MockQuerier, *cache.MockCache)
		wantErr bool
		errMsg  string
	}{
		{
			name: "success",
			req: &productpb.DeleteProductRequest{
				ProductId: uuid.New().String(),
			},
			setup: func(mq *db.MockQuerier, mc *cache.MockCache) {
				mq.On("DeleteProduct", mock.Anything, mock.Anything).
					Return(nil)
				mc.On("Delete", mock.Anything, mock.Anything).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "product not found",
			req: &productpb.DeleteProductRequest{
				ProductId: uuid.New().String(),
			},
			setup: func(mq *db.MockQuerier, mc *cache.MockCache) {
				mq.On("DeleteProduct", mock.Anything, mock.Anything).
					Return(errors.New("not found"))
			},
			wantErr: true,
			errMsg:  "failed to delete product",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockQueries, mockCache := setupTestService(t)
			tt.setup(mockQueries, mockCache)

			got, err := service.DeleteProduct(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
			}

			mockQueries.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

func TestProductService_SearchProducts(t *testing.T) {
	tests := []struct {
		name    string
		req     *productpb.SearchProductsRequest
		setup   func(*db.MockQuerier, *cache.MockCache)
		wantErr bool
		errMsg  string
	}{
		{
			name: "success with results",
			req: &productpb.SearchProductsRequest{
				Query:       "test product",
				InStockOnly: true,
				Pagination: &sharedpb.Pagination{
					Page:  1,
					Limit: 10,
				},
			},
			setup: func(mq *db.MockQuerier, mc *cache.MockCache) {
				products := []db.Product{
					{
						ID:            uuid.New(),
						Name:          "Test Product",
						PriceUnits:    1000,
						PriceCurrency: "JPY",
						Sku:           "SKU001",
						Active:        true,
						StockQuantity: 10,
						CreatedAt:     time.Now(),
						UpdatedAt:     time.Now(),
					},
				}
				mq.On("SearchProducts", mock.Anything, mock.Anything).
					Return(products, nil)
			},
			wantErr: false,
		},
		{
			name: "success with price range",
			req: &productpb.SearchProductsRequest{
				Query:    "test",
				MinPrice: &sharedpb.Money{Units: 1000, Currency: "JPY"},
				MaxPrice: &sharedpb.Money{Units: 5000, Currency: "JPY"},
				Pagination: &sharedpb.Pagination{
					Page:  1,
					Limit: 10,
				},
			},
			setup: func(mq *db.MockQuerier, mc *cache.MockCache) {
				mq.On("SearchProducts", mock.Anything, mock.Anything).
					Return([]db.Product{}, nil)
			},
			wantErr: false,
		},
		{
			name: "empty results",
			req: &productpb.SearchProductsRequest{
				Query: "nonexistent",
				Pagination: &sharedpb.Pagination{
					Page:  1,
					Limit: 10,
				},
			},
			setup: func(mq *db.MockQuerier, mc *cache.MockCache) {
				mq.On("SearchProducts", mock.Anything, mock.Anything).
					Return([]db.Product{}, nil)
			},
			wantErr: false,
		},
		{
			name: "database error",
			req: &productpb.SearchProductsRequest{
				Query: "test",
				Pagination: &sharedpb.Pagination{
					Page:  1,
					Limit: 10,
				},
			},
			setup: func(mq *db.MockQuerier, mc *cache.MockCache) {
				mq.On("SearchProducts", mock.Anything, mock.Anything).
					Return([]db.Product(nil), errors.New("database error"))
			},
			wantErr: true,
			errMsg:  "failed to search products",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockQueries, mockCache := setupTestService(t)
			tt.setup(mockQueries, mockCache)

			got, err := service.SearchProducts(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
			}

			mockQueries.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

func TestProductService_GetProductVariants(t *testing.T) {
	tests := []struct {
		name    string
		req     *productpb.GetProductVariantsRequest
		setup   func(*db.MockQuerier, *cache.MockCache)
		wantErr bool
		errMsg  string
	}{
		{
			name: "success with variants",
			req: &productpb.GetProductVariantsRequest{
				ProductId: uuid.New().String(),
			},
			setup: func(mq *db.MockQuerier, mc *cache.MockCache) {
				variants := []db.ProductVariant{
					{
						ID:            uuid.New(),
						ProductID:     uuid.New(),
						Name:          "Variant 1",
						PriceUnits:    1000,
						PriceCurrency: "JPY",
						StockQuantity: 10,
						CreatedAt:     time.Now(),
						UpdatedAt:     time.Now(),
					},
				}
				mq.On("GetProductVariants", mock.Anything, mock.Anything).
					Return(variants, nil)
			},
			wantErr: false,
		},
		{
			name: "success with no variants",
			req: &productpb.GetProductVariantsRequest{
				ProductId: uuid.New().String(),
			},
			setup: func(mq *db.MockQuerier, mc *cache.MockCache) {
				mq.On("GetProductVariants", mock.Anything, mock.Anything).
					Return([]db.ProductVariant{}, nil)
			},
			wantErr: false,
		},
		{
			name: "product not found",
			req: &productpb.GetProductVariantsRequest{
				ProductId: uuid.New().String(),
			},
			setup: func(mq *db.MockQuerier, mc *cache.MockCache) {
				mq.On("GetProductVariants", mock.Anything, mock.Anything).
					Return([]db.ProductVariant(nil), errors.New("not found"))
			},
			wantErr: true,
			errMsg:  "failed to get product variants",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockQueries, mockCache := setupTestService(t)
			tt.setup(mockQueries, mockCache)

			got, err := service.GetProductVariants(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
			}

			mockQueries.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

func TestProductService_productToProto(t *testing.T) {
	service, _, _ := setupTestService(t)

	tests := []struct {
		name string
		p    db.Product
	}{
		{
			name: "full product",
			p: db.Product{
				ID:            uuid.New(),
				Name:          "Test Product",
				Description:   strPtr("Test Description"),
				CategoryID:    uuidPtr(uuid.New()),
				PriceUnits:    1000,
				PriceCurrency: "JPY",
				Sku:           "SKU001",
				Active:        true,
				StockQuantity: 10,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
		},
		{
			name: "product without optional fields",
			p: db.Product{
				ID:            uuid.New(),
				Name:          "Test Product",
				PriceUnits:    1000,
				PriceCurrency: "JPY",
				Sku:           "SKU001",
				Active:        true,
				StockQuantity: 10,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			protoProduct := service.productToProto(tt.p)

			assert.NotNil(t, protoProduct)
			assert.Equal(t, tt.p.ID.String(), protoProduct.Id)
			assert.Equal(t, tt.p.Name, protoProduct.Name)
			assert.Equal(t, tt.p.Sku, protoProduct.Sku)
			assert.Equal(t, tt.p.Active, protoProduct.Active)
			assert.NotNil(t, protoProduct.Price)
			assert.Equal(t, tt.p.PriceUnits, protoProduct.Price.Units)
			assert.Equal(t, tt.p.PriceCurrency, protoProduct.Price.Currency)
		})
	}
}

func TestProductService_variantToProto(t *testing.T) {
	service, _, _ := setupTestService(t)

	tests := []struct {
		name string
		v    db.ProductVariant
	}{
		{
			name: "full variant",
			v: db.ProductVariant{
				ID:            uuid.New(),
				ProductID:     uuid.New(),
				Name:          "Variant 1",
				Attributes:    map[string]string{"color": "red", "size": "M"},
				PriceUnits:    1000,
				PriceCurrency: "JPY",
				Sku:           "VAR001",
				StockQuantity: 10,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			protoVariant := service.variantToProto(tt.v)

			assert.NotNil(t, protoVariant)
			assert.Equal(t, tt.v.ID.String(), protoVariant.Id)
			assert.Equal(t, tt.v.ProductID.String(), protoVariant.ProductId)
			assert.Equal(t, tt.v.Name, protoVariant.Name)
			assert.NotNil(t, protoVariant.Price)
			assert.Equal(t, tt.v.PriceUnits, protoVariant.Price.Units)
			assert.Equal(t, tt.v.PriceCurrency, protoVariant.Price.Currency)
			assert.NotNil(t, protoVariant.Attributes)
		})
	}
}

func strPtr(s string) *string {
	return &s
}

func uuidPtr(u uuid.UUID) *uuid.UUID {
	return &u
}
