package vectorstorage

type Config struct {
	NumOfResults        int    `yaml:"num_of_results"`
	MaxTokens           int    `yaml:"max_tokens"`
	EmbeddingDimensions int    `yaml:"embedding_dimensions"`
	PostgresURL         string `yaml:"postgres.url"`
}

func NewConfig(postgresURL string) (*Config, error) {
	// TODO: load from config file
	return &Config{
		PostgresURL:         postgresURL,
		NumOfResults:        10,
		MaxTokens:           2048,
		EmbeddingDimensions: 1024,
	}, nil
}
