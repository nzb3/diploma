package generator

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

type Generator struct {
	llm *ollama.LLM
}

func NewGenerator(llm *ollama.LLM) (*Generator, error) {
	return &Generator{
		llm: llm,
	}, nil
}

func (g *Generator) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	const op = "Generator.GenerateContent"

	response, err := g.llm.GenerateContent(ctx, messages, options...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return response, nil
}

func (g *Generator) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	const op = "Generator.Call"
	response, err := g.llm.Call(ctx, prompt, options...)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return response, nil
}
