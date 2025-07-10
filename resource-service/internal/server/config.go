package server

import (
	"time"
)

type Config struct {
	HTTP struct {
		Port              string        `yaml:"port"`
		ReadTimeout       time.Duration `yaml:"read_timeout"`
		ReadHeaderTimeout time.Duration `yaml:"read_header_timeout"`
		WriteTimeout      time.Duration `yaml:"write_timeout"`
		IdleTimeout       time.Duration `yaml:"idle_timeout"`
		MaxHeaderBytes    int           `yaml:"max_header_bytes"`
		ShutdownTimeout   time.Duration `yaml:"shutdown_timeout"`
	} `yaml:"http"`
}

func NewConfig() (*Config, error) {
	// TODO: load config from file
	return &Config{
		HTTP: HTTPConfig{
			Port:              "8081",
			ReadTimeout:       5 * time.Second,
			ReadHeaderTimeout: 2 * time.Second,
			WriteTimeout:      500 * time.Second,
			IdleTimeout:       300 * time.Second,
			MaxHeaderBytes:    1 << 20,
		},
	}, nil
}

type HTTPConfig struct {
	Port              string        `yaml:"port"`
	ReadTimeout       time.Duration `yaml:"read_timeout"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout"`
	WriteTimeout      time.Duration `yaml:"write_timeout"`
	IdleTimeout       time.Duration `yaml:"idle_timeout"`
	MaxHeaderBytes    int           `yaml:"max_header_bytes"`
	ShutdownTimeout   time.Duration `yaml:"shutdown_timeout"`
}
