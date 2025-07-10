package vectorstorage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/samber/lo"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/pgvector"

	"github.com/nzb3/diploma/search-service/internal/controllers/middleware"
	"github.com/nzb3/diploma/search-service/internal/domain/models"
	"github.com/nzb3/diploma/search-service/internal/domain/services/searchservice"
	"github.com/nzb3/diploma/search-service/internal/repository/vectorstorage/callback"
)

const userIDFilter = "user_id"

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
		pgvector.WithPreDeleteCollection(false),
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
	slog.DebugContext(ctx, "Vector storage initialized")
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
	text = clearText(text)
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

	userID, err := getUserID(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting user ID",
			"op", op,
			"error", err,
		)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	for i := range docs {
		docs[i].Metadata = map[string]any{
			userIDFilter: userID,
		}
	}

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

func (s *VectorStorage) GetAnswer(ctx context.Context, question string) (string, []models.Reference, error) {
	const op = "storage.GetAnswer"

	slog.DebugContext(ctx, "Getting answer",
		"question", question)

	answerCh, refsCh, errCh, _ := s.ask(ctx, question)

	select {
	case <-ctx.Done():
		slog.DebugContext(ctx, "Context cancelled",
			"question", question,
		)
		return "", nil, ctx.Err()
	case err := <-errCh:
		slog.DebugContext(ctx, "Error getting answer",
			"question", question,
			"error", err,
		)
		return "", nil, ctx.Err()
	case answer := <-answerCh:
		slog.DebugContext(ctx, "Successfully got answer",
			"question", question,
			"answer", answer,
		)
		return answer, <-refsCh, nil
	case refs := <-refsCh:
		slog.DebugContext(ctx, "Successfully got references",
			"question", question,
			"refs", refs,
		)
		return <-answerCh, <-refsCh, nil
	}
}

func (s *VectorStorage) GetAnswerStream(ctx context.Context, question string, opts ...searchservice.SearchOption) (<-chan string, <-chan []models.Reference, <-chan []byte, <-chan error) {
	const op = "VectorStorage.GetAnswerStream"
	slog.DebugContext(ctx, "Starting answer streaming", "question", question)

	chunkCh := make(chan []byte, 1)

	options := &searchservice.SearchOptions{
		NumberOfReferences: s.cfg.NumOfResults,
	}
	for _, opt := range opts {
		opt(options)
	}

	slog.DebugContext(ctx, "Configured answer stream",
		"question", question,
		"num_references", options.NumberOfReferences)

	answerCh, refsCh, errCh, doneCh := s.ask(
		ctx,
		question,
		chains.WithStreamingFunc(newChunkHandler(chunkCh)),
		searchservice.WithNumberOfReferences(options.NumberOfReferences),
	)

	go func() {
		select {
		case <-ctx.Done():
			return
		case <-doneCh:
			close(chunkCh)
		}
	}()

	return answerCh, refsCh, chunkCh, errCh
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

func (s *VectorStorage) ask(ctx context.Context, question string, opts ...interface{}) (<-chan string, <-chan []models.Reference, <-chan error, <-chan struct{}) {
	const op = "VectorStorage.ask"
	slog.DebugContext(ctx, "Processing question", "question", question)

	var chainOpts []chains.ChainCallOption
	numOfResults := s.cfg.NumOfResults

	for _, opt := range opts {
		switch o := opt.(type) {
		case chains.ChainCallOption:
			chainOpts = append(chainOpts, o)
		case searchservice.SearchOption:
			sOpts := &searchservice.SearchOptions{NumberOfReferences: s.cfg.NumOfResults}
			o(sOpts)
			numOfResults = sOpts.NumberOfReferences
		}
	}

	refsCh := make(chan []models.Reference)
	answerCh := make(chan string)
	errCh := make(chan error)

	doneCh := make(chan struct{})
	go func() {
		defer func() {
			close(refsCh)
			close(answerCh)
			close(errCh)
			close(doneCh)
		}()

		cb := callback.NewCallbackHandler(
			callback.WithRetrieverEndFunc(newRetrieverEndHandler(refsCh)),
		)

		userID, err := getUserID(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to get user ID", "op", op, "error", err)
			errCh <- fmt.Errorf("%s: %w", op, err)
			return
		}

		filters := map[string]interface{}{
			userIDFilter: userID,
		}

		retriever := s.setupRetriever(filters, numOfResults, cb)
		chain, err := s.setupChains(retriever)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to setup retriever", "op", op, "error", err)
			errCh <- fmt.Errorf("%s: %w", op, err)
		}

		chainOpts = append(chainOpts, chains.WithMaxTokens(s.cfg.MaxTokens), chains.WithCallback(cb))

		select {
		case <-ctx.Done():
			errCh <- ctx.Err()
		default:
			slog.DebugContext(ctx, "Running retrieval QA chain")
			answer, err := chains.Run(
				ctx,
				chain,
				question,
				chainOpts...,
			)
			if err != nil {
				errCh <- fmt.Errorf("%s:%w", op, err)
			}

			answerCh <- answer
		}
	}()

	return answerCh, refsCh, errCh, doneCh
}

func newRetrieverEndHandler(refsChains ...chan<- []models.Reference) func(ctx context.Context, query string, documents []schema.Document) {
	return func(ctx context.Context, query string, documents []schema.Document) {
		slog.Info("On retrieving was received documents", "documents_count", len(documents))
		select {
		case <-ctx.Done():
			return
		default:
			refs := parseReferences(documents)
			for _, ch := range refsChains {
				ch <- refs
			}
		}
	}
}

func getUserID(ctx context.Context) (string, error) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		return "", errors.New("user ID not found in context")
	}
	return userID, nil
}

func (s *VectorStorage) setupRetriever(filters map[string]interface{},
	numResults int,
	callbackHandler ...*callback.Handler,
) *vectorstores.Retriever {
	slog.DebugContext(context.Background(), "Configuring retriever",
		"num_results", numResults)
	retriever := vectorstores.ToRetriever(
		s.vectorStore,
		numResults,
		vectorstores.WithFilters(filters),
		vectorstores.WithScoreThreshold(0.5),
	)
	if len(callbackHandler) > 0 {
		retriever.CallbacksHandler = callbackHandler[0]
	}
	return &retriever
}

func (s *VectorStorage) setupChains(retriever *vectorstores.Retriever) (chains.Chain, error) {
	qaChain := s.setupRetrievalQA(retriever)

	return chains.NewSimpleSequentialChain(
		[]chains.Chain{qaChain},
	)
}

func (s *VectorStorage) setupRetrievalQA(retriever *vectorstores.Retriever) chains.RetrievalQA {
	customPromptText := `Use the following pieces of context to answer the question at the end. If you don't know the answer, just say that you don't know, don't try to make up an answer

{{.context}}

Question: {{.question}}

Helpful Answer:
`

	prompt := prompts.NewPromptTemplate(
		customPromptText,
		[]string{"context", "question"},
	)

	qaPromptSelector := chains.ConditionalPromptSelector{
		DefaultPrompt: prompt,
	}

	prompt = qaPromptSelector.GetPrompt(s.generator)

	llmChain := chains.NewLLMChain(s.generator, prompt)
	return chains.NewRetrievalQA(
		chains.NewStuffDocuments(llmChain),
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

func clearText(text string) string {
	re := regexp.MustCompile(`!\[[^\]]*\]\([^)]+\)`)
	return re.ReplaceAllString(text, "")
}
