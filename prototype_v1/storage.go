package prototype_v1

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/pgvector"
)

const (
	numOfResults int = 10
	maxTokens    int = 2048
)

type storage struct {
	vectorStore vectorstores.VectorStore
	retrievalQA chains.Chain
}

func NewStorage(ctx context.Context, embedder embeddings.Embedder, generator llms.Model) (*storage, error) {
	const op = "NewStorage"

	store, err := pgvector.New(
		ctx,
		pgvector.WithEmbedder(embedder),
		pgvector.WithConnectionURL("postgres://postgres:postgres@postgres:5432/postgres?sslmode=disable"),
	)

	if err != nil {
		slog.Error("error creating vector store", "op", op, "error", err.Error())
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	retrievalQAChain := chains.NewRetrievalQAFromLLM(
		generator,
		vectorstores.ToRetriever(store, numOfResults),
	)

	return &storage{
		vectorStore: &store,
		retrievalQA: retrievalQAChain,
	}, nil
}

func (s *storage) PutSite(ctx context.Context, source string) error {
	const op = "storage.PutSite"

	resp, err := http.Get(source)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	defer resp.Body.Close()

	docs, err := documentloaders.NewHTML(resp.Body).
		LoadAndSplit(
			ctx,
			textsplitter.NewTokenSplitter(),
		)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return s.addDocuments(ctx, docs)
}

func (s *storage) PutPDFFile(ctx context.Context, document []byte, opts ...documentloaders.PDFOptions) error {
	const op = "storage.PutPDFFile"

	docs, err := documentloaders.NewPDF(
		bytes.NewReader(document),
		int64(len(document)),
		opts...,
	).LoadAndSplit(
		ctx,
		textsplitter.NewTokenSplitter(),
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return s.addDocuments(ctx, docs)
}

func (s *storage) PutText(ctx context.Context, text string) error {
	const op = "storage.PutText"

	docs, err := documentloaders.NewText(strings.NewReader(text)).
		LoadAndSplit(
			ctx,
			textsplitter.NewMarkdownTextSplitter(),
		)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return s.addDocuments(ctx, docs)
}

func (s *storage) addDocuments(ctx context.Context, docs []schema.Document) error {
	const op = "storage.addDocuments"
	_, err := s.vectorStore.AddDocuments(ctx, docs)
	if err != nil {
		slog.Error("error adding texts to vector store", op, slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *storage) GetAnswer(ctx context.Context, question string) (string, error) {
	const op = "storage.GetAnswer"

	result, err := chains.Run(
		ctx,
		s.retrievalQA,
		question,
		chains.WithMaxTokens(maxTokens),
	)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return result, nil
}

func (s *storage) SemanticSearch(ctx context.Context, query string, maxResults int) ([]schema.Document, error) {
	const op = "storage.SemanticSearch"

	searchResults, err := s.vectorStore.SimilaritySearch(ctx, query, maxResults)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return searchResults, nil
}
