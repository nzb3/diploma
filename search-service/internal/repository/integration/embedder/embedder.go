package embedder

import (
	"context"
	"log/slog"

	"github.com/tmc/langchaingo/llms/ollama"
)

type Embedder struct {
	llm *ollama.LLM
}

func NewEmbedder(llm *ollama.LLM) (*Embedder, error) {

	return &Embedder{
		llm: llm,
	}, nil
}

func (e *Embedder) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	const op = "Embedder.EmbedDocuments"

	embeddedTexts, err := e.llm.CreateEmbedding(ctx, texts)
	if err != nil {
		slog.Error("failed to create embedding", op, slog.String("error", err.Error()))
		return nil, err
	}

	return embeddedTexts, nil
}

func (e *Embedder) EmbedQuery(ctx context.Context, query string) ([]float32, error) {
	const op = "Embedder.EmbedQuery"

	embeddedQuery, err := e.llm.CreateEmbedding(ctx, []string{query})
	if err != nil {
		slog.Error("failed to create embedding", op, slog.String("error", err.Error()))
		return nil, err
	}

	return embeddedQuery[0], nil
}
