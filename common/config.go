package common

import "github.com/kelseyhightower/envconfig"

type envConfig struct {
	Mode string `envconfig:"APP_MODE"`
	Port string `envconfig:"APP_PORT"`
	Env  string `envconfig:"APP_ENV"`

	DBUser string `envconfig:"DB_USER" default:"postgres"`
	DBPass string `envconfig:"DB_PASSWORD" default:"password"`
	DBName string `envconfig:"DB_DATABASE" default:"event"`
	DBHost string `envconfig:"DB_HOST" default:"localhost"`
	DBPort string `envconfig:"DB_PORT" default:"5432"`

	SendGridApiKey string `envconfig:"SEND_GRID_KEY" default:"SENDGRID_KEY_REDACTED"`
}

var Config = new(envConfig)

func LoadConfig() {
	envconfig.Process("LIST", Config)
}
