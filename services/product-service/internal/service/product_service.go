package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	commonv1 "github.com/shinkansen-commerce/shinkansen/gen/proto/go/common"
	productv1 "github.com/shinkansen-commerce/shinkansen/gen/proto/go/product"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ProductService struct {
	productv1.UnimplementedProductServiceServer
	queries *repository.Queries
	logger  *zap.Logger
}

func NewProductService(queries *repository.Queries, logger *zap.Logger) *ProductService {
	return &ProductService{
		queries: queries,
		logger:  logger,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, req *productv1.CreateProductRequest) (*productv1.CreateProductResponse, error) {
	s.logger.Info("Creating product", zap.String("name", req.Name))

	productID, err := s.queries.CreateProduct(ctx, repository.CreateProductParams{
		Name:          req.Name,
		Description:   req.Description,
		CategoryID:    req.CategoryId,
		PriceUnits:    req.Price.Units,
		PriceCurrency: req.Price.Currency,
		Sku:           req.Sku,
		Active:        true,
		StockQuantity: req.StockQuantity,
	})
	if err != nil {
		s.logger.Error("Failed to create product", zap.Error(err))
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return &productv1.CreateProductResponse{
		ProductId: productID.String(),
	}, nil
}

func (s *ProductService) GetProduct(ctx context.Context, req *productv1.GetProductRequest) (*productv1.GetProductResponse, error) {
	product, err := s.queries.GetProduct(ctx, req.ProductId)
	if err != nil {
		s.logger.Error("Failed to get product", zap.String("product_id", req.ProductId), zap.Error(err))
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &productv1.GetProductResponse{
		Product: &productv1.Product{
			Id:            product.ID.String(),
			Name:          product.Name,
			Description:   product.Description,
			CategoryId:    product.CategoryID.String(),
			Price:         &commonv1.Money{Units: product.PriceUnits, Currency: product.PriceCurrency},
			Sku:           product.Sku,
			Active:        product.Active,
			CreatedAt:     timestamppb.New(product.CreatedAt),
			UpdatedAt:     timestamppb.New(product.UpdatedAt),
			StockQuantity: int32(product.StockQuantity),
		},
	}, nil
}

func (s *ProductService) ListProducts(ctx context.Context, req *productv1.ListProductsRequest) (*productv1.ListProductsResponse, error) {
	products, err := s.queries.ListProducts(ctx, repository.ListProductsParams{
		ActiveOnly: req.ActiveOnly,
		CategoryID: req.CategoryId,
		Limit:      req.Pagination.Limit,
		Offset:     (req.Pagination.Page - 1) * req.Pagination.Limit,
	})
	if err != nil {
		s.logger.Error("Failed to list products", zap.Error(err))
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	var productList []*productv1.Product
	for _, p := range products {
		productList = append(productList, &productv1.Product{
			Id:            p.ID.String(),
			Name:          p.Name,
			Description:   p.Description,
			CategoryId:    p.CategoryID.String(),
			Price:         &commonv1.Money{Units: p.PriceUnits, Currency: p.PriceCurrency},
			Sku:           p.Sku,
			Active:        p.Active,
			CreatedAt:     timestamppb.New(p.CreatedAt),
			UpdatedAt:     timestamppb.New(p.UpdatedAt),
			StockQuantity: int32(p.StockQuantity),
		})
	}

	return &productv1.ListProductsResponse{
		Products:   productList,
		Pagination: req.Pagination,
	}, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, req *productv1.UpdateProductRequest) (*productv1.UpdateProductResponse, error) {
	s.logger.Info("Updating product", zap.String("product_id", req.ProductId))

	updatedProduct, err := s.queries.UpdateProduct(ctx, repository.UpdateProductParams{
		ID:          uuid.MustParse(req.ProductId),
		Name:        req.Name.Value,
		Description: req.Description.Value,
		CategoryID:  req.CategoryId.Value,
		PriceUnits:  req.Price.Units,
		Active:      req.Active.Value,
	})
	if err != nil {
		s.logger.Error("Failed to update product", zap.Error(err))
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return &productv1.UpdateProductResponse{
		Product: &productv1.Product{
			Id:          updatedProduct.ID.String(),
			Name:        updatedProduct.Name,
			Description: updatedProduct.Description,
			CategoryId:  updatedProduct.CategoryID.String(),
			Price:       &commonv1.Money{Units: updatedProduct.PriceUnits, Currency: updatedProduct.PriceCurrency},
			Sku:         updatedProduct.Sku,
			Active:      updatedProduct.Active,
			CreatedAt:   timestamppb.New(updatedProduct.CreatedAt),
			UpdatedAt:   timestamppb.New(updatedProduct.UpdatedAt),
		},
	}, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, req *productv1.DeleteProductRequest) (*commonv1.Empty, error) {
	s.logger.Info("Deleting product", zap.String("product_id", req.ProductId))

	err := s.queries.DeleteProduct(ctx, req.ProductId)
	if err != nil {
		s.logger.Error("Failed to delete product", zap.Error(err))
		return nil, fmt.Errorf("failed to delete product: %w", err)
	}

	return &commonv1.Empty{}, nil
}

func (s *ProductService) SearchProducts(ctx context.Context, req *productv1.SearchProductsRequest) (*productv1.SearchProductsResponse, error) {
	products, err := s.queries.SearchProducts(ctx, repository.SearchProductsParams{
		Query:       req.Query,
		CategoryID:  req.CategoryId,
		MinPrice:    req.MinPrice.Units,
		MaxPrice:    req.MaxPrice.Units,
		InStockOnly: req.InStockOnly,
		Limit:       req.Pagination.Limit,
		Offset:      (req.Pagination.Page - 1) * req.Pagination.Limit,
	})
	if err != nil {
		s.logger.Error("Failed to search products", zap.Error(err))
		return nil, fmt.Errorf("failed to search products: %w", err)
	}

	var productList []*productv1.Product
	for _, p := range products {
		productList = append(productList, &productv1.Product{
			Id:            p.ID.String(),
			Name:          p.Name,
			Description:   p.Description,
			CategoryId:    p.CategoryID.String(),
			Price:         &commonv1.Money{Units: p.PriceUnits, Currency: p.PriceCurrency},
			Sku:           p.Sku,
			Active:        p.Active,
			CreatedAt:     timestamppb.New(p.CreatedAt),
			UpdatedAt:     timestamppb.New(p.UpdatedAt),
			StockQuantity: int32(p.StockQuantity),
		})
	}

	return &productv1.SearchProductsResponse{
		Products:   productList,
		Pagination: req.Pagination,
	}, nil
}

func (s *ProductService) GetProductVariants(ctx context.Context, req *productv1.GetProductVariantsRequest) (*productv1.GetProductVariantsResponse, error) {
	variants, err := s.queries.GetProductVariants(ctx, req.ProductId)
	if err != nil {
		s.logger.Error("Failed to get product variants", zap.Error(err))
		return nil, fmt.Errorf("failed to get product variants: %w", err)
	}

	var variantList []*productv1.ProductVariant
	for _, v := range variants {
		variantList = append(variantList, &productv1.ProductVariant{
			Id:            v.ID.String(),
			ProductId:     v.ProductID.String(),
			Name:          v.Name,
			Price:         &commonv1.Money{Units: v.PriceUnits, Currency: v.PriceCurrency},
			Sku:           v.Sku,
			StockQuantity: int32(v.StockQuantity),
		})
	}

	return &productv1.GetProductVariantsResponse{
		Variants: variantList,
	}, nil
}
