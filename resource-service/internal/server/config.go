package server

import (
	"time"

	"github.com/nzb3/diploma/resource-service/internal/configurator"
)

// Config holds HTTP server configuration
type Config struct {
	Host              string        `yaml:"host" mapstructure:"host" validate:"required"`
	Port              string        `yaml:"port" mapstructure:"port" validate:"required"`
	ReadTimeout       time.Duration `yaml:"read_timeout" mapstructure:"read_timeout"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout" mapstructure:"read_header_timeout"`
	WriteTimeout      time.Duration `yaml:"write_timeout" mapstructure:"write_timeout"`
	IdleTimeout       time.Duration `yaml:"idle_timeout" mapstructure:"idle_timeout"`
	MaxHeaderBytes    int           `yaml:"max_header_bytes" mapstructure:"max_header_bytes"`
	ShutdownTimeout   time.Duration `yaml:"shutdown_timeout" mapstructure:"shutdown_timeout"`
}

// NewConfig loads server configuration from config file and environment variables
func NewConfig() (*Config, error) {
	// Set defaults
	setDefaults()

	// Parse configuration from "server" section
	config, err := configurator.ParseConfig[Config]("server")
	if err != nil {
		return nil, err
	}

	return config, nil
}

// setDefaults sets default values for server configuration
func setDefaults() {
	// Note: This is a placeholder. In practice, we'd use viper.SetDefault
	// The actual defaults are set in the configurator package
}
