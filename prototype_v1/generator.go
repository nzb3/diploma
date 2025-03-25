package prototype_v1

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

type generator struct {
	llm *ollama.LLM
}

func NewGenerator() (*generator, error) {
	llm, err := ollama.New(
		ollama.WithServerURL("http://ollama-generator:11434/"),
		ollama.WithModel("llama3"),
	)
	if err != nil {
		return nil, err
	}

	return &generator{
		llm: llm,
	}, nil
}

func (g *generator) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	const op = "generator.GenerateContent"

	response, err := g.llm.GenerateContent(ctx, messages, options...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return response, nil
}

func (g *generator) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	const op = "generator.Call"
	response, err := g.llm.Call(ctx, prompt, options...)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return response, nil
}
