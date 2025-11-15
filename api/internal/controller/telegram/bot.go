package telegram

import (
	"bytes"
	"errors"
	"net"
	"runtime/debug"
	"strings"
	"time"

	"github.com/DataDog/gostackparse"
	"github.com/taraslis453/territory-service-bot/config"
	"github.com/taraslis453/territory-service-bot/internal/service"
	"github.com/taraslis453/territory-service-bot/pkg/logging"

	tb "gopkg.in/telebot.v3"
)

type Options struct {
	Services service.Services
	Storages service.Storages
	Logger   logging.Logger
	Config   *config.Config
}

const (
	maxRetries   = 5
	initialDelay = 1 * time.Second
	maxDelay     = 30 * time.Second
)

func NewBot(options *Options) error {
	pref := tb.Settings{
		Token:  options.Config.Telegram.BotToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	}

	var b *tb.Bot
	err := retryWithBackoff(options.Logger, func() error {
		var retryErr error
		b, retryErr = tb.NewBot(pref)
		return retryErr
	}, "telegram.NewBot")
	if err != nil {
		return err
	}

	b.Handle("/start", func(c tb.Context) error {
		return wrapHandler(c, b, options.Logger, options.Services.Bot.HandleStart)
	})
	b.Handle(tb.OnText, func(c tb.Context) error {
		return wrapHandler(c, b, options.Logger, options.Services.Bot.HandleMessage)
	})
	b.Handle("/menu", func(c tb.Context) error {
		return wrapHandler(c, b, options.Logger, options.Services.Bot.RenderMenu)
	})
	b.Handle(tb.OnCallback, func(c tb.Context) error {
		return wrapHandler(c, b, options.Logger, options.Services.Bot.HandleInlineButton)
	})
	b.Handle(tb.OnPhoto, func(c tb.Context) error {
		return wrapHandler(c, b, options.Logger, options.Services.Bot.HandleImageUpload)
	})
	b.Handle(tb.OnDocument, func(c tb.Context) error {
		return wrapHandler(c, b, options.Logger, options.Services.Bot.HandleDocumentUpload)
	})

	b.Start()

	return nil
}

// isRetryableError checks if an error is retryable (network/TLS errors)
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	// Check for TLS handshake timeout
	if strings.Contains(errStr, "TLS handshake timeout") {
		return true
	}
	// Check for other network-related errors
	if strings.Contains(errStr, "timeout") {
		return true
	}
	if strings.Contains(errStr, "connection refused") {
		return true
	}
	if strings.Contains(errStr, "no such host") {
		return true
	}
	// Check for network errors
	var netErr net.Error
	return errors.As(err, &netErr)
}

// retryWithBackoff retries a function with exponential backoff
func retryWithBackoff(logger logging.Logger, fn func() error, operation string) error {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Calculate delay with exponential backoff: 1s, 2s, 4s, 8s, 16s (capped at maxDelay)
			delay := initialDelay * time.Duration(1<<uint(attempt-1))
			if delay > maxDelay {
				delay = maxDelay
			}
			logger.Warn("retrying operation",
				"operation", operation,
				"attempt", attempt+1,
				"maxAttempts", maxRetries,
				"delay", delay,
				"lastError", lastErr)
			time.Sleep(delay)
		}

		err := fn()
		if err == nil {
			if attempt > 0 {
				logger.Info("operation succeeded after retries",
					"operation", operation,
					"attempts", attempt+1)
			}
			return nil
		}

		lastErr = err

		// Only retry if it's a retryable error
		if !isRetryableError(err) {
			logger.Error("non-retryable error encountered",
				"operation", operation,
				"error", err)
			return err
		}

		logger.Warn("retryable error encountered",
			"operation", operation,
			"attempt", attempt+1,
			"error", err)
	}

	logger.Error("operation failed after all retries",
		"operation", operation,
		"maxAttempts", maxRetries,
		"lastError", lastErr)
	return lastErr
}

func wrapHandler(c tb.Context, b *tb.Bot, logger logging.Logger, handler func(c tb.Context, b *tb.Bot) error) error {
	defer func() {
		if r := recover(); r != nil {
			stacktrace, errors := gostackparse.Parse(bytes.NewReader(debug.Stack()))
			if len(errors) > 0 || len(stacktrace) == 0 {
				logger.Error("get stacktrace errors", "stacktraceErrors", errors, "stacktrace", "unknown", "err", r)
			} else {
				logger.Error("unhandled error", "err", r, "stacktrace", stacktrace)
			}
		}
	}()

	return handler(c, b)
}
