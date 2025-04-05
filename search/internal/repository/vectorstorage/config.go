package vectorstorage

type Config struct {
	NumOfResults        int `yaml:"num_of_results"`
	MaxTokens           int `yaml:"max_tokens"`
	EmbeddingDimensions int `yaml:"embedding_dimensions"`
}

func NewConfig() (*Config, error) {
	// TODO: load from config file
	return &Config{
		NumOfResults:        10,
		MaxTokens:           2048,
		EmbeddingDimensions: 384,
	}, nil
}
