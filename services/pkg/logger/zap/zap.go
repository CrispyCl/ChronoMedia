package zap

import (
	"context"
	"fmt"
	"os"

	"github.com/CrispyCl/ChronoMedia/services/pkg/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type zapLogger struct {
	serviceName string
	logger      *zap.Logger
}

func (l *zapLogger) Debug(ctx context.Context, msg string, fields ...logger.Field) {
	l.logger.Debug(msg, l.withContext(ctx, fields...)...)
}

func (l *zapLogger) Info(ctx context.Context, msg string, fields ...logger.Field) {
	l.logger.Info(msg, l.withContext(ctx, fields...)...)
}

func (l *zapLogger) Warn(ctx context.Context, msg string, fields ...logger.Field) {
	l.logger.Warn(msg, l.withContext(ctx, fields...)...)
}

func (l *zapLogger) Error(ctx context.Context, msg string, fields ...logger.Field) {
	l.logger.Error(msg, l.withContext(ctx, fields...)...)
}

func (l *zapLogger) Fatal(ctx context.Context, msg string, fields ...logger.Field) {
	l.logger.Fatal(msg, l.withContext(ctx, fields...)...)
}

func (l *zapLogger) With(fields ...logger.Field) logger.Logger {
	zapFields := l.toZapFields(fields...)

	return &zapLogger{
		serviceName: l.serviceName,
		logger:      l.logger.With(zapFields...),
	}
}

func (l *zapLogger) Sync() error {
	return l.logger.Sync()
}

func (l *zapLogger) withContext(ctx context.Context, fields ...logger.Field) []zap.Field {
	zapFields := l.toZapFields(fields...)

	zapFields = append(zapFields, zap.String(string(logger.ServiceName), l.serviceName))
	if v := ctx.Value(logger.RequestIDKey); v != nil {
		if reqID, ok := v.(string); ok {
			zapFields = append(zapFields, zap.String(string(logger.RequestID), reqID))
		}
	}
	return zapFields
}

func (l *zapLogger) toZapFields(fields ...logger.Field) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, f := range fields {
		if err, ok := f.Value.(error); ok {
			zapFields[i] = zap.Error(err)
		} else if s, ok := f.Value.(string); ok {
			zapFields[i] = zap.String(f.Key, s)
		} else if i, ok := f.Value.(int); ok {
			zapFields[i] = zap.Int(f.Key, i)
		} else {
			zapFields[i] = zap.Any(f.Key, f.Value)
		}
	}
	return zapFields
}

func NewZapLogger(serviceName string, env string) (*zapLogger, error) {
	if env == "dev" || env == "prod" {
		if err := os.MkdirAll("logs", 0755); err != nil {
			return nil, fmt.Errorf("failed to create logs directory: %w", err)
		}
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}

	options := []zap.Option{
		zap.AddCaller(),
		zap.AddCallerSkip(1), // Skips wrapper layer cleanly
	}

	switch env {
	case "test":
		return &zapLogger{
			serviceName: serviceName,
			logger:      zap.NewNop(),
		}, nil

	case "dev":
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05.000")

		consoleCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			zap.DebugLevel,
		)

		rotator := &lumberjack.Logger{
			Filename:   "./logs/dev.log",
			MaxSize:    50,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		}

		encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
		fileCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(rotator),
			zap.DebugLevel,
		)

		options = append(options, zap.AddStacktrace(zap.ErrorLevel))
		baseLogger := zap.New(zapcore.NewTee(consoleCore, fileCore), options...)
		return &zapLogger{serviceName: serviceName, logger: baseLogger}, nil

	case "prod":
		consoleCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			zap.InfoLevel,
		)

		rotator := &lumberjack.Logger{
			Filename:   "./logs/app.log",
			MaxSize:    100,
			MaxBackups: 5,
			MaxAge:     60,
			Compress:   true,
		}
		fileCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(rotator),
			zap.InfoLevel,
		)

		options := append(options, zap.AddStacktrace(zap.ErrorLevel))
		baseLogger := zap.New(zapcore.NewTee(consoleCore, fileCore), options...)
		return &zapLogger{serviceName: serviceName, logger: baseLogger}, nil

	case "local":
		fallthrough
	default:
		cfg := zap.Config{
			Level:             zap.NewAtomicLevelAt(zap.DebugLevel),
			Development:       true,
			DisableCaller:     false,
			DisableStacktrace: true,
			Encoding:          "console",
			OutputPaths:       []string{"stdout"},
			ErrorOutputPaths:  []string{"stderr"},
		}

		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05.000")
		cfg.EncoderConfig = encoderConfig

		baseLogger, err := cfg.Build(options...)
		if err != nil {
			return nil, fmt.Errorf("failed to build local configuration logger: %+v", err)
		}

		return &zapLogger{serviceName: serviceName, logger: baseLogger}, nil
	}
}
