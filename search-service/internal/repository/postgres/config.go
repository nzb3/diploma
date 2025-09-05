package postgres

import (
	"fmt"

	"github.com/nzb3/diploma/search-service/internal/configurator"
)

// Config holds PostgreSQL database configuration
type Config struct {
	Host     string `yaml:"host" mapstructure:"host" validate:"required"`
	Port     string `yaml:"port" mapstructure:"port" validate:"required"`
	User     string `yaml:"user" mapstructure:"user" validate:"required"`
	Password string `yaml:"password" mapstructure:"password" validate:"required"`
	DBName   string `yaml:"dbname" mapstructure:"dbname" validate:"required"`
	SSLMode  string `yaml:"sslmode" mapstructure:"sslmode"`
}

// GetConnectionString builds PostgreSQL connection string
func (c *Config) GetConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode)
}

// NewConfig loads PostgreSQL configuration using the configurator package
func NewConfig() (*Config, error) {
	config := new(Config)

	config.Host = configurator.GetString("postgres.host")
	config.Port = configurator.GetString("postgres.port")
	config.User = configurator.GetString("postgres.user")
	config.Password = configurator.GetString("postgres.password")
	config.DBName = configurator.GetString("postgres.dbname")
	config.SSLMode = configurator.GetString("postgres.sslmode")

	return config, nil
}
