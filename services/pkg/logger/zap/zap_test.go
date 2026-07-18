package zap

import (
	"context"
	"errors"
	"testing"

	"github.com/CrispyCl/ChronoMedia/services/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewZapLogger(t *testing.T) {
	testCases := []struct {
		name        string
		serviceName string
		env         string
	}{
		{
			name:        "test environment initialization",
			serviceName: "test-service",
			env:         "test",
		},
		{
			name:        "local environment initialization",
			serviceName: "test-service",
			env:         "local",
		},
		{
			name:        "dev environment initialization",
			serviceName: "test-service",
			env:         "dev",
		},
		{
			name:        "prod environment initialization",
			serviceName: "test-service",
			env:         "prod",
		},
		{
			name:        "default fallback initialization",
			serviceName: "test-service",
			env:         "unknown",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			l, err := NewZapLogger(tC.serviceName, tC.env)

			assert.NoError(t, err)
			assert.NotNil(t, l)
			assert.Equal(t, tC.serviceName, l.serviceName)
		})
	}
}

func TestZapLogger_Lifecycle(t *testing.T) {
	const serviceName = "test-service"

	core, observedLogs := observer.New(zapcore.DebugLevel)
	baseLogger := zap.New(core)

	l := &zapLogger{
		serviceName: serviceName,
		logger:      baseLogger,
	}

	t.Run("successful logger with all levels and fields mapping", func(t *testing.T) {
		observedLogs.TakeAll()

		testErr := errors.New("database connection timeout")

		l.Debug(t.Context(), "debug msg", logger.String("str_key", "str_val"))
		l.Info(t.Context(), "info msg", logger.Int("int_key", 42))
		l.Warn(t.Context(), "warn msg", logger.Error(testErr))
		l.Error(t.Context(), "error msg", logger.Any("any_key", []int{1, 2}))

		logs := observedLogs.All()
		require.Len(t, logs, 4)

		assert.Equal(t, zapcore.DebugLevel, logs[0].Level)
		assert.Equal(t, "debug msg", logs[0].Message)
		fields0 := logs[0].ContextMap()
		assert.Equal(t, "str_val", fields0["str_key"])
		assert.Equal(t, serviceName, fields0[string(logger.ServiceName)])

		assert.Equal(t, zapcore.InfoLevel, logs[1].Level)
		assert.Equal(t, "info msg", logs[1].Message)
		fields1 := logs[1].ContextMap()
		assert.Equal(t, int64(42), fields1["int_key"])
		assert.Equal(t, serviceName, fields1[string(logger.ServiceName)])

		assert.Equal(t, zapcore.WarnLevel, logs[2].Level)
		assert.Equal(t, "warn msg", logs[2].Message)
		fields2 := logs[2].ContextMap()
		assert.Equal(t, testErr.Error(), fields2["error"])
		assert.Equal(t, serviceName, fields2[string(logger.ServiceName)])

		assert.Equal(t, zapcore.ErrorLevel, logs[3].Level)
		assert.Equal(t, "error msg", logs[3].Message)
		fields3 := logs[3].ContextMap()
		assert.Equal(t, []any{1, 2}, fields3["any_key"])
		assert.Equal(t, serviceName, fields3[string(logger.ServiceName)])
	})

	t.Run("context integration with request id", func(t *testing.T) {
		observedLogs.TakeAll()

		reqID := "test-req-id-123"
		ctxWithReqID := context.WithValue(t.Context(), logger.RequestIDKey, reqID)

		l.Info(ctxWithReqID, "operation complete")

		logs := observedLogs.All()
		require.Len(t, logs, 1)

		fields := logs[0].ContextMap()
		assert.Equal(t, serviceName, fields[string(logger.ServiceName)])
		assert.Equal(t, reqID, fields[string(logger.RequestID)])
	})

	t.Run("with method success sub-logger creating", func(t *testing.T) {
		observedLogs.TakeAll()

		subLogger := l.With(logger.String("component", "handler"), logger.Int("version", 2))
		require.NotNil(t, subLogger)

		subLogger.Info(t.Context(), "test request", logger.String("extra", "value"))

		logs := observedLogs.All()
		require.Len(t, logs, 1)

		fields := logs[0].ContextMap()
		assert.Equal(t, serviceName, fields[string(logger.ServiceName)])
		assert.Equal(t, "handler", fields["component"])
		assert.Equal(t, int64(2), fields["version"])
		assert.Equal(t, "value", fields["extra"])
	})

	t.Run("sync triggers without panic", func(t *testing.T) {
		err := l.Sync()
		assert.NoError(t, err)
	})

	t.Run("logger context helpers", func(t *testing.T) {
		logFromCtx, ok := logger.GetLoggerFromContext(t.Context())
		assert.False(t, ok)
		assert.Nil(t, logFromCtx)

		ctxWithLog := logger.AddLoggerToContext(t.Context(), l)

		logFromCtx, ok = logger.GetLoggerFromContext(ctxWithLog)
		assert.True(t, ok)
		assert.NotNil(t, logFromCtx)
	})
}
