package pgx

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nzb3/diploma/resource-service/database/sqlc"
)

type Repository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

func NewRepository(ctx context.Context, pool *pgxpool.Pool) (*Repository, error) {
	queries := sqlc.New(pool)

	return &Repository{
		db:      pool,
		queries: queries,
	}, nil
}

// Close closes the database connection pool
func (r *Repository) Close() {
	if r.db != nil {
		r.db.Close()
	}
}

// DB returns the underlying database connection pool
func (r *Repository) DB() *pgxpool.Pool {
	return r.db
}

// Queries returns the sqlc queries instance
func (r *Repository) Queries() *sqlc.Queries {
	return r.queries
}

// Health checks if the database connection is healthy
func (r *Repository) Health(ctx context.Context) error {
	return r.db.Ping(ctx)
}

func StringToPgType(s string) pgtype.Text {
	return pgtype.Text{
		String: s,
		Valid:  s != "",
	}
}

func UuidToPgType(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: id,
		Valid: id != uuid.Nil,
	}
}

func PgTypeToUUID(pguuid pgtype.UUID) uuid.UUID {
	if !pguuid.Valid {
		return uuid.Nil
	}
	return pguuid.Bytes
}

func PgTypeToString(pgtext pgtype.Text) string {
	if !pgtext.Valid {
		return ""
	}
	return pgtext.String
}
