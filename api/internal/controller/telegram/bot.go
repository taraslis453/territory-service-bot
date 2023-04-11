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

	b.Handle(tb.OnPhoto, func(c tb.Context) error {
		return nil
	})
	b.Handle("/menu", options.Services.Bot.RenderMenu)
	b.Handle(tb.OnCallback, options.Services.Bot.HandleMenu)

	b.Start()

	return nil
}
