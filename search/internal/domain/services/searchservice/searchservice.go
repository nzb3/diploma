package searchservice

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/nzb3/diploma/search/internal/domain/models"
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

type repository interface {
	GetResourceIDByReference(ctx context.Context, reference models.Reference) (uuid.UUID, error)
}

type Service struct {
	vectorStorage vectorStorage
	repository    repository
}

func NewService(vs vectorStorage, r repository) *Service {
	slog.Debug("Initializing search service",
		"vector_storage_type", fmt.Sprintf("%T", vs),
		"repository_type", fmt.Sprintf("%T", r))
	return &Service{vectorStorage: vs, repository: r}
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

	answerCh, rawRefsCh, chunkCh, getAnswerErrCh := s.vectorStorage.GetAnswerStream(
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
			case refs := <-rawRefsCh:
				processedRefs, err := s.processReferences(ctx, refs)
				if err != nil {
					slog.Error("Error processing references", "err", err)
					errOutputCh <- fmt.Errorf("%s: %w", op, err)
					return
				}
				processedRefsCh <- processedRefs
				refsOutputCh <- processedRefs
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

	answer, rawRefs, err := s.vectorStorage.GetAnswer(ctx, question)
	if err != nil {
		slog.Error("Error getting answer", "err", err)
		return models.SearchResult{}, fmt.Errorf("%s: %w", op, err)
	}

	processedRefs, err := s.processReferences(ctx, rawRefs)
	if err != nil {
		slog.Error("Error processing references", "err", err)
		return models.SearchResult{}, fmt.Errorf("%s: %w", op, err)
	}

	return models.SearchResult{
		Answer:     answer,
		References: processedRefs,
	}, nil
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

		slog.DebugContext(ctx, "Adding resource IDs to references",
			"references_count", len(references))
		references, err = s.processReferences(ctx, references)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to provide references with resource IDs",
				"op", op,
				"error", err)
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		slog.InfoContext(ctx, "Semantic search completed",
			"references_count", len(references))
		return references, nil
	}
}

func (s *Service) processResult(ctx context.Context, result models.SearchResult) (models.SearchResult, error) {
	const op = "Service.processResult"
	slog.DebugContext(ctx, "Processing search result",
		"references_count", len(result.References))
	select {
	case <-ctx.Done():
		return models.SearchResult{}, ctx.Err()
	default:
		refs, err := s.processReferences(ctx, result.References)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to provide references with resource IDs",
				"op", op,
				"error", err)
			return models.SearchResult{}, fmt.Errorf("%s: %w", op, err)
		}

		result.References = refs
		slog.DebugContext(ctx, "Result processing completed")
		return result, nil
	}
}

func (s *Service) processReferences(ctx context.Context, refs []models.Reference) ([]models.Reference, error) {
	const op = "Service.processReferences"
	slog.DebugContext(ctx, "Adding resource IDs to references",
		"references_count", len(refs))

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		for i, ref := range refs {
			slog.DebugContext(ctx, "Getting resource ID for reference",
				"reference_index", i,
				"content_length", len(ref.Content))

			resID, err := s.repository.GetResourceIDByReference(ctx, ref)
			if err != nil {
				slog.ErrorContext(ctx, "Failed to get resource ID for reference",
					"op", op,
					"reference_index", i,
					"error", err)
				return nil, fmt.Errorf("%s: %w", op, err)
			}

			refs[i].ResourceID = resID
			slog.DebugContext(ctx, "Added resource ID to reference",
				"reference_index", i,
				"resource_id", resID)
		}

		slog.DebugContext(ctx, "Successfully added resource IDs to all references",
			"references_count", len(refs))
		return refs, nil
	}
}
