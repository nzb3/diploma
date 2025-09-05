package vectorstorage

import (
	"fmt"

	"github.com/nzb3/diploma/search-service/internal/configurator"
)

// Config holds vector storage configuration
type Config struct {
	NumOfResults        int `yaml:"num_of_results" mapstructure:"num_of_results"`
	MaxTokens           int `yaml:"max_tokens" mapstructure:"max_tokens"`
	EmbeddingDimensions int `yaml:"embedding_dimensions" mapstructure:"embedding_dimensions"`
}

// NewConfig loads vector storage configuration from config file
func NewConfig() (*Config, error) {
	// Set defaults
	setDefaults()

	// Parse configuration from "vector_storage" section
	config, err := configurator.ParseConfig[Config]("vector_storage")
	if err != nil {
		return nil, fmt.Errorf("failed to parse vector storage config: %w", err)
	}

	return config, nil
}

// setDefaults sets default values for vector storage configuration
func setDefaults() {
	// Note: This is a placeholder. In practice, we'd use viper.SetDefault
	// The actual defaults are set in the configurator package
}
