package config

import (
	"log"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	DBSource      string `env:"DB_SOURCE,required"`
	JWTSecret     string `env:"JWT_SECRET,required"`
	ServerAddress string `env:"SERVER_ADDRESS" envDefault:":8080"`
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	cfg := &Config{}
	return cfg, env.Parse(cfg)
}
