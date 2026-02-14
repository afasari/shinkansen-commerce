package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type StockItem struct {
	ID                uuid.UUID
	ProductID         uuid.UUID
	VariantID         *uuid.UUID
	WarehouseID       uuid.UUID
	Quantity          int
	ReservedQuantity  int
	AvailableQuantity int
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type StockMovement struct {
	ID           uuid.UUID
	StockItemID  uuid.UUID
	MovementType string
	Quantity     int
	Reference    *string
	CreatedAt    time.Time
}

type StockReservation struct {
	ID          uuid.UUID
	OrderID     uuid.UUID
	StockItemID uuid.UUID
	Quantity    int
	ExpiresAt   time.Time
	CreatedAt   time.Time
}

type CreateStockItemParams struct {
	ProductID   uuid.UUID
	VariantID   *uuid.UUID
	WarehouseID uuid.UUID
	Quantity    int
}

type UpdateStockParams struct {
	ProductID   uuid.UUID
	VariantID   *uuid.UUID
	WarehouseID uuid.UUID
	Delta       int
}

type ReserveStockParams struct {
	OrderID     uuid.UUID
	StockItemID uuid.UUID
	Quantity    int
	ExpiresAt   time.Time
}

type ReleaseStockParams struct {
	OrderID uuid.UUID
}

type Querier interface {
	CreateStockItem(ctx context.Context, arg CreateStockItemParams) (uuid.UUID, error)
	GetStock(ctx context.Context, productID, variantID, warehouseID uuid.UUID) (StockItem, error)
	GetStockByID(ctx context.Context, id uuid.UUID) (StockItem, error)
	UpdateStockQuantity(ctx context.Context, arg UpdateStockParams) error
	ReserveStock(ctx context.Context, arg ReserveStockParams) error
	ReleaseStock(ctx context.Context, arg ReleaseStockParams) error
	CreateStockMovement(ctx context.Context, arg CreateStockMovementParams) error
	ListStockMovements(ctx context.Context, stockItemID uuid.UUID, limit, offset int) ([]StockMovement, error)
}

type Queries struct {
	db *DB
}

func NewQueries(db *DB) *Queries {
	return &Queries{db: db}
}

func (q *Queries) CreateStockItem(ctx context.Context, arg CreateStockItemParams) (uuid.UUID, error) {
	const sql = `
		INSERT INTO inventory.stock_items (product_id, variant_id, warehouse_id, quantity, reserved_quantity, available_quantity, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 0, $4, NOW(), NOW())
		ON CONFLICT (product_id, variant_id, warehouse_id) 
		DO UPDATE SET quantity = EXCLUDED.quantity + $4
		RETURNING id
	`
	row := q.db.pool.QueryRow(ctx, sql, arg.ProductID, arg.VariantID, arg.WarehouseID, arg.Quantity)
	var id uuid.UUID
	err := row.Scan(&id)
	return id, err
}

func (q *Queries) GetStock(ctx context.Context, productID, variantID, warehouseID uuid.UUID) (StockItem, error) {
	const sql = `
		SELECT id, product_id, variant_id, warehouse_id, quantity, reserved_quantity, available_quantity, created_at, updated_at
		FROM inventory.stock_items
		WHERE product_id = $1 AND variant_id = $2 AND warehouse_id = $3
	`
	row := q.db.pool.QueryRow(ctx, sql, productID, variantID, warehouseID)
	var s StockItem
	err := row.Scan(
		&s.ID, &s.ProductID, &s.VariantID, &s.WarehouseID,
		&s.Quantity, &s.ReservedQuantity, &s.AvailableQuantity,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return StockItem{}, nil
	}
	return s, err
}

func (q *Queries) GetStockByID(ctx context.Context, id uuid.UUID) (StockItem, error) {
	const sql = `
		SELECT id, product_id, variant_id, warehouse_id, quantity, reserved_quantity, available_quantity, created_at, updated_at
		FROM inventory.stock_items
		WHERE id = $1
	`
	row := q.db.pool.QueryRow(ctx, sql, id)
	var s StockItem
	err := row.Scan(
		&s.ID, &s.ProductID, &s.VariantID, &s.WarehouseID,
		&s.Quantity, &s.ReservedQuantity, &s.AvailableQuantity,
		&s.CreatedAt, &s.UpdatedAt,
	)
	return s, err
}

func (q *Queries) UpdateStockQuantity(ctx context.Context, arg UpdateStockParams) error {
	const sql = `
		INSERT INTO inventory.stock_items (product_id, variant_id, warehouse_id, quantity, reserved_quantity, available_quantity, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 0, $4, NOW(), NOW())
		ON CONFLICT (product_id, variant_id, warehouse_id) 
		DO UPDATE SET 
			quantity = GREATEST(0, inventory.stock_items.quantity + $4),
			updated_at = NOW()
		WHERE id IN (
			SELECT id FROM inventory.stock_items 
			WHERE product_id = $1 AND variant_id = $2 AND warehouse_id = $3 
			FOR UPDATE
		)
		RETURNING id
	`
	_, err := q.db.pool.Exec(ctx, sql, arg.ProductID, arg.VariantID, arg.WarehouseID, arg.Delta)
	return err
}

func (q *Queries) ReserveStock(ctx context.Context, arg ReserveStockParams) error {
	tx, err := q.db.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const updateSQL = `
		UPDATE inventory.stock_items
		SET reserved_quantity = reserved_quantity + $1,
		    updated_at = NOW()
		WHERE id = $2 AND available_quantity >= $1
		RETURNING id
	`
	var stockItemID uuid.UUID
	err = tx.QueryRow(ctx, updateSQL, arg.Quantity, arg.StockItemID).Scan(&stockItemID)
	if err != nil {
		return err
	}

	const insertSQL = `
		INSERT INTO inventory.stock_reservations (order_id, stock_item_id, quantity, expires_at, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (order_id, stock_item_id) DO UPDATE SET
			quantity = stock_reservations.quantity + $3,
			expires_at = $4
	`
	_, err = tx.Exec(ctx, insertSQL, arg.OrderID, arg.StockItemID, arg.Quantity, arg.ExpiresAt)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (q *Queries) ReleaseStock(ctx context.Context, arg ReleaseStockParams) error {
	const sql = `
		WITH released AS (
			SELECT sr.stock_item_id, sr.quantity
			FROM inventory.stock_reservations sr
			WHERE sr.order_id = $1
		)
		UPDATE inventory.stock_items si
		SET 
			reserved_quantity = si.reserved_quantity - released.quantity,
			updated_at = NOW()
		FROM released
		WHERE si.id = released.stock_item_id
	`
	_, err := q.db.pool.Exec(ctx, sql, arg.OrderID)
	if err != nil {
		return err
	}

	const deleteSQL = `DELETE FROM inventory.stock_reservations WHERE order_id = $1`
	_, err = q.db.pool.Exec(ctx, deleteSQL, arg.OrderID)
	return err
}

type CreateStockMovementParams struct {
	StockItemID  uuid.UUID
	MovementType string
	Quantity     int
	Reference    *string
}

func (q *Queries) CreateStockMovement(ctx context.Context, arg CreateStockMovementParams) error {
	const sql = `
		INSERT INTO inventory.stock_movements (stock_item_id, movement_type, quantity, reference, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`
	_, err := q.db.pool.Exec(ctx, sql, arg.StockItemID, arg.MovementType, arg.Quantity, arg.Reference)
	return err
}

func (q *Queries) ListStockMovements(ctx context.Context, stockItemID uuid.UUID, limit, offset int) ([]StockMovement, error) {
	const sql = `
		SELECT id, stock_item_id, movement_type, quantity, reference, created_at
		FROM inventory.stock_movements
		WHERE stock_item_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := q.db.pool.Query(ctx, sql, stockItemID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movements []StockMovement
	for rows.Next() {
		var m StockMovement
		err := rows.Scan(
			&m.ID, &m.StockItemID, &m.MovementType,
			&m.Quantity, &m.Reference, &m.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		movements = append(movements, m)
	}
	return movements, rows.Err()
}
