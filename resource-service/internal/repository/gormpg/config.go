package gormpg

type Config struct {
	DSN      string `yaml:"url"`
	DBName   string `yaml:"db_name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
}

func NewConfig() (*Config, error) {
	return &Config{
		DSN:      "postgres://postgres:postgres@search_database:5432/postgres?sslmode=disable",
		DBName:   "postgres",
		User:     "postgres",
		Password: "postgres",
		Host:     "postgres",
		Port:     "5432",
	}, nil
}
