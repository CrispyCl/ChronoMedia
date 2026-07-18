package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/CrispyCl/ChronoMedia/services/auth/internal/app"
	"github.com/CrispyCl/ChronoMedia/services/auth/internal/config"
	"github.com/CrispyCl/ChronoMedia/services/pkg/logger"
	"github.com/CrispyCl/ChronoMedia/services/pkg/logger/zap"
)

const (
	serviceName = "auth"
)

func main() {
	cfg := config.MustLoad()

	mainLogger, err := zap.NewZapLogger(serviceName, cfg.Env)
	if err != nil {
		panic(fmt.Errorf("failed to load logger: %w", err))
	}
	defer mainLogger.Sync()

	ctx := logger.AddLoggerToContext(context.Background(), mainLogger)
	mainLogger.Debug(ctx, "config loaded", logger.Any("config", cfg))

	server, err := app.NewApp(ctx, cfg)
	if err != nil {
		mainLogger.Fatal(ctx, "failed to initializing server", logger.Error(err))
	}

	graceCh := make(chan os.Signal, 1)
	signal.Notify(graceCh, syscall.SIGTERM, syscall.SIGINT)

	errCh := make(chan error, 1)

	go func() {
		mainLogger.Info(ctx, "server.start")
		if err := server.Start(ctx); err != nil {
			errCh <- err
		}
	}()

	select {
	case sig := <-graceCh:
		mainLogger.Info(ctx, "received signal, stopping server", logger.String("signal", sig.String()))
	case err := <-errCh:
		mainLogger.Error(ctx, "server stopped unexpectedly", logger.Error(err))
	}

	if err := server.Stop(ctx); err != nil {
		mainLogger.Error(ctx, "error while stopping the server", logger.Error(err))
	}
	mainLogger.Info(ctx, "server.stop")
}
