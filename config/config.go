package config

type Config struct {
	Database struct {
		Host     string `env:"DATABASE_HOST" env-default:"localhost" env-description:"Database host for service"`
		Port     int    `env:"DATABASE_PORT" env-default:"5432" env-description:"Database port for service"`
		Name     string `env:"DATABASE_NAME" env-default:"book_manager_db" env-description:"Database name for service"`
		User     string `env:"DATABASE_USER" env-default:"postgres" env-description:"Database user for service"`
		Password string `env:"DATABASE_PASSWORD" env-default:"postgresdev82" env-description:"Database password for service"`
	}
	JwtExpirationInMinutes int64 `env:"JWT_EXP_MINUTES" env-default:"10" env-description:"Jwt expiration minutes"`
}
