package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	productpb "github.com/shinkansen-commerce/shinkansen/gen/proto/go/product"
	sharedpb "github.com/shinkansen-commerce/shinkansen/gen/proto/go/shared"
	"github.com/shinkansen-commerce/shinkansen/services/product-service/internal/cache"
	"github.com/shinkansen-commerce/shinkansen/services/product-service/internal/db"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type ProductService struct {
	productpb.UnimplementedProductServiceServer
	queries db.Querier
	cache   cache.Cache
	logger  *zap.Logger
}

func NewProductService(queries db.Querier, cacheClient cache.Cache, logger *zap.Logger) *ProductService {
	return &ProductService{
		queries: queries,
		cache:   cacheClient,
		logger:  logger,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, req *productpb.CreateProductRequest) (*productpb.CreateProductResponse, error) {
	s.logger.Info("Creating product", zap.String("name", req.Name))

	var categoryID *uuid.UUID
	if req.CategoryId != "" {
		id := uuid.MustParse(req.CategoryId)
		categoryID = &id
	}

	productID, err := s.queries.CreateProduct(ctx, db.CreateProductParams{
		Name:          req.Name,
		Description:   &req.Description,
		CategoryID:    categoryID,
		PriceUnits:    req.Price.Units,
		PriceCurrency: req.Price.Currency,
		Sku:           req.Sku,
		StockQuantity: req.StockQuantity,
	})
	if err != nil {
		s.logger.Error("Failed to create product", zap.Error(err))
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return &productpb.CreateProductResponse{
		ProductId: productID.String(),
	}, nil
}

func (s *ProductService) GetProduct(ctx context.Context, req *productpb.GetProductRequest) (*productpb.GetProductResponse, error) {
	cacheKey := cache.ProductCacheKey(req.ProductId)
	var cached db.Product

	if err := s.cache.Get(ctx, cacheKey, &cached); err == nil {
		s.logger.Debug("Product cache hit", zap.String("product_id", req.ProductId))
		return &productpb.GetProductResponse{
			Product: s.productToProto(cached),
		}, nil
	}

	s.logger.Debug("Product cache miss", zap.String("product_id", req.ProductId))

	productID := uuid.MustParse(req.ProductId)
	product, err := s.queries.GetProduct(ctx, productID)
	if err != nil {
		s.logger.Error("Failed to get product", zap.String("product_id", req.ProductId), zap.Error(err))
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	if err := s.cache.Set(ctx, cacheKey, product, cache.DefaultTTL); err != nil {
		s.logger.Warn("Failed to cache product", zap.Error(err))
	}

	return &productpb.GetProductResponse{
		Product: s.productToProto(product),
	}, nil
}

func (s *ProductService) ListProducts(ctx context.Context, req *productpb.ListProductsRequest) (*productpb.ListProductsResponse, error) {
	s.logger.Info("Listing products",
		zap.String("category_id", req.CategoryId),
		zap.Bool("active_only", req.ActiveOnly),
	)

	var categoryID *uuid.UUID
	if req.CategoryId != "" {
		id := uuid.MustParse(req.CategoryId)
		categoryID = &id
	}

	products, err := s.queries.ListProducts(ctx, db.ListProductsParams{
		CategoryID: categoryID,
		ActiveOnly: req.ActiveOnly,
		Limit:      req.Pagination.Limit,
		Offset:     (req.Pagination.Page - 1) * req.Pagination.Limit,
	})
	if err != nil {
		s.logger.Error("Failed to list products", zap.Error(err))
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	var productList []*productpb.Product
	for _, p := range products {
		productList = append(productList, s.productToProto(p))
	}

	return &productpb.ListProductsResponse{
		Products:   productList,
		Pagination: req.Pagination,
	}, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, req *productpb.UpdateProductRequest) (*productpb.UpdateProductResponse, error) {
	s.logger.Info("Updating product", zap.String("product_id", req.ProductId))

	var name *string
	if req.Name != nil {
		name = &req.Name.Value
	}

	var description *string
	if req.Description != nil {
		description = &req.Description.Value
	}

	var categoryID *uuid.UUID
	if req.CategoryId != nil {
		id := uuid.MustParse(req.CategoryId.Value)
		categoryID = &id
	}

	var priceUnits *int64
	if req.Price != nil {
		priceUnits = &req.Price.Units
	}

	var priceCurrency *string
	if req.Price != nil {
		priceCurrency = &req.Price.Currency
	}

	updatedProduct, err := s.queries.UpdateProduct(ctx, db.UpdateProductParams{
		ID:            uuid.MustParse(req.ProductId),
		Name:          name,
		Description:   description,
		CategoryID:    categoryID,
		PriceUnits:    priceUnits,
		PriceCurrency: priceCurrency,
		Active:        activePtr(req.Active),
	})
	if err != nil {
		s.logger.Error("Failed to update product", zap.String("product_id", req.ProductId), zap.Error(err))
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	cacheKey := cache.ProductCacheKey(req.ProductId)
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		s.logger.Warn("Failed to invalidate product cache", zap.Error(err))
	}

	return &productpb.UpdateProductResponse{
		Product: s.productToProto(updatedProduct),
	}, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, req *productpb.DeleteProductRequest) (*sharedpb.Empty, error) {
	s.logger.Info("Deleting product", zap.String("product_id", req.ProductId))

	productID := uuid.MustParse(req.ProductId)
	if err := s.queries.DeleteProduct(ctx, productID); err != nil {
		s.logger.Error("Failed to delete product", zap.String("product_id", req.ProductId), zap.Error(err))
		return nil, fmt.Errorf("failed to delete product: %w", err)
	}

	cacheKey := cache.ProductCacheKey(req.ProductId)
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		s.logger.Warn("Failed to invalidate product cache", zap.Error(err))
	}

	return &sharedpb.Empty{}, nil
}

func (s *ProductService) SearchProducts(ctx context.Context, req *productpb.SearchProductsRequest) (*productpb.SearchProductsResponse, error) {
	s.logger.Info("Searching products", zap.String("query", req.Query))

	var categoryID *uuid.UUID
	if req.CategoryId != "" {
		id := uuid.MustParse(req.CategoryId)
		categoryID = &id
	}

	var minPrice *int64
	if req.MinPrice != nil {
		minPrice = &req.MinPrice.Units
	}

	var maxPrice *int64
	if req.MaxPrice != nil {
		maxPrice = &req.MaxPrice.Units
	}

	products, err := s.queries.SearchProducts(ctx, db.SearchProductsParams{
		Query:       req.Query,
		CategoryID:  categoryID,
		MinPrice:    minPrice,
		MaxPrice:    maxPrice,
		InStockOnly: req.InStockOnly,
		Limit:       req.Pagination.Limit,
		Offset:      (req.Pagination.Page - 1) * req.Pagination.Limit,
	})
	if err != nil {
		s.logger.Error("Failed to search products", zap.Error(err))
		return nil, fmt.Errorf("failed to search products: %w", err)
	}

	var productList []*productpb.Product
	for _, p := range products {
		productList = append(productList, s.productToProto(p))
	}

	return &productpb.SearchProductsResponse{
		Products:   productList,
		Pagination: req.Pagination,
	}, nil
}

func (s *ProductService) GetProductVariants(ctx context.Context, req *productpb.GetProductVariantsRequest) (*productpb.GetProductVariantsResponse, error) {
	s.logger.Info("Getting product variants", zap.String("product_id", req.ProductId))

	productID := uuid.MustParse(req.ProductId)
	variants, err := s.queries.GetProductVariants(ctx, productID)
	if err != nil {
		s.logger.Error("Failed to get product variants", zap.String("product_id", req.ProductId), zap.Error(err))
		return nil, fmt.Errorf("failed to get product variants: %w", err)
	}

	var variantList []*productpb.ProductVariant
	for _, v := range variants {
		variantList = append(variantList, s.variantToProto(v))
	}

	return &productpb.GetProductVariantsResponse{
		Variants: variantList,
	}, nil
}

func (s *ProductService) productToProto(p db.Product) *productpb.Product {
	categoryID := ""
	if p.CategoryID != nil {
		categoryID = p.CategoryID.String()
	}

	description := ""
	if p.Description != nil {
		description = *p.Description
	}

	return &productpb.Product{
		Id:            p.ID.String(),
		Name:          p.Name,
		Description:   description,
		CategoryId:    categoryID,
		Price:         s.moneyToProto(p.PriceUnits, p.PriceCurrency),
		Sku:           p.Sku,
		Active:        p.Active,
		CreatedAt:     protoTime(p.CreatedAt),
		UpdatedAt:     protoTime(p.UpdatedAt),
		StockQuantity: p.StockQuantity,
		ImageUrls:     []string{},
	}
}

func (s *ProductService) variantToProto(v db.ProductVariant) *productpb.ProductVariant {
	sku := ""
	if v.Sku != "" {
		sku = v.Sku
	}

	return &productpb.ProductVariant{
		Id:            v.ID.String(),
		ProductId:     v.ProductID.String(),
		Name:          v.Name,
		Price:         s.moneyToProto(v.PriceUnits, v.PriceCurrency),
		Sku:           sku,
		StockQuantity: v.StockQuantity,
		Attributes:    v.Attributes,
	}
}

func (s *ProductService) moneyToProto(units int64, currency string) *sharedpb.Money {
	return &sharedpb.Money{
		Units:    units,
		Currency: currency,
	}
}

func protoTime(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}

func activePtr(val *wrapperspb.BoolValue) *bool {
	if val == nil {
		return nil
	}
	return &val.Value
}
