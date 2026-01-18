package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Queries struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Queries {
	return &Queries{db: db}
}

type CreateProductParams struct {
	Name          string
	Description   string
	CategoryID    uuid.UUID
	PriceUnits    int64
	PriceCurrency string
	Sku           string
	Active        bool
	StockQuantity int32
}

func (q *Queries) CreateProduct(ctx context.Context, arg CreateProductParams) (uuid.UUID, error) {
	return uuid.New(), nil
}

type ListProductsParams struct {
	ActiveOnly bool
	CategoryID string
	Limit      int32
	Offset     int32
}

type ProductRow struct {
	ID            uuid.UUID
	Name          string
	Description   string
	CategoryID    uuid.UUID
	PriceUnits    int64
	PriceCurrency string
	Sku           string
	Active        bool
	StockQuantity int32
	CreatedAt     any
	UpdatedAt     any
}

func (q *Queries) ListProducts(ctx context.Context, arg ListProductsParams) ([]ProductRow, error) {
	return []ProductRow{}, nil
}

func (q *Queries) GetProduct(ctx context.Context, id string) (ProductRow, error) {
	return ProductRow{}, nil
}

type UpdateProductParams struct {
	ID          uuid.UUID
	Name        string
	Description string
	CategoryID  uuid.UUID
	PriceUnits  int64
	Active      bool
}

func (q *Queries) UpdateProduct(ctx context.Context, arg UpdateProductParams) (ProductRow, error) {
	return ProductRow{}, nil
}

func (q *Queries) DeleteProduct(ctx context.Context, id string) error {
	return nil
}

type SearchProductsParams struct {
	Query       string
	CategoryID  string
	MinPrice    int64
	MaxPrice    int64
	InStockOnly bool
	Limit       int32
	Offset      int32
}

func (q *Queries) SearchProducts(ctx context.Context, arg SearchProductsParams) ([]ProductRow, error) {
	return []ProductRow{}, nil
}

type ProductVariantRow struct {
	ID            uuid.UUID
	ProductID     uuid.UUID
	Name          string
	PriceUnits    int64
	PriceCurrency string
	Sku           string
	StockQuantity int32
}

func (q *Queries) GetProductVariants(ctx context.Context, productID string) ([]ProductVariantRow, error) {
	return []ProductVariantRow{}, nil
}
