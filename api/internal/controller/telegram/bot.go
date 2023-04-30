package telegram

import (
	"time"

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

	b.Handle("/start", options.Services.Bot.HandleStart)
	b.Handle(tb.OnText, func(c tb.Context) error {
		return options.Services.Bot.HandleMessage(c, b)
	})
	b.Handle("/menu", options.Services.Bot.RenderMenu)
	b.Handle(tb.OnCallback, func(c tb.Context) error {
		return options.Services.Bot.HandleInlineButton(c, b)
	})
	b.Handle(tb.OnPhoto, options.Services.Bot.HandleImageUpload)

	b.Start()

	return nil
}
