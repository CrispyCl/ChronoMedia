package redis

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testHost string
	testPort int
	ctx      = context.Background()
)

func TestMain(m *testing.M) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:8-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp"),
	}

	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("could not start redis container: %v", err)
	}

	testHost, err = redisC.Host(ctx)
	if err != nil {
		log.Fatalf("could not get container host: %v", err)
	}

	port, err := redisC.MappedPort(ctx, "6379")
	if err != nil {
		log.Fatalf("could not get mapped port: %v", err)
	}
	testPort = int(port.Num())

	code := m.Run()

	if err := redisC.Terminate(ctx); err != nil {
		log.Printf("failed to terminate redis container: %v", err)
	}

	os.Exit(code)
}

func TestNewConnection(t *testing.T) {
	baseCfg := Config{
		Host:        testHost,
		Port:        testPort,
		DB:          0,
		MaxRetries:  1,
		DialTimeout: 1 * time.Second,
		Timeout:     1 * time.Second,
	}

	t.Run("connection success", func(t *testing.T) {
		cache, err := NewConnection(t.Context(), baseCfg)

		assert.NoError(t, err)
		require.NotNil(t, cache)
		assert.NotNil(t, cache.Client)

		if cache != nil {
			err = cache.Client.Set(t.Context(), "test_key", "active", 0).Err()
			assert.NoError(t, err)

			val, err := cache.Client.Get(t.Context(), "test_key").Result()
			assert.NoError(t, err)
			assert.Equal(t, "active", val)

			err = cache.Close()
			assert.NoError(t, err)
		}
	})

	t.Run("context cancelled", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(t.Context())
		cancel()

		cache, err := NewConnection(cancelCtx, baseCfg)

		assert.Nil(t, cache)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("max attempts reached", func(t *testing.T) {
		timeoutCtx, cancel := context.WithTimeout(t.Context(), 4*time.Second)
		defer cancel()

		deadCfg := baseCfg
		deadCfg.Port = 1111

		startTime := time.Now()
		cache, err := NewConnection(timeoutCtx, deadCfg)
		duration := time.Since(startTime)

		assert.Nil(t, cache)
		assert.Error(t, err)
		assert.GreaterOrEqual(t, duration.Seconds(), 1.0, "The retry backoff loop might be broken")
	})
}
