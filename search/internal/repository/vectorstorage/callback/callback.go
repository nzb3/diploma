package callback

import (
	"context"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

type (
	HandleTextFunc                    func(ctx context.Context, text string)
	HandleLLMStartFunc                func(ctx context.Context, prompts []string)
	HandleLLMGenerateContentStartFunc func(ctx context.Context, ms []llms.MessageContent)
	HandleLLMGenerateContentEndFunc   func(ctx context.Context, res *llms.ContentResponse)
	HandleLLMErrorFunc                func(ctx context.Context, err error)
	HandleChainStartFunc              func(ctx context.Context, inputs map[string]any)
	HandleChainEndFunc                func(ctx context.Context, outputs map[string]any)
	HandleChainErrorFunc              func(ctx context.Context, err error)
	HandleToolStartFunc               func(ctx context.Context, input string)
	HandleToolEndFunc                 func(ctx context.Context, output string)
	HandleToolErrorFunc               func(ctx context.Context, err error)
	HandleAgentActionFunc             func(ctx context.Context, action schema.AgentAction)
	HandleAgentFinishFunc             func(ctx context.Context, finish schema.AgentFinish)
	HandleRetrieverStartFunc          func(ctx context.Context, query string)
	HandleRetrieverEndFunc            func(ctx context.Context, query string, documents []schema.Document)
	HandleStreamingFuncFunc           func(ctx context.Context, chunk []byte)
)

type Handler struct {
	handleTextFunc                    HandleTextFunc
	handleLLMStartFunc                HandleLLMStartFunc
	handleLLMGenerateContentStartFunc HandleLLMGenerateContentStartFunc
	handleLLMGenerateContentEndFunc   HandleLLMGenerateContentEndFunc
	handleLLMErrorFunc                HandleLLMErrorFunc
	handleChainStartFunc              HandleChainStartFunc
	handleChainEndFunc                HandleChainEndFunc
	handleChainErrorFunc              HandleChainErrorFunc
	handleToolStartFunc               HandleToolStartFunc
	handleToolEndFunc                 HandleToolEndFunc
	handleToolErrorFunc               HandleToolErrorFunc
	handleAgentActionFunc             HandleAgentActionFunc
	handleAgentFinishFunc             HandleAgentFinishFunc
	handleRetrieverStartFunc          HandleRetrieverStartFunc
	handleRetrieverEndFunc            HandleRetrieverEndFunc
	handleStreamingFuncFunc           HandleStreamingFuncFunc
}

func NewCallbackHandler(opts ...Option) *Handler {
	c := &Handler{}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Handler) ApplyOption(opt Option) {
	opt(c)
}

func (c *Handler) HandleText(ctx context.Context, text string) {
	if c.handleTextFunc != nil {
		c.handleTextFunc(ctx, text)
	}
}

func (c *Handler) HandleLLMStart(ctx context.Context, prompts []string) {
	if c.handleLLMStartFunc != nil {
		c.handleLLMStartFunc(ctx, prompts)
	}
}

func (c *Handler) HandleLLMGenerateContentStart(ctx context.Context, ms []llms.MessageContent) {
	if c.handleLLMGenerateContentStartFunc != nil {
		c.handleLLMGenerateContentStartFunc(ctx, ms)
	}
}

func (c *Handler) HandleLLMGenerateContentEnd(ctx context.Context, res *llms.ContentResponse) {
	if c.handleLLMGenerateContentEndFunc != nil {
		c.handleLLMGenerateContentEndFunc(ctx, res)
	}
}

func (c *Handler) HandleLLMError(ctx context.Context, err error) {
	if c.handleLLMErrorFunc != nil {
		c.handleLLMErrorFunc(ctx, err)
	}
}

func (c *Handler) HandleChainStart(ctx context.Context, inputs map[string]any) {
	if c.handleChainStartFunc != nil {
		c.handleChainStartFunc(ctx, inputs)
	}
}

func (c *Handler) HandleChainEnd(ctx context.Context, outputs map[string]any) {
	if c.handleChainEndFunc != nil {
		c.handleChainEndFunc(ctx, outputs)
	}
}

func (c *Handler) HandleChainError(ctx context.Context, err error) {
	if c.handleChainErrorFunc != nil {
		c.handleChainErrorFunc(ctx, err)
	}
}

func (c *Handler) HandleToolStart(ctx context.Context, input string) {
	if c.handleToolStartFunc != nil {
		c.handleToolStartFunc(ctx, input)
	}
}

func (c *Handler) HandleToolEnd(ctx context.Context, output string) {
	if c.handleToolEndFunc != nil {
		c.handleToolEndFunc(ctx, output)
	}
}

func (c *Handler) HandleToolError(ctx context.Context, err error) {
	if c.handleToolErrorFunc != nil {
		c.handleToolErrorFunc(ctx, err)
	}
}

func (c *Handler) HandleAgentAction(ctx context.Context, action schema.AgentAction) {
	if c.handleAgentActionFunc != nil {
		c.handleAgentActionFunc(ctx, action)
	}
}

func (c *Handler) HandleAgentFinish(ctx context.Context, finish schema.AgentFinish) {
	if c.handleAgentFinishFunc != nil {
		c.handleAgentFinishFunc(ctx, finish)
	}
}

func (c *Handler) HandleRetrieverStart(ctx context.Context, query string) {
	if c.handleRetrieverStartFunc != nil {
		c.handleRetrieverStartFunc(ctx, query)
	}
}

func (c *Handler) HandleRetrieverEnd(ctx context.Context, query string, documents []schema.Document) {
	if c.handleRetrieverEndFunc != nil {
		c.handleRetrieverEndFunc(ctx, query, documents)
	}
}

func (c *Handler) HandleStreamingFunc(ctx context.Context, chunk []byte) {
	if c.handleStreamingFuncFunc != nil {
		c.handleStreamingFuncFunc(ctx, chunk)
	}
}
