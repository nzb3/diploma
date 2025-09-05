package searchservice

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nzb3/diploma/search-service/internal/domain/models"
)

type SearchOption func(*SearchOptions)

type SearchOptions struct {
	NumberOfReferences int
}

func WithNumberOfReferences(n int) SearchOption {
	return func(o *SearchOptions) {
		o.NumberOfReferences = n
	}
}

type vectorStorage interface {
	GetAnswer(ctx context.Context, question string) (string, []models.Reference, error)
	GetAnswerStream(ctx context.Context, question string, opts ...SearchOption) (<-chan string, <-chan []models.Reference, <-chan []byte, <-chan error)
	SemanticSearch(ctx context.Context, query string) ([]models.Reference, error)
}

type eventPublisher interface {
	PublishEvent(ctx context.Context, topic string, eventName string, data interface{}) error
}

type Service struct {
	vectorStorage  vectorStorage
	eventPublisher eventPublisher // Optional event publisher
}

// NewService creates a new search service with optional event publisher
func NewService(vs vectorStorage, eventPublisher ...eventPublisher) *Service {
	slog.Debug("Initializing search service",
		"vector_storage_type", fmt.Sprintf("%T", vs),
		"repository_type", fmt.Sprintf("%T"))

	service := &Service{vectorStorage: vs}
	if len(eventPublisher) > 0 {
		service.eventPublisher = eventPublisher[0]
		slog.Debug("Event publisher configured for search service")
	}
	return service
}

func (s *Service) GetAnswerStream(
	ctx context.Context,
	question string,
	numReferences int,
) (
	<-chan models.SearchResult,
	<-chan []models.Reference,
	<-chan []byte,
	<-chan error,
) {
	const op = "Service.GetAnswerStream"

	errOutputCh := make(chan error, 1)
	refsOutputCh := make(chan []models.Reference)
	searchResultOutputCh := make(chan models.SearchResult)

	answerCh, refsCh, chunkCh, getAnswerErrCh := s.vectorStorage.GetAnswerStream(
		ctx,
		question,
		WithNumberOfReferences(numReferences),
	)

	go func() {
		defer func() {
			close(refsOutputCh)
			close(errOutputCh)
			close(searchResultOutputCh)
		}()

		processedRefsCh := make(chan []models.Reference, 1)
		defer close(processedRefsCh)

		for {
			select {
			case refs := <-refsCh:
				processedRefsCh <- refs
				refsOutputCh <- refs
			case <-ctx.Done():
				slog.Debug("Context cancelled")
				errOutputCh <- ctx.Err()
				return
			case err := <-getAnswerErrCh:
				slog.Error("Error getting answer stream", "err", err)
				errOutputCh <- fmt.Errorf("%s: %w", op, err)
				return
			case answer := <-answerCh:
				slog.Info("Processing answer", "question", question)

				searchResult := models.SearchResult{
					Answer:     answer,
					References: <-processedRefsCh,
				}

				searchResultOutputCh <- searchResult
				return
			}
		}
	}()

	return searchResultOutputCh, refsOutputCh, chunkCh, errOutputCh
}

func (s *Service) GetAnswer(ctx context.Context, question string) (models.SearchResult, error) {
	const op = "Service.GetAnswer"
	slog.InfoContext(ctx, "Getting answer",
		"question", question)

	answer, refs, err := s.vectorStorage.GetAnswer(ctx, question)
	if err != nil {
		slog.Error("Error getting answer", "err", err)
		return models.SearchResult{}, fmt.Errorf("%s: %w", op, err)
	}

	result := models.SearchResult{
		Answer:     answer,
		References: refs,
	}

	// Publish search event if event publisher is available
	if s.eventPublisher != nil {
		searchEvent := map[string]interface{}{
			"question":         question,
			"answer_length":    len(answer),
			"references_count": len(refs),
			"operation":        "get_answer",
		}
		if err := s.eventPublisher.PublishEvent(ctx, "search", "search.performed", searchEvent); err != nil {
			slog.WarnContext(ctx, "Failed to publish search event", "error", err)
			// Don't fail the main operation if event publishing fails
		}
	}

	return result, nil
}

func (s *Service) SemanticSearch(ctx context.Context, query string) ([]models.Reference, error) {
	const op = "Service.SemanticSearch"
	slog.InfoContext(ctx, "Performing semantic search",
		"query", query)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		references, err := s.vectorStorage.SemanticSearch(ctx, query)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to perform semantic search",
				"op", op,
				"error", err)
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		slog.InfoContext(ctx, "Semantic search completed",
			"references_count", len(references))

		// Publish semantic search event if event publisher is available
		if s.eventPublisher != nil {
			searchEvent := map[string]interface{}{
				"query":            query,
				"references_count": len(references),
				"operation":        "semantic_search",
			}
			if err := s.eventPublisher.PublishEvent(ctx, "search", "search.semantic_performed", searchEvent); err != nil {
				slog.WarnContext(ctx, "Failed to publish semantic search event", "error", err)
				// Don't fail the main operation if event publishing fails
			}
		}

		return references, nil
	}
}
