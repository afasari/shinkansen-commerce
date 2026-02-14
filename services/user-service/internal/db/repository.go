package db

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	Name         string
	Phone        string
	Active       bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Address struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	Name         string
	Phone        string
	PostalCode   string
	Prefecture   string
	City         string
	AddressLine1 string
	AddressLine2 string
	IsDefault    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CreateUserParams struct {
	Email        string
	PasswordHash string
	Name         string
	Phone        string
}

type UpdateUserParams struct {
	ID     uuid.UUID
	Name   *string
	Phone  *string
	Active *bool
}

type CreateAddressParams struct {
	UserID       uuid.UUID
	Name         string
	Phone        string
	PostalCode   string
	Prefecture   string
	City         string
	AddressLine1 string
	AddressLine2 string
	IsDefault    bool
}

type UpdateAddressParams struct {
	ID           uuid.UUID
	Name         *string
	Phone        *string
	PostalCode   *string
	Prefecture   *string
	City         *string
	AddressLine1 *string
	AddressLine2 *string
	IsDefault    *bool
}

type DeleteAddressParams struct {
	ID uuid.UUID
}

type GetDefaultAddressParams struct {
	UserID uuid.UUID
}

type UpdateDefaultAddressParams struct {
	ID uuid.UUID
}

type SetDefaultAddressParams struct {
	ID     uuid.UUID
	UserID uuid.UUID
}

type Querier interface {
	CreateUser(ctx context.Context, arg CreateUserParams) (uuid.UUID, error)
	GetUser(ctx context.Context, id uuid.UUID) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	UpdateUser(ctx context.Context, arg UpdateUserParams) error
	CreateAddress(ctx context.Context, arg CreateAddressParams) (uuid.UUID, error)
	ListAddresses(ctx context.Context, userID uuid.UUID) ([]Address, error)
	GetAddress(ctx context.Context, id uuid.UUID) (Address, error)
	UpdateAddress(ctx context.Context, arg UpdateAddressParams) error
	DeleteAddress(ctx context.Context, arg DeleteAddressParams) error
	GetDefaultAddress(ctx context.Context, arg GetDefaultAddressParams) (Address, error)
	SetDefaultAddress(ctx context.Context, arg SetDefaultAddressParams) error
}

type Queries struct {
	db *DB
}

func NewQueries(db *DB) *Queries {
	return &Queries{db: db}
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (uuid.UUID, error) {
	const sql = `
		INSERT INTO users.users (email, password_hash, name, phone, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id
	`
	row := q.db.pool.QueryRow(ctx, sql, arg.Email, arg.PasswordHash, arg.Name, arg.Phone)
	var id uuid.UUID
	err := row.Scan(&id)
	return id, err
}

func (q *Queries) GetUser(ctx context.Context, id uuid.UUID) (User, error) {
	const sql = `
		SELECT id, email, password_hash, name, phone, active, created_at, updated_at
		FROM users.users
		WHERE id = $1
	`
	row := q.db.pool.QueryRow(ctx, sql, id)
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Phone, &u.Active, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
	const sql = `
		SELECT id, email, password_hash, name, phone, active, created_at, updated_at
		FROM users.users
		WHERE email = $1
	`
	row := q.db.pool.QueryRow(ctx, sql, email)
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Phone, &u.Active, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) error {
	const sql = `
		UPDATE users.users
		SET
			name = COALESCE($2, name),
			phone = COALESCE($3, phone),
			active = COALESCE($4, active),
			updated_at = NOW()
		WHERE id = $1
	`
	_, err := q.db.pool.Exec(ctx, sql, arg.ID, arg.Name, arg.Phone, arg.Active)
	return err
}

func (q *Queries) CreateAddress(ctx context.Context, arg CreateAddressParams) (uuid.UUID, error) {
	const sql = `
		INSERT INTO users.addresses (
			user_id, name, phone, postal_code, prefecture, city, address_line1, address_line2, is_default, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id
	`
	row := q.db.pool.QueryRow(ctx, sql,
		arg.UserID, arg.Name, arg.Phone, arg.PostalCode, arg.Prefecture, arg.City, arg.AddressLine1, arg.AddressLine2, arg.IsDefault)
	var id uuid.UUID
	err := row.Scan(&id)
	return id, err
}

func (q *Queries) ListAddresses(ctx context.Context, userID uuid.UUID) ([]Address, error) {
	const sql = `
		SELECT id, user_id, name, phone, postal_code, prefecture, city, address_line1, address_line2, is_default, created_at, updated_at
		FROM users.addresses
		WHERE user_id = $1
		ORDER BY is_default DESC, created_at ASC
	`
	rows, err := q.db.pool.Query(ctx, sql, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var addresses []Address
	for rows.Next() {
		var a Address
		err := rows.Scan(
			&a.ID, &a.UserID, &a.Name, &a.Phone, &a.PostalCode,
			&a.Prefecture, &a.City, &a.AddressLine1, &a.AddressLine2,
			&a.IsDefault, &a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, a)
	}
	return addresses, rows.Err()
}

func (q *Queries) GetAddress(ctx context.Context, id uuid.UUID) (Address, error) {
	const sql = `
		SELECT id, user_id, name, phone, postal_code, prefecture, city, address_line1, address_line2, is_default, created_at, updated_at
		FROM users.addresses
		WHERE id = $1
	`
	row := q.db.pool.QueryRow(ctx, sql, id)
	var a Address
	err := row.Scan(
		&a.ID, &a.UserID, &a.Name, &a.Phone, &a.PostalCode,
		&a.Prefecture, &a.City, &a.AddressLine1, &a.AddressLine2,
		&a.IsDefault, &a.CreatedAt, &a.UpdatedAt,
	)
	return a, err
}

func (q *Queries) UpdateAddress(ctx context.Context, arg UpdateAddressParams) error {
	const sql = `
		UPDATE users.addresses
		SET
			name = COALESCE($2, name),
			phone = COALESCE($3, phone),
			postal_code = COALESCE($4, postal_code),
			prefecture = COALESCE($5, prefecture),
			city = COALESCE($6, city),
			address_line1 = COALESCE($7, address_line1),
			address_line2 = COALESCE($8, address_line2),
			is_default = COALESCE($9, is_default),
			updated_at = NOW()
		WHERE id = $1
	`
	_, err := q.db.pool.Exec(ctx, sql,
		arg.ID, arg.Name, arg.Phone,
		arg.PostalCode, arg.Prefecture, arg.City, arg.AddressLine1, arg.AddressLine2, arg.IsDefault)
	return err
}

func (q *Queries) DeleteAddress(ctx context.Context, arg DeleteAddressParams) error {
	const sql = `
		UPDATE users.addresses
		SET is_default = false, updated_at = NOW()
		WHERE id = $1
	`
	_, err := q.db.pool.Exec(ctx, sql, arg.ID)
	return err
}

func (q *Queries) GetDefaultAddress(ctx context.Context, arg GetDefaultAddressParams) (Address, error) {
	const sql = `
		SELECT id, user_id, name, phone, postal_code, prefecture, city, address_line1, address_line2, is_default, created_at, updated_at
		FROM users.addresses
		WHERE user_id = $1 AND is_default = true
		LIMIT 1
	`
	row := q.db.pool.QueryRow(ctx, sql, arg.UserID)
	var a Address
	err := row.Scan(
		&a.ID, &a.UserID, &a.Name, &a.Phone, &a.PostalCode,
		&a.Prefecture, &a.City, &a.AddressLine1, &a.AddressLine2,
		&a.IsDefault, &a.CreatedAt, &a.UpdatedAt,
	)
	return a, err
}

func (q *Queries) SetDefaultAddress(ctx context.Context, arg SetDefaultAddressParams) error {
	const sql = `
		UPDATE users.addresses
		SET is_default = false, updated_at = NOW()
		WHERE user_id = $1
		`

	_, err1 := q.db.pool.Exec(ctx, sql, arg.UserID)

	if arg.ID != uuid.Nil {
		const sql2 = `
			UPDATE users.addresses
			SET is_default = true, updated_at = NOW()
			WHERE id = $1
		`
		_, err2 := q.db.pool.Exec(ctx, sql2, arg.ID)
		if err2 != nil {
			return err2
		}
	}
	return err1
}
