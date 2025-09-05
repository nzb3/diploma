package middleware

import (
	"fmt"

	"github.com/nzb3/diploma/search-service/internal/configurator"
)

// AuthConfig holds necessary configuration for Keycloak authentication
type AuthConfig struct {
	Host         string `yaml:"host" mapstructure:"host" validate:"required"`
	Port         string `yaml:"port" mapstructure:"port" validate:"required"`
	Realm        string `yaml:"realm" mapstructure:"realm" validate:"required"`
	ClientID     string `yaml:"client_id" mapstructure:"client_id" validate:"required"`
	ClientSecret string `yaml:"client_secret" mapstructure:"client_secret" validate:"required"`
}

// GetKeycloakURL constructs the full Keycloak URL
func (c *AuthConfig) GetKeycloakURL() string {
	return fmt.Sprintf("http://%s:%s", c.Host, c.Port)
}

// NewAuthConfig loads authentication configuration from config file and environment variables
func NewAuthConfig() (*AuthConfig, error) {
	// Set defaults
	setDefaults()

	config := new(AuthConfig)
	config.Realm = configurator.GetString("auth.realm")
	config.ClientID = configurator.GetString("auth.client_id")
	config.ClientSecret = configurator.GetString("auth.client_secret")
	config.Host = configurator.GetString("auth.host")
	config.Port = configurator.GetString("auth.port")

	return config, nil
}

// setDefaults sets default values for authentication configuration
func setDefaults() {
	// Note: This is a placeholder. In practice, we'd use viper.SetDefault
	// The actual defaults are set in the configurator package
}
