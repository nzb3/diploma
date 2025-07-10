package events

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/lo"

	"github.com/nzb3/diploma/resource-service/database/sqlc"
	"github.com/nzb3/diploma/resource-service/internal/domain/models/eventmodel"
	"github.com/nzb3/diploma/resource-service/internal/repository/pgx"
)

type baseRepository interface {
	Close()
	DB() *pgxpool.Pool
	Queries() *sqlc.Queries
	Health(ctx context.Context) error
}

type Repository struct {
	baseRepository
}

func NewEventRepository(ctx context.Context, repository baseRepository) *Repository {
	return &Repository{
		baseRepository: repository,
	}
}

// GetNotSentEvents retrieves all events that have not been sent
func (r *Repository) GetNotSentEvents(ctx context.Context, limit int, offset int) ([]eventmodel.Event, error) {
	sqlcEvents, err := r.Queries().GetNotSentEvents(ctx, sqlc.GetNotSentEventsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	return lo.Map(sqlcEvents, func(sqlcEvent sqlc.Events, _ int) eventmodel.Event {
		return sqlcEventToModel(sqlcEvent)
	}), nil
}

// CreateEvent saves a new event to the database
func (r *Repository) CreateEvent(ctx context.Context, event eventmodel.Event) (eventmodel.Event, error) {
	params := sqlc.CreateEventParams{
		Name:    event.Name,
		Topic:   event.Topic,
		Payload: event.Payload,
	}

	sqlcEvent, err := r.Queries().CreateEvent(ctx, params)
	if err != nil {
		return eventmodel.Event{}, err
	}

	return sqlcEventToModel(sqlcEvent), nil
}

func sqlcEventToModel(sqlcEvent sqlc.Events) eventmodel.Event {
	return eventmodel.Event{
		ID:        pgx.PgTypeToUUID(sqlcEvent.ID),
		Name:      sqlcEvent.Name,
		Topic:     sqlcEvent.Topic,
		Payload:   sqlcEvent.Payload,
		Sent:      sqlcEvent.Sent,
		EventTime: sqlcEvent.EventTime.Time,
	}
}
