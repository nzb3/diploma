package callback

type Option func(*Handler)

func WithTextFunc(fn HandleTextFunc) Option {
	return func(c *Handler) {
		c.handleTextFunc = fn
	}
}

func WithLLMStartFunc(fn HandleLLMStartFunc) Option {
	return func(c *Handler) {
		c.handleLLMStartFunc = fn
	}
}

func WithLLMGenerateContentStartFunc(fn HandleLLMGenerateContentStartFunc) Option {
	return func(c *Handler) {
		c.handleLLMGenerateContentStartFunc = fn
	}
}

func WithLLMGenerateContentEndFunc(fn HandleLLMGenerateContentEndFunc) Option {
	return func(c *Handler) {
		c.handleLLMGenerateContentEndFunc = fn
	}
}

func WithLLMErrorFunc(fn HandleLLMErrorFunc) Option {
	return func(c *Handler) {
		c.handleLLMErrorFunc = fn
	}
}

func WithChainStartFunc(fn HandleChainStartFunc) Option {
	return func(c *Handler) {
		c.handleChainStartFunc = fn
	}
}

func WithChainEndFunc(fn HandleChainEndFunc) Option {
	return func(c *Handler) {
		c.handleChainEndFunc = fn
	}
}

func WithChainErrorFunc(fn HandleChainErrorFunc) Option {
	return func(c *Handler) {
		c.handleChainErrorFunc = fn
	}
}

func WithToolStartFunc(fn HandleToolStartFunc) Option {
	return func(c *Handler) {
		c.handleToolStartFunc = fn
	}
}

func WithToolEndFunc(fn HandleToolEndFunc) Option {
	return func(c *Handler) {
		c.handleToolEndFunc = fn
	}
}

func WithToolErrorFunc(fn HandleToolErrorFunc) Option {
	return func(c *Handler) {
		c.handleToolErrorFunc = fn
	}
}

func WithAgentActionFunc(fn HandleAgentActionFunc) Option {
	return func(c *Handler) {
		c.handleAgentActionFunc = fn
	}
}

func WithAgentFinishFunc(fn HandleAgentFinishFunc) Option {
	return func(c *Handler) {
		c.handleAgentFinishFunc = fn
	}
}

func WithRetrieverStartFunc(fn HandleRetrieverStartFunc) Option {
	return func(c *Handler) {
		c.handleRetrieverStartFunc = fn
	}
}

func WithRetrieverEndFunc(fn HandleRetrieverEndFunc) Option {
	return func(c *Handler) {
		c.handleRetrieverEndFunc = fn
	}
}

func WithStreamingFuncFunc(fn HandleStreamingFuncFunc) Option {
	return func(c *Handler) {
		c.handleStreamingFuncFunc = fn
	}
}
