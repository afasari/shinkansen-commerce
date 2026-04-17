package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	productpb "github.com/afasari/shinkansen-commerce/gen/proto/go/product"
	sharedpb "github.com/afasari/shinkansen-commerce/gen/proto/go/shared"
	"github.com/afasari/shinkansen-commerce/services/product-service/internal/cache"
	"github.com/afasari/shinkansen-commerce/services/product-service/internal/db"
	"github.com/afasari/shinkansen-commerce/services/product-service/internal/pkg/pgutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func isDuplicateKeyError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.SQLState() == "23505"
}

func (s *ProductService) CreateProduct(ctx context.Context, req *productpb.CreateProductRequest) (*productpb.CreateProductResponse, error) {
	s.logger.Info("Creating product", zap.String("name", req.Name))

	var categoryID pgtype.UUID
	if req.CategoryId != "" {
		parsedUUID, err := uuid.Parse(req.CategoryId)
		if err != nil {
			return nil, fmt.Errorf("invalid category_id: %w", err)
		}
		categoryID = pgutil.ToPG(parsedUUID)
	}

	name := req.Name
	description := req.Description
	priceUnits := req.Price.Units
	priceCurrency := req.Price.Currency
	sku := req.Sku
	if sku == "" {
		// Generate a unique SKU based on the product name and UUID
		sku = "SKU-" + uuid.New().String()[:8]
	}
	stockQty := req.StockQuantity
	if stockQty == 0 {
		stockQty = 100 // Default to 100 items in stock for new products
	}

	productID, err := s.queries.CreateProduct(ctx, db.CreateProductParams{
		Name:          &name,
		Description:   &description,
		CategoryID:    categoryID,
		PriceUnits:    &priceUnits,
		PriceCurrency: &priceCurrency,
		Sku:           &sku,
		StockQuantity: &stockQty,
	})
	if err != nil {
		s.logger.Error("Failed to create product", zap.Error(err))
		if isDuplicateKeyError(err) {
			return nil, status.Error(codes.AlreadyExists, "product with this SKU already exists")
		}
		return nil, status.Error(codes.Internal, "failed to create product")
	}

	return &productpb.CreateProductResponse{
		ProductId: pgutil.FromPG(productID),
	}, nil
}

func (s *ProductService) GetProduct(ctx context.Context, req *productpb.GetProductRequest) (*productpb.GetProductResponse, error) {
	ctx, span := otel.Tracer("product-service").Start(ctx, "ProductService.GetProduct",
		trace.WithAttributes(attribute.String("product.id", req.ProductId)),
	)
	defer span.End()

	cacheKey := cache.ProductCacheKey(req.ProductId)
	var cached db.GetProductRow

	if err := s.cache.Get(ctx, cacheKey, &cached); err == nil {
		s.logger.Debug("Product cache hit", zap.String("product_id", req.ProductId))
		return &productpb.GetProductResponse{
			Product: s.productRowToProto(cached),
		}, nil
	}

	s.logger.Debug("Product cache miss", zap.String("product_id", req.ProductId))

	productID, err := uuid.Parse(req.ProductId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product_id")
	}

	product, err := s.queries.GetProduct(ctx, pgutil.ToPG(productID))
	if err != nil {
		s.logger.Error("Failed to get product", zap.String("product_id", req.ProductId), zap.Error(err))
		return nil, status.Error(codes.NotFound, "product not found")
	}

	if err := s.cache.Set(ctx, cacheKey, product, cache.DefaultTTL); err != nil {
		s.logger.Warn("Failed to cache product", zap.Error(err))
	}

	return &productpb.GetProductResponse{
		Product: s.productRowToProto(product),
	}, nil
}

func (s *ProductService) ListProducts(ctx context.Context, req *productpb.ListProductsRequest) (*productpb.ListProductsResponse, error) {
	s.logger.Info("Listing products",
		zap.String("category_id", req.CategoryId),
		zap.Bool("active_only", req.ActiveOnly),
	)

	var categoryID pgtype.UUID
	if req.CategoryId != "" {
		parsedUUID, err := uuid.Parse(req.CategoryId)
		if err != nil {
			return nil, fmt.Errorf("invalid category_id: %w", err)
		}
		categoryID = pgutil.ToPG(parsedUUID)
	}

	var activeOnly *bool
	if req.ActiveOnly {
		val := true
		activeOnly = &val
	}
	offset := (req.Pagination.Page - 1) * req.Pagination.Limit
	limit := req.Pagination.Limit
	products, err := s.queries.ListProducts(ctx, db.ListProductsParams{
		CategoryID: categoryID,
		ActiveOnly: activeOnly,
		Offset:     &offset,
		Limit:      &limit,
	})
	if err != nil {
		s.logger.Error("Failed to list products", zap.Error(err))
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	var productList []*productpb.Product
	for _, p := range products {
		productList = append(productList, s.listProductRowToProto(p))
	}

	return &productpb.ListProductsResponse{
		Products:   productList,
		Pagination: req.Pagination,
	}, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, req *productpb.UpdateProductRequest) (*productpb.UpdateProductResponse, error) {
	s.logger.Info("Updating product", zap.String("product_id", req.ProductId))

	id, err := uuid.Parse(req.ProductId)
	if err != nil {
		return nil, fmt.Errorf("invalid product_id: %w", err)
	}
	idpg := pgutil.ToPG(id)

	var categoryID pgtype.UUID
	if req.CategoryId != nil {
		parsedUUID, err := uuid.Parse(req.CategoryId.Value)
		if err != nil {
			return nil, fmt.Errorf("invalid category_id: %w", err)
		}
		categoryID = pgutil.ToPG(parsedUUID)
	}

	var name *string
	if req.Name != nil {
		name = &req.Name.Value
	}

	var description *string
	if req.Description != nil {
		description = &req.Description.Value
	}

	var priceUnits *int64
	if req.Price != nil {
		priceUnits = &req.Price.Units
	}

	var active *bool
	if req.Active != nil {
		active = &req.Active.Value
	}

	updatedProduct, err := s.queries.UpdateProduct(ctx, db.UpdateProductParams{
		ID:          idpg,
		Name:        name,
		Description: description,
		CategoryID:  categoryID,
		PriceUnits:  priceUnits,
		Active:      active,
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
		Product: s.updateProductRowToProto(updatedProduct),
	}, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, req *productpb.DeleteProductRequest) (*sharedpb.Empty, error) {
	s.logger.Info("Deleting product", zap.String("product_id", req.ProductId))

	productID, err := uuid.Parse(req.ProductId)
	if err != nil {
		return nil, fmt.Errorf("invalid product_id: %w", err)
	}
	if err := s.queries.DeleteProduct(ctx, pgutil.ToPG(productID)); err != nil {
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

	var categoryID pgtype.UUID
	if req.CategoryId != "" {
		parsedUUID, err := uuid.Parse(req.CategoryId)
		if err != nil {
			return nil, fmt.Errorf("invalid category_id: %w", err)
		}
		categoryID = pgutil.ToPG(parsedUUID)
	}

	var minPrice *int64
	if req.MinPrice != nil {
		minPrice = &req.MinPrice.Units
	}

	var maxPrice *int64
	if req.MaxPrice != nil {
		maxPrice = &req.MaxPrice.Units
	}

	var inStockOnly *bool
	if req.InStockOnly {
		inStockOnly = &req.InStockOnly
	}

	query := req.Query
	offset := (req.Pagination.Page - 1) * req.Pagination.Limit
	limit := req.Pagination.Limit
	products, err := s.queries.SearchProducts(ctx, db.SearchProductsParams{
		Query:       &query,
		CategoryID:  categoryID,
		MinPrice:    minPrice,
		MaxPrice:    maxPrice,
		InStockOnly: inStockOnly,
		Offset:      &offset,
		Limit:       &limit,
	})
	if err != nil {
		s.logger.Error("Failed to search products", zap.Error(err))
		return nil, fmt.Errorf("failed to search products: %w", err)
	}

	var productList []*productpb.Product
	for _, p := range products {
		productList = append(productList, s.searchProductRowToProto(p))
	}

	return &productpb.SearchProductsResponse{
		Products:   productList,
		Pagination: req.Pagination,
	}, nil
}

func (s *ProductService) GetProductVariants(ctx context.Context, req *productpb.GetProductVariantsRequest) (*productpb.GetProductVariantsResponse, error) {
	s.logger.Info("Getting product variants", zap.String("product_id", req.ProductId))

	productID, err := uuid.Parse(req.ProductId)
	if err != nil {
		return nil, fmt.Errorf("invalid product_id: %w", err)
	}
	variants, err := s.queries.GetProductVariants(ctx, pgutil.ToPG(productID))
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

func (s *ProductService) productRowToProto(p db.GetProductRow) *productpb.Product {
	return s.getProductBase(p.ID, p.Name, p.Description, p.CategoryID, p.PriceUnits, p.PriceCurrency, p.Sku, p.Active, p.StockQuantity)
}

func (s *ProductService) listProductRowToProto(p db.ListProductsRow) *productpb.Product {
	return s.getProductBase(p.ID, p.Name, p.Description, p.CategoryID, p.PriceUnits, p.PriceCurrency, p.Sku, p.Active, p.StockQuantity)
}

func (s *ProductService) updateProductRowToProto(p db.UpdateProductRow) *productpb.Product {
	return s.getProductBase(p.ID, p.Name, p.Description, p.CategoryID, p.PriceUnits, p.PriceCurrency, p.Sku, p.Active, p.StockQuantity)
}

func (s *ProductService) searchProductRowToProto(p db.SearchProductsRow) *productpb.Product {
	return s.getProductBase(p.ID, p.Name, p.Description, p.CategoryID, p.PriceUnits, p.PriceCurrency, p.Sku, p.Active, p.StockQuantity)
}

func (s *ProductService) getProductBase(id pgtype.UUID, name string, description *string, categoryID pgtype.UUID, priceUnits int64, priceCurrency string, sku string, active *bool, stockQuantity *int32) *productpb.Product {
	desc := ""
	if description != nil {
		desc = *description
	}

	activeVal := false
	if active != nil {
		activeVal = *active
	}

	stockQty := int32(0)
	if stockQuantity != nil {
		stockQty = *stockQuantity
	}

	return &productpb.Product{
		Id:          pgutil.FromPG(id),
		Name:        name,
		Description: desc,
		CategoryId:  pgutil.FromPG(categoryID),
		Price: &sharedpb.Money{
			Units:    priceUnits,
			Currency: priceCurrency,
		},
		Sku:           sku,
		Active:        activeVal,
		StockQuantity: stockQty,
		ImageUrls:     []string{},
	}
}

func (s *ProductService) variantToProto(v db.CatalogProductVariants) *productpb.ProductVariant {
	sku := ""
	if v.Sku != nil {
		sku = *v.Sku
	}

	stockQty := int32(0)
	if v.StockQuantity != nil {
		stockQty = *v.StockQuantity
	}

	attributes := make(map[string]string)
	if len(v.Attributes) > 0 {
		_ = json.Unmarshal(v.Attributes, &attributes)
	}

	return &productpb.ProductVariant{
		Id:        pgutil.FromPG(v.ID),
		ProductId: pgutil.FromPG(v.ProductID),
		Name:      v.Name,
		Price: &sharedpb.Money{
			Units:    v.PriceUnits,
			Currency: v.PriceCurrency,
		},
		Sku:           sku,
		StockQuantity: stockQty,
		Attributes:    attributes,
	}
}
