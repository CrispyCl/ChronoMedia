package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	User        string        `env:"REDIS_USER" env-default:""`
	Host        string        `env:"REDIS_HOST" env-default:"localhost"`
	Port        int           `env:"REDIS_PORT" env-default:"6379"`
	DB          int           `env:"REDIS_DB" env-default:"0"`
	Password    string        `env:"REDIS_PASSWORD" env-default:""`
	MaxRetries  int           `env:"REDIS_MAX_RETRIES" env-default:"3"`
	DialTimeout time.Duration `env:"REDIS_DIAL_TIMEOUT" env-default:"5s"`
	Timeout     time.Duration `env:"REDIS_TIMEOUT" env-default:"5s"`
}

type Cache struct {
	Client *redis.Client
}

func NewConnection(ctx context.Context, cfg Config) (*Cache, error) {
	const op = "storage.redis.NewConnection"
	const attemptsCount = 5

	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Username:     cfg.User,
		Password:     cfg.Password,
		DB:           cfg.DB,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		PoolSize:     25,
		MinIdleConns: 5,
	})

	var pingErr error

	for attempts := range attemptsCount {
		pingErr = rdb.Ping(ctx).Err()
		if pingErr == nil {
			return &Cache{Client: rdb}, nil
		}

		select {
		case <-ctx.Done():
			rdb.Close()
			return nil, fmt.Errorf("%s: %w", op, ctx.Err())
		case <-time.After(time.Duration(attempts+1) * time.Second):
			continue
		}
	}

	rdb.Close()
	return nil, fmt.Errorf("%s: redis unreachable after %d attempts :%w", op, attemptsCount, pingErr)
}

func (c *Cache) Close() error {
	const op = "storage.redis.Cache.Close"

	if c.Client != nil {
		if err := c.Client.Close(); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}
