package configurator

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/viper"

	"github.com/nzb3/diploma/search-service/internal/validator"
)

// Configurator provides configuration management utilities
type Configurator struct {
	viper *viper.Viper
}

// NewConfigurator creates a new configurator instance
func NewConfigurator() *Configurator {
	v := viper.New()
	return &Configurator{viper: v}
}

// LoadConfig initializes global configuration from config file and environment variables
func LoadConfig(configPath, configName, configType string) error {
	// Setup environment variable mappings
	SetupEnvironmentMapping()

	viper.AddConfigPath(configPath)
	viper.SetConfigName(configName)
	viper.SetConfigType(configType)

	// Enable automatic environment variable reading
	viper.AutomaticEnv()

	// Try to read from config file
	if err := viper.ReadInConfig(); err != nil {
		slog.Warn("Failed to read config file, using defaults and environment variables",
			"error", err,
			"config_path", configPath,
			"config_name", configName)
		return nil
	}

	slog.Info("Successfully loaded config file",
		"config_file", viper.ConfigFileUsed())
	return nil
}

// ParseConfig parses configuration from the given config path into the provided struct
func ParseConfig[T any](configPath string) (*T, error) {
	config := new(T)

	// Parse from the specified config section
	if configPath != "" {
		var sub *viper.Viper

		mode := os.Getenv("GIN_MODE")
		if mode == "release" {
			sub = viper.Sub("production")
		} else {
			sub = viper.Sub("debug")
		}

		// If we couldn't get the mode-specific config, try the root level
		if sub == nil {
			sub = viper.Sub(configPath)
		} else {
			// If we got the mode-specific config, get the subsection from it
			sub = sub.Sub(configPath)
		}

		if sub == nil {
			return nil, fmt.Errorf("config section '%s' not found", configPath)
		}
		if err := sub.Unmarshal(config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config section '%s': %w", configPath, err)
		}
	} else {
		// Parse from root
		if err := viper.Unmarshal(config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}

	// Validate configuration
	if err := validator.Validate(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	slog.Debug("Configuration parsed and validated successfully", "section", configPath)
	return config, nil
}

// GetString gets a string value from config with environment variable fallback
func GetString(key string) string {
	return viper.GetString(key)
}

// GetInt gets an int value from config with environment variable fallback
func GetInt(key string) int {
	return viper.GetInt(key)
}

// GetBool gets a bool value from config with environment variable fallback
func GetBool(key string) bool {
	return viper.GetBool(key)
}

// GetStringSlice gets a string slice from config with environment variable fallback
func GetStringSlice(key string) []string {
	return viper.GetStringSlice(key)
}

// SetupEnvironmentMapping configures viper to map environment variables to config keys
func SetupEnvironmentMapping() {
	// No prefix for environment variables
	viper.SetEnvPrefix("")
	viper.AutomaticEnv()

	// Enable reading from environment with underscore separation
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Server configuration
	viper.BindEnv("server.host", "SERVER_HOST")
	viper.BindEnv("server.port", "SERVER_PORT")

	// Database configuration
	viper.BindEnv("postgres.host", "SEARCH_DB_HOST")
	// Use container port for internal Docker network connections
	containerPort := viper.GetString("SEARCH_DB_CONTAINER_PORT")
	if containerPort == "" {
		containerPort = "5432" // default PostgreSQL port
	}
	viper.Set("postgres.port", containerPort)
	// Note: We're not binding SEARCH_DB_PORT directly to postgres.port
	// because within Docker network, we need to use the container port (5432)
	viper.BindEnv("postgres.user", "SEARCH_DB_USER")
	viper.BindEnv("postgres.password", "SEARCH_DB_PASSWORD")
	viper.BindEnv("postgres.dbname", "SEARCH_DB_NAME")
	viper.BindEnv("postgres.sslmode", "SEARCH_DB_SSL_MODE")

	// Auth configuration
	viper.BindEnv("auth.host", "AUTH_HOST")
	viper.BindEnv("auth.port", "AUTH_PORT")
	viper.BindEnv("auth.realm", "AUTH_REALM")
	viper.BindEnv("auth.client_id", "AUTH_SEARCH_SERVICE_CLIENT_ID")
	viper.BindEnv("auth.client_secret", "AUTH_SEARCH_SERVICE_CLIENT_SECRET")

	// Kafka configuration
	viper.BindEnv("kafka.brokers", "KAFKA_BROKERS")
	viper.BindEnv("kafka.consumer_group_id", "KAFKA_CONSUMER_GROUP_ID")
	viper.BindEnv("kafka.topics.resource", "KAFKA_TOPIC_RESOURCE")

	// Vector storage configuration (from config file only)
	// No environment bindings for these as they should be in config.yml

	// Logger configuration
	viper.BindEnv("logger.level", "LOG_LEVEL")

	// Handle Kafka brokers specially (comma-separated list)
	if brokersEnv := viper.GetString("KAFKA_BROKERS"); brokersEnv != "" {
		brokers := strings.Split(brokersEnv, ",")
		for i, broker := range brokers {
			brokers[i] = strings.TrimSpace(broker)
		}
		viper.Set("kafka.brokers", brokers)
	}

	slog.Debug("Environment variable mappings configured")
}
