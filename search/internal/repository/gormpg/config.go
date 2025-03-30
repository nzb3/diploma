package gormpg

type Config struct {
	DSN      string
	DBName   string
	User     string
	Password string
	Host     string
	Port     string
}

func NewConfig() (*Config, error) {
	// TODO: read from env
	return &Config{
		DSN:      "postgres://postgres:postgres@postgres:5432/postgres?sslmode=disable",
		DBName:   "postgres",
		User:     "postgres",
		Password: "postgres",
		Host:     "postgres",
		Port:     "5432",
	}, nil
}
