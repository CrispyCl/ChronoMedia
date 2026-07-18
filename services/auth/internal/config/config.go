package config

import (
	"fmt"
	"os"

	"github.com/CrispyCl/ChronoMedia/services/pkg/storage/postgres"
	"github.com/CrispyCl/ChronoMedia/services/pkg/storage/redis"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Postgres postgres.Config
	Redis    redis.Config

	Env            string `env:"ENV" env-default:"local"`
	HTTPServerPort int    `env:"HTTP_SERVER_PORT" env-default:"8081"`
	GRPCServerPort int    `env:"GRPC_SERVER_PORT" env-default:"50051"`
}

func MustLoad() Config {
	const op = "config.MustLoad"
	const configPath = "config/.env"

	var cfg Config

	if _, err := os.Stat(configPath); err == nil {
		if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
			panic(fmt.Errorf("%s: failed to read config file from %s: %w", op, configPath, err))
		}
		return cfg
	}

	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		panic(fmt.Errorf("%s: failed to read system environment variables: %w", op, err))
	}
	return cfg
}
