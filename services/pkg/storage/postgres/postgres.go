package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	UserName string `env:"POSTGRES_USER" env-default:"root"`
	Password string `env:"POSTGRES_PASSWORD" env-default:"111"`
	Host     string `env:"POSTGRES_HOST" env-default:"localhost"`
	Port     int    `env:"POSTGRES_PORT" env-default:"5432"`
	DBName   string `env:"POSTGRES_DB" env-default:"GriBD"`
	SSLMode  string `env:"POSTGRES_SSL" env-default:"disable"`
}

type DB struct {
	Pool *pgxpool.Pool
}

func NewConnection(ctx context.Context, cfg Config) (*DB, error) {
	const op = "storage.postgres.NewConnection"
	const attemptsCount = 5

	dsn := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s host=%s port=%d", cfg.UserName, cfg.Password, cfg.DBName, cfg.SSLMode, cfg.Host, cfg.Port)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to parse postgres dsn string (%s): %w", op, dsn, err)
	}

	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 5 * time.Minute
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	var pool *pgxpool.Pool
	var pingErr error

	for attempts := range attemptsCount {
		pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
		if err == nil {
			pingErr = pool.Ping(ctx)
			if pingErr == nil {
				return &DB{Pool: pool}, nil
			}
			pool.Close()
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%s: %w", op, ctx.Err())
		case <-time.After(time.Duration(attempts+1) * time.Second):
			continue
		}
	}

	return nil, fmt.Errorf("%s: database unreachable after %d attempts: %w", op, attemptsCount, pingErr)
}

func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}
