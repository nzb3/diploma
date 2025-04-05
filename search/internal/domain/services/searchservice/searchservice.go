package searchservice

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/nzb3/diploma/search/internal/domain/models"
)

type vectorStorage interface {
	GetAnswer(ctx context.Context, question string, refsChan chan<- []models.Reference) (models.SearchResult, error)
	GetAnswerStream(ctx context.Context, question string, refsChan chan<- []models.Reference, chunkChan chan<- []byte) (models.SearchResult, error)
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

func (ss *Service) GetAnswerStream(ctx context.Context, question string) (<-chan models.SearchResult, <-chan []models.Reference, <-chan []byte, <-chan error) {
	const op = "Service.GetAnswerStream"
	slog.InfoContext(ctx, "Starting answer stream",
		"question", question)

	chunks := make(chan []byte, 100)
	results := make(chan models.SearchResult)
	refs := make(chan []models.Reference, 1)
	defer close(refs)
	errs := make(chan error)

	go func() {
		defer func() {
			close(chunks)
			close(results)
			close(errs)
		}()

		select {
		case <-ctx.Done():
			errs <- ctx.Err()
		default:
			slog.DebugContext(ctx, "Requesting answer stream from vector storage")
			result, err := ss.vectorStorage.GetAnswerStream(ctx, question, refs, chunks)
			if err != nil {
				slog.ErrorContext(ctx, "Failed to get answer stream", "op", op, "error", err)
				errs <- fmt.Errorf("%s: %w", op, err)
				return
			}

			slog.DebugContext(ctx, "Processing stream result")
			result, err = ss.processResult(ctx, result)
			if err != nil {
				slog.ErrorContext(ctx, "Failed to process result",
					"op", op,
					"error", err)
				errs <- fmt.Errorf("%s: %w", op, err)
				return
			}

			slog.InfoContext(ctx, "Sending final result",
				"references_count", len(result.References))
			results <- result
		}
	}()

	processedRefs, err := ss.provideReferencesWithResourceID(ctx, <-refs)
	if err != nil {
		errs <- fmt.Errorf("%s: %w", op, err)
	}

	refs <- processedRefs

	return results, refs, chunks, errs
}

func (ss *Service) GetAnswer(ctx context.Context, question string, refsChan chan<- []models.Reference) (models.SearchResult, error) {
	const op = "Service.GetAnswer"
	slog.InfoContext(ctx, "Getting answer",
		"question", question)
	select {
	case <-ctx.Done():
		return models.SearchResult{}, ctx.Err()
	default:
		result, err := ss.vectorStorage.GetAnswer(ctx, question, refsChan)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to get answer from vector storage",
				"op", op,
				"error", err)
			return models.SearchResult{}, fmt.Errorf("%s: %w", op, err)
		}

		slog.DebugContext(ctx, "Processing answer result",
			"references_count", len(result.References))
		result, err = ss.processResult(ctx, result)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to process result",
				"op", op,
				"error", err)
			return models.SearchResult{}, fmt.Errorf("%s: %w", op, err)
		}

		slog.InfoContext(ctx, "Successfully retrieved answer",
			"references_count", len(result.References))
		return result, nil
	}
}

func (ss *Service) SemanticSearch(ctx context.Context, query string) ([]models.Reference, error) {
	const op = "Service.SemanticSearch"
	slog.InfoContext(ctx, "Performing semantic search",
		"query", query)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		references, err := ss.vectorStorage.SemanticSearch(ctx, query)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to perform semantic search",
				"op", op,
				"error", err)
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		slog.DebugContext(ctx, "Adding resource IDs to references",
			"references_count", len(references))
		references, err = ss.provideReferencesWithResourceID(ctx, references)
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

func (ss *Service) processResult(ctx context.Context, result models.SearchResult) (models.SearchResult, error) {
	const op = "Service.processResult"
	slog.DebugContext(ctx, "Processing search result",
		"references_count", len(result.References))
	select {
	case <-ctx.Done():
		return models.SearchResult{}, ctx.Err()
	default:
		refs, err := ss.provideReferencesWithResourceID(ctx, result.References)
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

func (ss *Service) provideReferencesWithResourceID(ctx context.Context, refs []models.Reference) ([]models.Reference, error) {
	const op = "Service.provideReferencesWithResourceID"
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

			resID, err := ss.repository.GetResourceIDByReference(ctx, ref)
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
