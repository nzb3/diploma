package pgx

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nzb3/diploma/search-service/database/sqlc"
	"github.com/nzb3/diploma/search-service/internal/domain/models/eventmodel"
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

// CreateEvent creates a new event in the database
func (r *Repository) CreateEvent(ctx context.Context, event eventmodel.Event) (eventmodel.Event, error) {
	const op = "EventRepository.CreateEvent"

	params := sqlc.CreateEventParams{
		Name:    event.Name,
		Topic:   event.Topic,
		Payload: event.Payload,
	}

	row, err := r.queries.CreateEvent(ctx, params)
	if err != nil {
		return eventmodel.Event{}, fmt.Errorf("%s: failed to create event: %w", op, err)
	}

	return eventmodel.Event{
		ID:        PgTypeToUUID(row.ID),
		Name:      row.Name,
		Topic:     row.Topic,
		Payload:   row.Payload,
		Sent:      row.Sent,
		EventTime: PgTypeToTime(row.EventTime),
	}, nil
}

// GetNotSentEvents retrieves events that haven't been sent yet
func (r *Repository) GetNotSentEvents(ctx context.Context, limit int, offset int) ([]eventmodel.Event, error) {
	const op = "EventRepository.GetNotSentEvents"

	params := sqlc.GetNotSentEventsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	rows, err := r.queries.GetNotSentEvents(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get unsent events: %w", op, err)
	}

	events := make([]eventmodel.Event, len(rows))
	for i, row := range rows {
		events[i] = eventmodel.Event{
			ID:        PgTypeToUUID(row.ID),
			Name:      row.Name,
			Topic:     row.Topic,
			Payload:   row.Payload,
			Sent:      row.Sent,
			EventTime: PgTypeToTime(row.EventTime),
		}
	}

	return events, nil
}

// MarkEventAsSent marks an event as successfully sent
func (r *Repository) MarkEventAsSent(ctx context.Context, eventID uuid.UUID) error {
	const op = "EventRepository.MarkEventAsSent"

	pgUUID := UuidToPgType(eventID)
	err := r.queries.MarkEventAsSent(ctx, pgUUID)
	if err != nil {
		return fmt.Errorf("%s: failed to mark event as sent: %w", op, err)
	}

	return nil
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

// Helper functions for type conversion
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

func PgTypeToTime(pgtime pgtype.Timestamp) time.Time {
	if !pgtime.Valid {
		return time.Time{}
	}
	return pgtime.Time
}

func TimeToPgType(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{
		Time:  t,
		Valid: !t.IsZero(),
	}
}
