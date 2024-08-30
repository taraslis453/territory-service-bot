package telegram

import (
	"bytes"
	"runtime/debug"
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

func NewBot(options *Options) error {
	pref := tb.Settings{
		Token:  options.Config.Telegram.BotToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tb.NewBot(pref)
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
