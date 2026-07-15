package postgres

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
		Image:        "postgres:18-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "root",
			"POSTGRES_PASSWORD": "111",
			"POSTGRES_DB":       "test",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
	}

	postgresC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("could not start container: %v", err)
	}

	testHost, err = postgresC.Host(ctx)
	if err != nil {
		log.Fatalf("could not get container host: %v", err)
	}

	port, err := postgresC.MappedPort(ctx, "5432")
	if err != nil {
		log.Fatalf("could not get mapped port: %v", err)
	}
	testPort = int(port.Num())

	code := m.Run()

	if err := postgresC.Terminate(ctx); err != nil {
		log.Printf("failed to terminate postgres container: %v", err)
	}

	os.Exit(code)
}

func TestNewConnection(t *testing.T) {
	cfg := Config{
		UserName: "root",
		Password: "111",
		Host:     testHost,
		Port:     testPort,
		DBName:   "test",
		SSLMode:  "disable",
	}

	t.Run("connection success", func(t *testing.T) {
		db, err := NewConnection(t.Context(), cfg)

		assert.NoError(t, err)
		assert.NotNil(t, db)
		if db != nil {
			assert.NotNil(t, db.Pool)
			db.Close()
		}
	})

	t.Run("context cancelled", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(t.Context())
		cancel()

		db, err := NewConnection(cancelCtx, cfg)

		assert.Nil(t, db)
		assert.ErrorIs(t, err, context.Canceled)
	})
}
