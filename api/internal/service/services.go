package service

import (
	"fmt"
	"time"

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
	HandleStart(c tb.Context, b *tb.Bot) error
	HandleMessage(c tb.Context, b *tb.Bot) error
	RenderMenu(c tb.Context, b *tb.Bot) error
	HandleInlineButton(c tb.Context, b *tb.Bot) error
	HandleImageUpload(c tb.Context, b *tb.Bot) error
	HandleDocumentUpload(c tb.Context, b *tb.Bot) error
}

func GenerateContactLink(fullName string, userID int64) string {
	return fmt.Sprintf("[%s](tg://user?id=%d)", fullName, userID)
}

var (
	MessageEnterFullName               = "Як мені тебе запам'ятати? (ім’я та фамілія) ✍️"
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
		}// we can delete the username ?? 

		//NOTE: if the user disabled  forwarding in the privacy settings the link won't work
		contactLink := fmt.Sprintf("<a href=\"tg://user?id=%d\">%s</a>",  options.UserID, userFullName)
		message := fmt.Sprintf("%s хоче приєднатися", contactLink)
		return message
	}

	MessageCongregationJoinRequestApprovedDone = func(fullName string, userID int64) string {
		contactLink := GenerateContactLink(fullName, userID)
		return fmt.Sprintf("Вісника %s приєднано до збору ✅", contactLink)
	}
	MessageCongregationJoinRequestRejectedDone = func(fullName string, userID int64) string {
		contactLink := GenerateContactLink(fullName, userID)
		return fmt.Sprintf("Користувача %s відхилено ❌", contactLink)
	}
	MessageCongregationJoinRequestApproved = "Запит на приєднання до збору прийнято 🎉"
	MessageCongregationJoinRequestRejected = "Запит на приєднання до збору відхилено 😔"

	MessageHowCanIHelpYou          = "Чим можу допомогти? 🙂"
	MessageAddTerritoryInstruction = "Надішли зображення або документ території де повідомлення відповідає зразку: *Група_назва* \nНаприклад: *Львів_123-а*, *Рівне_200* 📸"
	MessageTerritoryExistsInGroup  = func(title string, groupTitle string) string {
		return fmt.Sprintf("Територія з назвою *%s* вже існує в групі *%s* 🤷", title, groupTitle)
	}
	MessageNoTerritoriesFound              = "Території не знайдено 🤷"
	MessageTerritoryNotFound               = "Територія не знайдена 🤷"
	MessageTerritoryNotAvailable           = "Територія не доступна 🤷"
	MessageTerritoryList                   = "Список доступних територій: "
	MessageMyTerritoryListTerritoryCaption = func(title string, lastTakenAt time.Time, notes []string) string {
		caption := fmt.Sprintf("Територія: %s\n%s", title, lastTakenAt.Format("02.01.2006"))
		if len(notes) > 0 {
			caption += "\n\n"
			caption += "Нотатки:\n"
			for _, note := range notes {
				caption += fmt.Sprintf("📌 %s\n", note)
			}
		}
		return caption
	}
	MessageTerritoryListTerritoryCaption = func(options MessageTerritoryListTerritoryCaptionOptions) string {
		caption := fmt.Sprintf("Територія: %s", options.Title)
		if !options.LastTakenAt.IsZero() {
			caption += fmt.Sprintf("\nОстаннє опрацювання: *%s*", options.LastTakenAt.Format("02.01.2006"))
		}

		if options.UserRole == entity.UserRoleAdmin {
			if options.InUseByFullName != "" {
				clickableName := GenerateContactLink(options.InUseByFullName, options.InUseByUserID)
				caption += fmt.Sprintf("\nВикористовує: %s", clickableName)
			}

			if len(options.Notes) > 0 {
				caption += "\n\n"
				caption += "Нотатки:\n"
				for _, note := range options.Notes {
					caption += fmt.Sprintf("📌 %s\n", note)
				}
			}
		}
		return caption
	}

	MessageTakeTerritoryRequest = func(user *entity.User, territoryTitle string) string {
		contactLink := fmt.Sprintf("<a href=\"tg://user?id=%s\">%s</a>", user.MessengerUserID, user.FullName)
		return fmt.Sprintf("%s хоче взяти %s", contactLink, territoryTitle)
	}

	MessageTakeTerritoryRequestSent = "Запит на взяття території відправлено. Очікуй відповідь 😌"

	MessageTakeTerritoryRequestApproved = func(territoryTitle string, notes []string) string {
		message := fmt.Sprintf("Запит на взяття території *%s* прийнято ✅", territoryTitle)
		if len(notes) > 0 {
			message += "\n\n"
			message += "Нотатки:\n"
			for _, note := range notes {
				message += fmt.Sprintf("📌 %s\n", note)
			}
		}
		return message
	}
	MessageTakeTerritoryRequestApprovedDone = func(fullName string, userID int64, territoryName string) string {
		clickableName := GenerateContactLink(fullName, userID)
		return fmt.Sprintf("Вісника %s призначено на територію *%s* ✅", clickableName, territoryName)
	}//dont make the username link *bold*, it brakes the formatting 

	MessageTakeTerritoryRequestRejected = func(territoryTitle string) string {
		return fmt.Sprintf("Запит на взяття території *%s* відхилено ❌", territoryTitle)
	}
	MessageTakeTerritoryRequestRejectedDone = func(fullName string, userID int64, territoryTitle string) string {
		clickableName := GenerateContactLink(fullName, userID)
		return fmt.Sprintf("Вісника %s відхилено на територію *%s* ❌", clickableName, territoryTitle)
	}

	MessagePublisherReturnedTerritory = func(fullName string, userID int64, territoryTitle string) string {
		contactLink := GenerateContactLink(fullName, userID)
		return fmt.Sprintf("Вісник %s повернув територію *%s* ✅", contactLink, territoryTitle)
	}
	MessageLeaveTerritoryNote = func(territoryTitle string) string {
		return fmt.Sprintf("Залишіть нотатку для території %s ✍️", territoryTitle)
	}
	MessageTerritoryNotInUse        = "Територія не використовується 🤷"
	MessageTerritoryCannotLeaveNote = "Ви не можете залишити нотатку для цієї території 🤷"
	MessageTerritoryNoteSaved       = "Нотатку збережено ✅"

	MessageTerritoryReturned = "Територію повернуто ✅"

	MessagePublisherNotFound = "Вісника не знайдено 🤷"
)

type MessageNewJoinRequestOptions struct {
	FirstName string
	LastName  string
	Username  string
	UserID    int64
}

type MessageTerritoryListTerritoryCaptionOptions struct {
	UserRole        entity.UserRole
	Title           string
	LastTakenAt     time.Time
	Notes           []string
	InUseByFullName string
	InUseByUserID   int64
}
