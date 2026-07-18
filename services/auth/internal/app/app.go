package app

import (
	"context"
	"fmt"

	"github.com/CrispyCl/ChronoMedia/services/auth/internal/config"
	"github.com/CrispyCl/ChronoMedia/services/pkg/logger"
	"golang.org/x/sync/errgroup"
)

type App struct{}

func NewApp(ctx context.Context, cfg config.Config) (*App, error) {
	const op = "app.NewApp"

	_, ok := logger.GetLoggerFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("%s: failed to get logger from context", op)
	}

	return &App{}, nil
}

func (a *App) Start(ctx context.Context) error {
	const op = "App.Start"

	eg, ctx := errgroup.WithContext(ctx)

	// TODO: add http and gRPC servers starting

	return eg.Wait()
}

func (a *App) Stop(ctx context.Context) error {
	const op = "App.Stop"

	// TODO: add http and gRPC servers graceful stop

	return nil
}
