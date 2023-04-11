package service

import (
	"fmt"
	"strings"

	"github.com/taraslis453/territory-service-bot/internal/entity"
	tb "gopkg.in/telebot.v3"
)

type botService struct {
	serviceContext
}

var _ BotService = (*botService)(nil)

func NewBotService(options *Options) *botService {
	return &botService{
		serviceContext: serviceContext{
			cfg:      options.Cfg,
			logger:   options.Logger.Named("BotService"),
			storages: options.Storages,
		},
	}
}

func (s *botService) RenderMenu(c tb.Context) error {
	logger := s.logger.
		Named("RenderMenu").
		With(c)

	var buttons [][]string
	isAdmin := true
	if isAdmin {
		buttons = [][]string{
			{entity.ViewTerritoryListButton},
			{entity.AddTerritoryButton},
		}
	} else {
		buttons = [][]string{
			{entity.ViewTerritoryListButton},
		}
	}
	logger.With("buttons", buttons)
	logger.Info("successfully rendered menu buttons")

	return c.Send("Чим можу допомогти?", &tb.SendOptions{
		ReplyMarkup: s.renderInlineKeyboard(buttons),
	})
}

func (s *botService) renderInlineKeyboard(buttons [][]string) *tb.ReplyMarkup {
	inlineKeyboard := &tb.ReplyMarkup{}

	for _, row := range buttons {
		var buttonRow []tb.InlineButton
		for _, buttonText := range row {
			button := tb.InlineButton{
				// NOTE: we relly on the fact that button text is unique
				Unique: buttonText,
				Text:   buttonText,
			}
			buttonRow = append(buttonRow, button)
		}
		inlineKeyboard.InlineKeyboard = append(inlineKeyboard.InlineKeyboard, buttonRow)
	}

	return inlineKeyboard
}

func (s *botService) HandleMenu(c tb.Context) error {
	var err error
	defer func() {
		err = c.Respond() // Respond to the callback to remove the loading state
	}()
	if err != nil {
		return err
	}

	// TODO: check for admin or publisher allow actions

	data := c.Data()
	data = strings.Replace(data, "\f", "", -1)

	switch data {
	case entity.ViewTerritoryListButton:
		return c.Send("You clicked view territory list button!")
	case entity.AddTerritoryButton:
		return c.Send("You clicked add territory button!")
	default:
		return fmt.Errorf("unknown button: %s", data)
	}
}
