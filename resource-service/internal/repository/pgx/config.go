package pgx

import (
	"time"

	"github.com/nzb3/diploma/resource-service/internal/configurator"
)

// Config holds the configuration for PostgreSQL connection pool
type Config struct {
	Host            string        `yaml:"host" mapstructure:"host" validate:"required"`
	Port            string        `yaml:"port" mapstructure:"port" validate:"required"`
	Database        string        `yaml:"dbname" mapstructure:"dbname" validate:"required"`
	Username        string        `yaml:"user" mapstructure:"user" validate:"required"`
	Password        string        `yaml:"password" mapstructure:"password" validate:"required"`
	SSLMode         string        `yaml:"sslmode" mapstructure:"sslmode"`
	MaxOpenConns    int           `yaml:"max_open_conns" mapstructure:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns" mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time" mapstructure:"conn_max_idle_time"`
}

// NewConfig creates a new configuration with values from environment variables
func NewConfig() (*Config, error) {
	config := &Config{
		Host:     configurator.GetString("database.host"),
		Port:     configurator.GetString("database.port"),
		Database: configurator.GetString("database.dbname"),
		Username: configurator.GetString("database.user"),
		Password: configurator.GetString("database.password"),
		SSLMode:  configurator.GetString("database.sslmode"),
	}

	// Set defaults for empty values
	if config.Host == "" {
		config.Host = "localhost"
	}
	if config.Port == "" {
		config.Port = "5432"
	}
	if config.Database == "" {
		config.Database = "postgres"
	}
	if config.Username == "" {
		config.Username = "postgres"
	}
	if config.Password == "" {
		config.Password = "postgres"
	}
	if config.SSLMode == "" {
		config.SSLMode = "disable"
	}

	// Parse integer values with defaults
	maxOpenConns := configurator.GetInt("database.max_open_conns")
	if maxOpenConns > 0 {
		config.MaxOpenConns = maxOpenConns
	} else {
		config.MaxOpenConns = 25
	}

	maxIdleConns := configurator.GetInt("database.max_idle_conns")
	if maxIdleConns > 0 {
		config.MaxIdleConns = maxIdleConns
	} else {
		config.MaxIdleConns = 5
	}

	// Parse duration values with defaults
	if connMaxLifetimeStr := configurator.GetString("database.conn_max_lifetime"); connMaxLifetimeStr != "" {
		if duration, err := time.ParseDuration(connMaxLifetimeStr); err == nil {
			config.ConnMaxLifetime = duration
		} else {
			config.ConnMaxLifetime = 30 * time.Minute
		}
	} else {
		config.ConnMaxLifetime = 30 * time.Minute
	}

	if connMaxIdleTimeStr := configurator.GetString("database.conn_max_idle_time"); connMaxIdleTimeStr != "" {
		if duration, err := time.ParseDuration(connMaxIdleTimeStr); err == nil {
			config.ConnMaxIdleTime = duration
		} else {
			config.ConnMaxIdleTime = 5 * time.Minute
		}
	} else {
		config.ConnMaxIdleTime = 5 * time.Minute
	}

	return config, nil
}

// GetDSN returns the data source name for the database connection
func (c *Config) GetDSN() string {
	return "postgres://" + c.Username + ":" + c.Password + "@" + c.Host + ":" + c.Port + "/" + c.Database + "?sslmode=" + c.SSLMode
}
