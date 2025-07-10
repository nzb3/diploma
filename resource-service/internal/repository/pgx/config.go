package pgx

import (
	"time"
)

// Config holds the configuration for PostgreSQL connection pool
type Config struct {
	DSN             string        `yaml:"dsn" env:"DB_DSN"`
	Host            string        `yaml:"host" env:"DB_HOST"`
	Port            string        `yaml:"port" env:"DB_PORT"`
	Database        string        `yaml:"database" env:"DB_NAME"`
	Username        string        `yaml:"username" env:"DB_USER"`
	Password        string        `yaml:"password" env:"DB_PASSWORD"`
	SSLMode         string        `yaml:"ssl_mode" env:"DB_SSL_MODE"`
	MaxOpenConns    int           `yaml:"max_open_conns" env:"DB_MAX_OPEN_CONNS"`
	MaxIdleConns    int           `yaml:"max_idle_conns" env:"DB_MAX_IDLE_CONNS"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" env:"DB_CONN_MAX_LIFETIME"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time" env:"DB_CONN_MAX_IDLE_TIME"`
}

// NewConfig creates a new configuration with default values
func NewConfig() *Config {
	return &Config{
		DSN:             "postgres://postgres:postgres@resource_database:5432/postgres?sslmode=disable",
		Host:            "localhost",
		Port:            "5432",
		Database:        "postgres",
		Username:        "postgres",
		Password:        "postgres",
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}
}

// GetDSN returns the data source name for the database connection
func (c *Config) GetDSN() string {
	if c.DSN != "" {
		return c.DSN
	}

	return "postgres://" + c.Username + ":" + c.Password + "@" + c.Host + ":" + c.Port + "/" + c.Database + "?sslmode=" + c.SSLMode
}
