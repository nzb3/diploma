package vectorstorage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
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

var (
	UnsupportedResourceTypeError Error = errors.New("unsupported resource type")
)

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
		pgvector.WithEmbedder(embedder),
		pgvector.WithConnectionURL("postgres://postgres:postgres@postgres:5432/postgres?sslmode=disable"),
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
		"source", resource.Source,
		"content_length", len(resource.ExtractedContent))

	var chunkIDs []string
	var err error

	switch resource.Type {
	case models.ResourceTypeURL:
		slog.DebugContext(ctx, "Handling URL resource",
			"url", resource.Source)
		chunkIDs, err = s.PutURL(ctx, resource.Source)
	case models.ResourceTypePDF:
		slog.DebugContext(ctx, "Handling PDF resource",
			"content_size", len(resource.RawContent))
		chunkIDs, err = s.PutPDFFile(ctx, resource.RawContent)
	case models.ResourceTypeText:
		slog.DebugContext(ctx, "Handling text resource",
			"content_length", len(resource.ExtractedContent))
		chunkIDs, err = s.PutText(ctx, resource.ExtractedContent)
	default:
		slog.ErrorContext(ctx, "Unsupported resource type",
			"op", op,
			"type", resource.Type)
		return nil, fmt.Errorf("%s:%w %s", op, UnsupportedResourceTypeError, resource.Type)
	}

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

func (s *VectorStorage) PutURL(ctx context.Context, source string) ([]string, error) {
	const op = "VectorStorage.PutURL"
	slog.DebugContext(ctx, "Fetching URL content",
		"url", source)

	resp, err := http.Get(source)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to fetch URL",
			"op", op,
			"url", source,
			"error", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer resp.Body.Close()

	slog.DebugContext(ctx, "Loading HTML documents",
		"url", source)
	docs, err := documentloaders.NewHTML(resp.Body).
		LoadAndSplit(
			ctx,
			textsplitter.NewTokenSplitter(),
		)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to process HTML",
			"op", op,
			"error", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	slog.DebugContext(ctx, "Adding HTML documents to vector store",
		"documents_count", len(docs))
	return s.addDocuments(ctx, docs)
}

func (s *VectorStorage) PutPDFFile(ctx context.Context, document []byte) ([]string, error) {
	const op = "VectorStorage.PutPDFFile"
	slog.DebugContext(ctx, "Processing PDF document",
		"file_size", len(document))

	docs, err := documentloaders.NewPDF(
		bytes.NewReader(document),
		int64(len(document)),
	).LoadAndSplit(
		ctx,
		textsplitter.NewTokenSplitter(),
	)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to process PDF",
			"op", op,
			"error", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	slog.DebugContext(ctx, "Adding PDF documents to vector store",
		"documents_count", len(docs))
	return s.addDocuments(ctx, docs)
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

func (s *VectorStorage) GetAnswer(ctx context.Context, question string) (models.SearchResult, error) {
	const op = "storage.GetAnswer"
	slog.DebugContext(ctx, "Getting answer",
		"question", question)

	result, err := s.ask(ctx, question)
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

func (s *VectorStorage) GetAnswerStream(ctx context.Context, question string, chunks chan<- []byte) (models.SearchResult, error) {
	const op = "VectorStorage.GetAnswerStream"
	slog.DebugContext(ctx, "Starting answer streaming",
		"question", question)

	result, err := s.ask(ctx, question, chains.WithStreamingFunc(s.streamAnswer(chunks)))
	if err != nil {
		slog.ErrorContext(ctx, "Streaming failed",
			"op", op,
			"question", question,
			"error", err)
		return models.SearchResult{}, fmt.Errorf("%s: %w", op, err)
	}

	slog.InfoContext(ctx, "Answer stream completed",
		"chunks_sent", len(result.Answer))
	return result, err
}

func (s *VectorStorage) ask(ctx context.Context, question string, opts ...chains.ChainCallOption) (models.SearchResult, error) {
	const op = "VectorStorage.ask"
	slog.DebugContext(ctx, "Processing question",
		"question", question)

	docsChan := make(chan []schema.Document)
	cb := callback.NewCallbackHandler(
		callback.WithRetrieverEndFunc(
			func(ctx context.Context, query string, documents []schema.Document) {
				slog.DebugContext(ctx, "Retriever completed",
					"documents_found", len(documents))
				go func() {
					defer close(docsChan)
					docsChan <- documents
				}()
			},
		),
	)

	retriever := s.setupRetriever(cb)
	retrievalQAChain := s.setupRetrievalQA(retriever)
	opts = append(opts, chains.WithMaxTokens(s.cfg.MaxTokens), chains.WithCallback(cb))

	slog.DebugContext(ctx, "Running retrieval QA chain")
	result, err := chains.Run(
		ctx,
		retrievalQAChain,
		question,
		opts...,
	)

	if err != nil {
		slog.ErrorContext(ctx, "Chain execution failed",
			"op", op,
			"error", err)
		return models.SearchResult{}, fmt.Errorf("%s: %w", op, err)
	}

	docs := <-docsChan
	slog.DebugContext(ctx, "Retrieved documents",
		"count", len(docs))

	return models.SearchResult{
		Answer:     result,
		References: parseReferences(docs),
	}, nil
}

func (s *VectorStorage) setupRetriever(callbackHandler ...*callback.Handler) vectorstores.Retriever {
	const op = "VectorStorage.setupRetriever"
	slog.DebugContext(context.Background(), "Configuring retriever",
		"num_results", s.cfg.NumOfResults)

	retriever := vectorstores.ToRetriever(s.vectorStore, s.cfg.NumOfResults)
	if len(callbackHandler) > 0 {
		retriever.CallbacksHandler = callbackHandler[0]
	}
	return retriever
}

func (s *VectorStorage) setupRetrievalQA(retriever vectorstores.Retriever) chains.RetrievalQA {
	const op = "VectorStorage.setupRetrievalQA"
	slog.DebugContext(context.Background(), "Initializing QA chain")
	return chains.NewRetrievalQAFromLLM(
		s.generator,
		retriever,
	)
}

func (s *VectorStorage) streamAnswer(chunkChan chan<- []byte) func(ctx context.Context, chunk []byte) error {
	const op = "VectorStorage.streamAnswer"
	return func(ctx context.Context, chunk []byte) error {
		select {
		case <-ctx.Done():
			slog.WarnContext(ctx, "Stream context cancelled",
				"op", op,
				"error", ctx.Err())
			return ctx.Err()
		default:
			slog.DebugContext(ctx, "Sending stream chunk",
				"chunk_length", len(chunk))
			chunkChan <- chunk
			return nil
		}
	}
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
