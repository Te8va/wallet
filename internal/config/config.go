package config

type Config struct {
	ServerAddress    string `env:"SERVER_ADDRESS"        envDefault:"0.0.0.0:8080"`
	ServicePort      int    `env:"SERVICE_PORT"          envDefault:"8080"`
	ServiceHost      string `env:"SERVICE_HOST"          envDefault:"0.0.0.0"`
	PostgresUser     string `env:"POSTGRES_USER"         envDefault:"wallet"`
	PostgresPassword string `env:"POSTGRES_PASSWORD"     envDefault:"wallet"`
	PostgresDB       string `env:"POSTGRES_DATABASE"     envDefault:"wallet"`
	PostgresPort     int    `env:"POSTGRES_PORT"         envDefault:"5432"`
	MigrationsPath   string `env:"MIGRATIONS_PATH"       envDefault:"migrations"`
	PostgresConn     string `env:"POSTGRES_CONN"         envDefault:"postgres://wallet:wallet@localhost:5432/wallet?sslmode=disable"`
}
