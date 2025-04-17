package vectorstorage

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/samber/lo"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/pgvector"

	"github.com/nzb3/diploma/search/internal/domain/models"
	"github.com/nzb3/diploma/search/internal/repository/vectorstorage/callback"
)

type Error error

type VectorStorage struct {
	vectorStore vectorstores.VectorStore
	generator   llms.Model
	embedder    embeddings.Embedder
	cfg         *Config
}

func NewVectorStorage(ctx context.Context, cfg *Config, embedder embeddings.Embedder, generator llms.Model) (*VectorStorage, error) {
	const op = "NewStorage"

	store, err := pgvector.New(
		ctx,
		pgvector.WithCollectionTableName("collections"),
		pgvector.WithEmbeddingTableName("embeddings"),
		pgvector.WithPreDeleteCollection(true),
		pgvector.WithVectorDimensions(cfg.EmbeddingDimensions),
		pgvector.WithEmbedder(embedder),
		pgvector.WithConnectionURL(cfg.PostgresURL),
	)

	if err != nil {
		slog.ErrorContext(ctx, "Error creating vector store",
			"op", op,
			"error", err)
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	slog.DebugContext(ctx, "Vector storage initialized",
		"collection_table", "collections",
		"embedding_table", "embeddings")
	return &VectorStorage{
		vectorStore: &store,
		embedder:    embedder,
		generator:   generator,
		cfg:         cfg,
	}, nil
}

func (s *VectorStorage) PutResource(ctx context.Context, resource models.Resource) ([]string, error) {
	const op = "VectorStorage.PutResource"
	slog.DebugContext(ctx, "Processing resource",
		"resource_type", resource.Type,
		"content_length", len(resource.ExtractedContent))

	slog.DebugContext(ctx, "Handling resource",
		"content_length", len(resource.ExtractedContent))
	chunkIDs, err := s.PutText(ctx, resource.ExtractedContent)

	if err != nil {
		slog.ErrorContext(ctx, "Failed to process resource",
			"op", op,
			"type", resource.Type,
			"error", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	slog.InfoContext(ctx, "Successfully processed resource",
		"chunks_count", len(chunkIDs),
		"resource_type", resource.Type)
	return chunkIDs, nil
}

func (s *VectorStorage) PutText(ctx context.Context, text string) ([]string, error) {
	const op = "VectorStorage.PutText"
	slog.DebugContext(ctx, "Processing text content",
		"content_length", len(text))

	docs, err := documentloaders.NewText(strings.NewReader(text)).
		LoadAndSplit(
			ctx,
			textsplitter.NewMarkdownTextSplitter(),
		)

	if err != nil {
		slog.ErrorContext(ctx, "Failed to process text",
			"op", op,
			"error", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	slog.DebugContext(ctx, "Adding text documents to vector store",
		"documents_count", len(docs))
	return s.addDocuments(ctx, docs)
}

func (s *VectorStorage) addDocuments(ctx context.Context, docs []schema.Document) ([]string, error) {
	const op = "VectorStorage.addDocuments"
	slog.DebugContext(ctx, "Adding documents to vector store",
		"documents_count", len(docs))

	ids, err := s.vectorStore.AddDocuments(ctx, docs)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to add documents",
			"op", op,
			"error", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	slog.InfoContext(ctx, "Documents added successfully",
		"documents_count", len(ids))
	return ids, nil
}

func (s *VectorStorage) SemanticSearch(ctx context.Context, query string) ([]models.Reference, error) {
	const op = "VectorStorage.SemanticSearch"
	slog.DebugContext(ctx, "Performing semantic search",
		"query", query)

	docs, err := s.vectorStore.SimilaritySearch(ctx, query, s.cfg.NumOfResults)
	if err != nil {
		slog.ErrorContext(ctx, "Semantic search failed",
			"op", op,
			"query", query,
			"error", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	slog.DebugContext(ctx, "Semantic search completed",
		"results_count", len(docs))
	return parseReferences(docs), nil
}

func (s *VectorStorage) GetAnswer(ctx context.Context, question string, refsCh chan<- []models.Reference) (models.SearchResult, error) {
	const op = "storage.GetAnswer"
	slog.DebugContext(ctx, "Getting answer",
		"question", question)

	result, err := s.ask(ctx, question, refsCh)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get answer",
			"op", op,
			"question", question,
			"error", err)
		return models.SearchResult{}, fmt.Errorf("%s: %w", op, err)
	}

	slog.InfoContext(ctx, "Answer generated",
		"answer_length", len(result.Answer),
		"references_count", len(result.References))
	return result, nil
}

func (s *VectorStorage) GetAnswerStream(ctx context.Context, question string, refsCh chan<- []models.Reference, chunkCh chan<- []byte) (models.SearchResult, error) {
	const op = "VectorStorage.GetAnswerStream"
	slog.DebugContext(ctx, "Starting answer streaming", "question", question)

	result, err := s.ask(
		ctx,
		question,
		refsCh,
		chains.WithStreamingFunc(newChunkHandler(chunkCh)),
	)

	if err != nil {
		slog.ErrorContext(ctx, "Streaming failed",
			"op", op,
			"question", question,
			"error", err,
		)
		return models.SearchResult{}, fmt.Errorf("%s: %w", op, err)
	}

	slog.InfoContext(ctx, "Answer stream completed", "chunks_sent", len(result.Answer))
	return result, err
}

func newChunkHandler(chunkCh chan<- []byte) func(ctx context.Context, chunk []byte) error {
	return func(ctx context.Context, chunk []byte) error {
		slog.Info("Received chunk", "chunk", string(chunk), "length", len(chunk))
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			chunkCh <- chunk
			return nil
		}
	}
}

func (s *VectorStorage) ask(ctx context.Context, question string, refsCh chan<- []models.Reference, opts ...chains.ChainCallOption) (models.SearchResult, error) {
	const op = "VectorStorage.ask"
	slog.DebugContext(ctx, "Processing question", "question", question)

	internalRefsCh := make(chan []models.Reference, 1)

	cb := callback.NewCallbackHandler(
		callback.WithRetrieverEndFunc(newRetrieverEndHandler(refsCh, internalRefsCh)),
	)

	retriever := s.setupRetriever(cb)
	retrievalQAChain := s.setupRetrievalQA(retriever)
	opts = append(opts, chains.WithMaxTokens(s.cfg.MaxTokens), chains.WithCallback(cb))

	answerCh := make(chan string)
	errCh := make(chan error)

	go func() {
		defer func() {
			close(answerCh)
			close(errCh)
		}()
		select {
		case <-ctx.Done():
			errCh <- ctx.Err()
		default:
			slog.DebugContext(ctx, "Running retrieval QA chain")
			answer, err := chains.Run(
				ctx,
				retrievalQAChain,
				question,
				opts...,
			)
			if err != nil {
				errCh <- err
			}

			answerCh <- answer
		}
	}()

	select {
	case <-ctx.Done():
		return models.SearchResult{}, ctx.Err()
	case err := <-errCh:
		return models.SearchResult{}, fmt.Errorf("%s: %w", op, err)
	case answer := <-answerCh:
		return models.SearchResult{
			Answer:     answer,
			References: <-internalRefsCh,
		}, nil
	}
}

func newRetrieverEndHandler(refsChans ...chan<- []models.Reference) func(ctx context.Context, query string, documents []schema.Document) {
	return func(ctx context.Context, query string, documents []schema.Document) {
		slog.Info("On retrieving was received documents", "documents_count", len(documents))
		select {
		case <-ctx.Done():
			return
		default:
			refs := parseReferences(documents)
			for _, ch := range refsChans {
				ch <- refs
			}
		}
	}
}

func (s *VectorStorage) setupRetriever(callbackHandler ...*callback.Handler) *vectorstores.Retriever {
	slog.DebugContext(context.Background(), "Configuring retriever",
		"num_results", s.cfg.NumOfResults)

	retriever := vectorstores.ToRetriever(s.vectorStore, s.cfg.NumOfResults)
	if len(callbackHandler) > 0 {
		retriever.CallbacksHandler = callbackHandler[0]
	}
	return &retriever
}

func (s *VectorStorage) setupRetrievalQA(retriever *vectorstores.Retriever) chains.RetrievalQA {
	slog.DebugContext(context.Background(), "Initializing QA chain")
	return chains.NewRetrievalQAFromLLM(
		s.generator,
		retriever,
	)
}

func parseReferences(docs []schema.Document) []models.Reference {
	slog.DebugContext(context.Background(), "Parsing references",
		"documents_count", len(docs))
	return lo.Map(docs, func(doc schema.Document, _ int) models.Reference {
		return models.Reference{
			Content: doc.PageContent,
			Score:   doc.Score,
		}
	})
}
