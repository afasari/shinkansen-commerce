package pgutil

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func ToPG(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: u,
		Valid: true,
	}
}

func ToPGPtr(u *uuid.UUID) *pgtype.UUID {
	if u == nil {
		return nil
	}
	return &pgtype.UUID{
		Bytes: *u,
		Valid: true,
	}
}

func ToPGFromString(s string) pgtype.UUID {
	if s == "" {
		return pgtype.UUID{Valid: false}
	}
	u, _ := uuid.Parse(s)
	return ToPG(u)
}

func ToPGPtrFromString(s string) *pgtype.UUID {
	if s == "" {
		return nil
	}
	u, _ := uuid.Parse(s)
	ret := ToPG(u)
	return &ret
}

func FromPG(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	return uuid.UUID(u.Bytes).String()
}

func FromPGPtr(u *pgtype.UUID) string {
	if u == nil || !u.Valid {
		return ""
	}
	return uuid.UUID(u.Bytes).String()
}
