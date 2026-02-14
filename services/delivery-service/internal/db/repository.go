package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, databaseURL string) (*DB, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	return &DB{pool: pool}, nil
}

func (db *DB) Close() error {
	db.pool.Close()
	return nil
}

type DeliverySlot struct {
	ID             uuid.UUID
	DeliveryZoneID uuid.UUID
	StartTime      time.Time
	EndTime        time.Time
	Capacity       int
	Reserved       int
	Available      int
	Date           time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Shipment struct {
	ID                  uuid.UUID
	OrderID             uuid.UUID
	TrackingNumber      *string
	Status              string
	EstimatedDeliveryAt *time.Time
	ActualDeliveryAt    *time.Time
	Carrier             string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type DeliveryReservation struct {
	ID        uuid.UUID
	SlotID    uuid.UUID
	OrderID   uuid.UUID
	CreatedAt time.Time
}

type Querier interface {
	GetDeliverySlots(ctx context.Context, deliveryZoneID uuid.UUID, date time.Time) ([]DeliverySlot, error)
	GetDeliverySlot(ctx context.Context, id uuid.UUID) (DeliverySlot, error)
	ReserveDeliverySlot(ctx context.Context, slotID, orderID uuid.UUID) (uuid.UUID, error)
	GetShipmentByOrderID(ctx context.Context, orderID uuid.UUID) (Shipment, error)
	GetShipment(ctx context.Context, id uuid.UUID) (Shipment, error)
	CreateShipment(ctx context.Context, orderID uuid.UUID) (uuid.UUID, error)
	UpdateShipmentStatus(ctx context.Context, id uuid.UUID, status string) error
	ReleaseDeliverySlot(ctx context.Context, orderID uuid.UUID) error
}

type Queries struct {
	db *DB
}

func NewQueries(db *DB) *Queries {
	return &Queries{db: db}
}

func (q *Queries) GetDeliverySlots(ctx context.Context, deliveryZoneID uuid.UUID, date time.Time) ([]DeliverySlot, error) {
	const sql = `
		SELECT id, delivery_zone_id, start_time, end_time, capacity, reserved, available, created_at, updated_at, date
		FROM delivery.delivery_slots
		WHERE delivery_zone_id = $1 AND date = $2 AND available > 0
		ORDER BY start_time ASC
	`
	rows, err := q.db.pool.Query(ctx, sql, deliveryZoneID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slots []DeliverySlot
	for rows.Next() {
		var s DeliverySlot
		err := rows.Scan(
			&s.ID, &s.DeliveryZoneID, &s.StartTime, &s.EndTime,
			&s.Capacity, &s.Reserved, &s.Available,
			&s.CreatedAt, &s.UpdatedAt, &s.Date,
		)
		if err != nil {
			return nil, err
		}
		slots = append(slots, s)
	}
	return slots, rows.Err()
}

func (q *Queries) GetDeliverySlot(ctx context.Context, id uuid.UUID) (DeliverySlot, error) {
	const sql = `
		SELECT id, delivery_zone_id, start_time, end_time, capacity, reserved, available, created_at, updated_at, date
		FROM delivery.delivery_slots
		WHERE id = $1
	`
	row := q.db.pool.QueryRow(ctx, sql, id)
	var s DeliverySlot
	err := row.Scan(
		&s.ID, &s.DeliveryZoneID, &s.StartTime, &s.EndTime,
		&s.Capacity, &s.Reserved, &s.Available,
		&s.CreatedAt, &s.UpdatedAt, &s.Date,
	)
	return s, err
}

func (q *Queries) ReserveDeliverySlot(ctx context.Context, slotID, orderID uuid.UUID) (uuid.UUID, error) {
	tx, err := q.db.pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const updateSQL = `
		UPDATE delivery.delivery_slots
		SET reserved = reserved + 1, updated_at = NOW()
		WHERE id = $1 AND available > 0
		RETURNING id
	`
	var updatedSlotID uuid.UUID
	err = tx.QueryRow(ctx, updateSQL, slotID).Scan(&updatedSlotID)
	if err != nil {
		return uuid.Nil, err
	}

	const insertSQL = `
		INSERT INTO delivery.delivery_reservations (slot_id, order_id, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (order_id) DO NOTHING
		RETURNING id
	`
	var reservationID uuid.UUID
	err = tx.QueryRow(ctx, insertSQL, slotID, orderID).Scan(&reservationID)
	if err != nil {
		return uuid.Nil, err
	}

	return reservationID, tx.Commit(ctx)
}

func (q *Queries) GetShipmentByOrderID(ctx context.Context, orderID uuid.UUID) (Shipment, error) {
	const sql = `
		SELECT id, order_id, tracking_number, status, estimated_delivery_at, actual_delivery_at, carrier, created_at, updated_at
		FROM delivery.shipments
		WHERE order_id = $1
	`
	row := q.db.pool.QueryRow(ctx, sql, orderID)
	var s Shipment
	err := row.Scan(
		&s.ID, &s.OrderID, &s.TrackingNumber, &s.Status,
		&s.EstimatedDeliveryAt, &s.ActualDeliveryAt, &s.Carrier,
		&s.CreatedAt, &s.UpdatedAt,
	)
	return s, err
}

func (q *Queries) GetShipment(ctx context.Context, id uuid.UUID) (Shipment, error) {
	const sql = `
		SELECT id, order_id, tracking_number, status, estimated_delivery_at, actual_delivery_at, carrier, created_at, updated_at
		FROM delivery.shipments
		WHERE id = $1
	`
	row := q.db.pool.QueryRow(ctx, sql, id)
	var s Shipment
	err := row.Scan(
		&s.ID, &s.OrderID, &s.TrackingNumber, &s.Status,
		&s.EstimatedDeliveryAt, &s.ActualDeliveryAt, &s.Carrier,
		&s.CreatedAt, &s.UpdatedAt,
	)
	return s, err
}

func (q *Queries) CreateShipment(ctx context.Context, orderID uuid.UUID) (uuid.UUID, error) {
	const sql = `
		INSERT INTO delivery.shipments (order_id, status, created_at, updated_at)
		VALUES ($1, 'SHIPMENT_STATUS_PREPARING', NOW(), NOW())
		RETURNING id
	`
	row := q.db.pool.QueryRow(ctx, sql, orderID)
	var id uuid.UUID
	err := row.Scan(&id)
	return id, err
}

func (q *Queries) UpdateShipmentStatus(ctx context.Context, id uuid.UUID, status string) error {
	const sql = `
		UPDATE delivery.shipments
		SET status = $2, updated_at = NOW()
		WHERE id = $1
	`
	_, err := q.db.pool.Exec(ctx, sql, id, status)
	return err
}

func (q *Queries) ReleaseDeliverySlot(ctx context.Context, orderID uuid.UUID) error {
	const sql = `
		WITH released AS (
			SELECT dr.slot_id
			FROM delivery.delivery_reservations dr
			WHERE dr.order_id = $1
		)
		UPDATE delivery.delivery_slots ds
		SET reserved = GREATEST(0, reserved - 1), updated_at = NOW()
		FROM released
		WHERE ds.id = released.slot_id
	`
	_, err := q.db.pool.Exec(ctx, sql, orderID)
	if err != nil {
		return err
	}

	deleteSQL := `DELETE FROM delivery.delivery_reservations WHERE order_id = $1`
	_, err = q.db.pool.Exec(ctx, deleteSQL, orderID)
	return err
}
