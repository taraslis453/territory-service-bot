package service

import (
	"fmt"

	"github.com/taraslis453/territory-service-bot/config"
	"github.com/taraslis453/territory-service-bot/internal/entity"
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
	HandleButton(c tb.Context, b *tb.Bot) error
	HandleImageUpload(c tb.Context) error
}

var (
	MessageEnterFullName               = "Як мені тебе запам'ятати? ✍️"
	MessageEnterCongregationName       = "З якого ти збору? ✍️"
	MessageUserNotFound                = "Ти не зареєстрований в системі. Звернись до адміністратора збору 📞"
	MessageCongregationNotFound        = "Збір не знайдено 🤷"
	MessageCongregationAdminNotFound   = "Адміністраторa збору не знайдено 🤷"
	MessageUserIsNotAdmin              = "Ти не є адміністратором збору 🤷"
	MessageCongregationJoinRequestSent = func(congregationName string) string {
		return fmt.Sprintf("Запит на приєднання до збору *%s* відправлено. Очікуй відповідь 😌", congregationName)
	}
	MessageWaitingForAdminApproval = "Очікуй підтвердження адміністратора збору 😌"
	MessageNewJoinRequest          = func(options *MessageNewJoinRequestOptions) string {
		userFullName := fmt.Sprintf("%s %s", options.FirstName, options.LastName)
		if options.Username != "" {
			userFullName += fmt.Sprintf(" (@%s)", options.Username)
		}
		message := fmt.Sprint(userFullName, " хоче приєднатися")
		return message
	}
	MessageCongregationJoinRequestApprovedDone = func(fullName string) string {
		return fmt.Sprintf("Вісника *%s* приєднано до збору ✅", fullName)
	}
	MessageCongregationJoinRequestRejectedDone = func(fullName string) string {
		return fmt.Sprintf("Користувача *%s* відхилено ❌", fullName)
	}
	MessageCongregationJoinRequestApproved = "Запит на приєднання до збору прийнято 🎉"
	MessageCongregationJoinRequestRejected = "Запит на приєднання до збору відхилено 😔"

	MessageHowCanIHelpYou          = "Чим можу допомогти? 🙂"
	MessageAddTerritoryInstruction = "Надішли зображення території де повідомлення відповідає зразку: *Група_назва* \nНаприклад: *Львів_123-а*, *Рівне_200* 📸"
	MessageTerritoryExistsInGroup  = func(title string, groupTitle string) string {
		return fmt.Sprintf("Територія з назвою *%s* вже існує в групі *%s* 🤷", title, groupTitle)
	}
	MessageNoTerritoriesFound    = "Території не знайдено 🤷"
	MessageTerritoryNotFound     = "Територія не знайдена 🤷"
	MessageTerritoryNotAvailable = "Територія не доступна 🤷"
	MessageTerritoryList         = "Список доступних територій: "
	MessageTakeTerritoryRequest  = func(user *entity.User, territoryTitle string) string {
		return fmt.Sprintf("%s хоче взяти %s", user.FullName, territoryTitle)
	}
	MessageTakeTerritoryRequestSent = "Запит на взяття території відправлено. Очікуй відповідь 😌"

	MessageTakeTerritoryRequestApproved = func(territoryTitle string) string {
		return fmt.Sprintf("Запит на взяття території %s прийнято ✅", territoryTitle)
	}
	MessageTakeTerritoryRequestApprovedDone = func(fullName string, territoryName string) string {
		return fmt.Sprintf("Вісника *%s* призначено на територію *%s* ✅", fullName, territoryName)
	}

	MessageTakeTerritoryRequestRejected = func(fullName string, territoryTitle string) string {
		return fmt.Sprintf("Запит на взяття території *%s* від *%s* відхилено ❌", territoryTitle, fullName)
	}

	MessagePublisherReturnedTerritory = func(fullName string, territoryTitle string) string {
		return fmt.Sprintf("Вісник *%s* повернув територію *%s* ✅", fullName, territoryTitle)
	}
	MessageTerritoryReturned = "Територію повернуто ✅"

	MessagePublisherNotFound = "Вісника не знайдено 🤷"
)

type MessageNewJoinRequestOptions struct {
	FirstName string
	LastName  string
	Username  string
}
