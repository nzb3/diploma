package main

import (
	"context"
	"log/slog"

	"github.com/tmc/langchaingo/llms/ollama"
)

type embedder struct {
	llm *ollama.LLM
}

func NewEmbedder() (*embedder, error) {
	llm, err := ollama.New(
		ollama.WithServerURL("http://ollama-embedder:11434/"),
		ollama.WithModel("all-minilm"),
	)

	if err != nil {
		return nil, err
	}

	return &embedder{
		llm: llm,
	}, nil
}

func (e *embedder) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	const op = "embedder.EmbedDocuments"

	embeddedTexts, err := e.llm.CreateEmbedding(ctx, texts)
	if err != nil {
		slog.Error("failed to create embedding", op, slog.String("error", err.Error()))
		return nil, err
	}

	return embeddedTexts, nil
}

func (e *embedder) EmbedQuery(ctx context.Context, query string) ([]float32, error) {
	const op = "embedder.EmbedQuery"

	embeddedQuery, err := e.llm.CreateEmbedding(ctx, []string{query})
	if err != nil {
		slog.Error("failed to create embedding", op, slog.String("error", err.Error()))
		return nil, err
	}

	return embeddedQuery[0], nil
}
