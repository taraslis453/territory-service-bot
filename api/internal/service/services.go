package service

import (
	"fmt"

	"github.com/taraslis453/territory-service-bot/config"
	"github.com/taraslis453/territory-service-bot/pkg/logging"
	tb "gopkg.in/telebot.v3"
)

// Services stores all service layer interfaces
type Services struct {
	Bot BotService
}

// Options provides options for creating a new service instance via New.
type Options struct {
	Cfg      *config.Config
	Logger   logging.Logger
	Storages Storages
}

// serviceContext provides a shared context for all services
type serviceContext struct {
	cfg      *config.Config
	logger   logging.Logger
	storages Storages
}

type BotService interface {
	HandleStart(c tb.Context) error
	HandleMessage(c tb.Context, b *tb.Bot) error
	RenderMenu(c tb.Context) error
	HandleButton(c tb.Context) error
	HandleImageUpload(c tb.Context) error
}

var (
	MessageEnterCongregationName       = "Введіть назву збору, яку вам надав адміністратор ✍️"
	MessageUserNotFound                = "Ви не зареєстровані в системі. Зверніться до адміністратора збору 📞"
	MessageCongregationNotFound        = "Збір не знайдено 🤷"
	MessageCongregationAdminNotFound   = "Адміністраторa збору не знайдено 🤷"
	MessageUserIsNotAdmin              = "Ви не є адміністратором збору 🤷"
	MessageCongregationJoinRequestSent = func(congregationName string) string {
		return fmt.Sprintf("Запит на приєднання до збору *%s* відправлено. Очікуйте відповідь 😌", congregationName)
	}
	MessageWaitingForAdminApproval = "Очікуйте підтвердження адміністратора збору 😌"
	MessageNewJoinRequest          = func(options *MessageNewJoinRequestOptions) string {
		userFullName := fmt.Sprintf("%s %s", options.FirstName, options.LastName)
		if options.Username != "" {
			userFullName += fmt.Sprintf(" (@%s)", options.Username)
		}
		message := fmt.Sprint("Користувач ", userFullName, " хоче приєднатися")
		return message
	}
	MessageHowCanIHelpYou          = "Чим можу допомогти? 🙂"
	MessageAddTerritoryInstruction = "Надішліть зображення території де повідомлення відповідає зразку: *Група_назва* \nНаприклад: *Львів_123-а*, *Рівне_200* 📸"
	MessageTerritoryExistsInGroup  = func(title string, groupTitle string) string {
		return fmt.Sprintf("Територія з назвою *%s* вже існує в групі *%s* 🤷", title, groupTitle)
	}
	MessageNoTerritoriesFound = "Території не знайдено 🤷"
	MessageTerritoryList      = "Список кількості доступних територій:"
)

type MessageNewJoinRequestOptions struct {
	FirstName string
	LastName  string
	Username  string
}
