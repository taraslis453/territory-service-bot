package logging

import (
	"context"
	"os"

	zapLib "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zap struct {
	logger *zapLib.SugaredLogger
}

var _ Logger = (*zap)(nil)

// NewZap - creates new zap logger instance with given level
func NewZap(level string) *zap {
	var l zapcore.Level
	l, err := zapcore.ParseLevel(level)
	if err != nil {
		l = zapLib.InfoLevel
	}

	// logger config
	config := zapLib.Config{
		Development:      false,
		Encoding:         "json",
		Level:            zapLib.NewAtomicLevelAt(l),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			EncodeDuration: zapcore.SecondsDurationEncoder,
			LevelKey:       "severity",
			EncodeLevel:    zapcore.CapitalLevelEncoder, // e.g. "Info"
			CallerKey:      "caller",
			EncodeCaller:   zapcore.ShortCallerEncoder, // e.g. package/file:line
			TimeKey:        "timestamp",
			EncodeTime:     zapcore.ISO8601TimeEncoder, // e.g. 2020-05-05T03:24:36.903+0300
			NameKey:        "name",
			EncodeName:     zapcore.FullNameEncoder, // e.g. GetSiteGeneralHandler
			MessageKey:     "message",
			StacktraceKey:  "",
			LineEnding:     "\n",
		},
	}

	// build logger from config
	logger, _ := config.Build()

	// configure and create logger
	return &zap{
		logger: logger.Sugar(),
	}
}

func (l *zap) Named(name string) Logger {
	return &zap{
		logger: l.logger.Named(name),
	}
}

func (l *zap) With(args ...interface{}) Logger {
	return &zap{
		logger: l.logger.With(args...),
	}
}

// TODO: remove RequestID and make the method generic (how?).
func (l *zap) WithContext(ctx context.Context) Logger {
	return l.With("RequestID", ctx.Value("RequestID"))
}

func (l *zap) Debug(message string, args ...interface{}) {
	l.logger.Debugw(message, args...)
}

func (l *zap) Info(message string, args ...interface{}) {
	l.logger.Infow(message, args...)
}

func (l *zap) Warn(message string, args ...interface{}) {
	l.logger.Warnw(message, args...)
}

func (l *zap) Error(message string, args ...interface{}) {
	l.logger.Errorw(message, args...)
}

func (l *zap) Fatal(message string, args ...interface{}) {
	l.logger.Fatalw(message, args...)
	os.Exit(1)
}

func (l *zap) Unwrap() *zapLib.Logger {
	return l.logger.Desugar()
}
