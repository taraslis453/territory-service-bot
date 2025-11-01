package service

import (
	"fmt"
	"html"
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

var (
	MessageEnterFullName               = "Як мені тебе запам'ятати? (ім’я та фамілія) ✍️"
	MessageEnterCongregationName       = "З якого ти збору? ✍️"
	MessageUserNotFound                = "Ти не зареєстрований в системі. Звернись до адміністратора збору 📞"
	MessageCongregationNotFound        = "Збір не знайдено 🤷"
	MessageCongregationAdminNotFound   = "Адміністраторa збору не знайдено 🤷"
	MessageUserIsNotAdmin              = "Ти не є адміністратором збору 🤷"
	MessageCongregationJoinRequestSent = func(congregationName string) string {
		escapedName := html.EscapeString(congregationName)
		return fmt.Sprintf("Запит на приєднання до збору <b>%s</b> відправлено. Очікуй відповідь 😌", escapedName)
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
		escapedFullName := html.EscapeString(fullName)
		return fmt.Sprintf("Вісника <b>%s</b> приєднано до збору ✅", escapedFullName)
	}
	MessageCongregationJoinRequestRejectedDone = func(fullName string) string {
		escapedFullName := html.EscapeString(fullName)
		return fmt.Sprintf("Користувача <b>%s</b> відхилено ❌", escapedFullName)
	}
	MessageCongregationJoinRequestApproved = "Запит на приєднання до збору прийнято 🎉"
	MessageCongregationJoinRequestRejected = "Запит на приєднання до збору відхилено 😔"

	MessageHowCanIHelpYou          = "Чим можу допомогти? 🙂"
	MessageAddTerritoryInstruction = "Надішли зображення або документ території де повідомлення відповідає зразку: *Група_назва* \nНаприклад: *Львів_123-а*, *Рівне_200* 📸"
	MessageTerritoryExistsInGroup  = func(title string, groupTitle string) string {
		escapedTitle := html.EscapeString(title)
		escapedGroupTitle := html.EscapeString(groupTitle)
		return fmt.Sprintf("Територія з назвою <b>%s</b> вже існує в групі <b>%s</b> 🤷", escapedTitle, escapedGroupTitle)
	}
	MessageNoTerritoriesFound              = "Території не знайдено 🤷"
	MessageTerritoryNotFound               = "Територія не знайдена 🤷"
	MessageTerritoryNotAvailable           = "Територія не доступна 🤷"
	MessageTerritoryList                   = "Список доступних територій: "
	MessageMyTerritoryListTerritoryCaption = func(title string, lastTakenAt time.Time, note string) string {
		// Use HTML to safely display user-generated content
		escapedTitle := html.EscapeString(title)
		caption := fmt.Sprintf("Територія: %s\n%s", escapedTitle, lastTakenAt.Format("02.01.2006"))
		if note != "" {
			escapedNote := html.EscapeString(note)
			caption += "\n\n"
			caption += "Нотатка:\n"
			caption += fmt.Sprintf("📌 %s\n", escapedNote)
		}
		return caption
	}
	MessageTerritoryListTerritoryCaption = func(options MessageTerritoryListTerritoryCaptionOptions) string {
		// Use HTML to safely display user-generated content
		escapedTitle := html.EscapeString(options.Title)
		caption := fmt.Sprintf("Територія: %s", escapedTitle)
		if !options.LastTakenAt.IsZero() {
			caption += fmt.Sprintf("\nОстаннє опрацювання: <b>%s</b>", options.LastTakenAt.Format("02.01.2006"))
		}

		if options.UserRole == entity.UserRoleAdmin {
			if options.InUseByFullName != "" {
				escapedFullName := html.EscapeString(options.InUseByFullName)
				caption += fmt.Sprintf("\nВикористовує: <b>%s</b>", escapedFullName)
			}

			if options.Note != "" {
				escapedNote := html.EscapeString(options.Note)
				caption += "\n\n"
				caption += "Нотатка:\n"
				caption += fmt.Sprintf("📌 %s\n", escapedNote)
			}
		}
		return caption
	}

	MessageTakeTerritoryRequest = func(user *entity.User, territoryTitle string) string {
		escapedFullName := html.EscapeString(user.FullName)
		escapedTitle := html.EscapeString(territoryTitle)
		return fmt.Sprintf("%s хоче взяти %s", escapedFullName, escapedTitle)
	}
	MessageTakeTerritoryRequestSent = "Запит на взяття території відправлено. Очікуй відповідь 😌"

	MessageTakeTerritoryRequestApproved = func(territoryTitle string, note string) string {
		escapedTitle := html.EscapeString(territoryTitle)
		message := fmt.Sprintf("Запит на взяття території <b>%s</b> прийнято ✅", escapedTitle)
		if note != "" {
			escapedNote := html.EscapeString(note)
			message += "\n\n"
			message += "Нотатка:\n"
			message += fmt.Sprintf("📌 %s\n", escapedNote)
		}
		return message
	}
	MessageTakeTerritoryRequestApprovedDone = func(fullName string, territoryName string) string {
		escapedFullName := html.EscapeString(fullName)
		escapedTerritoryName := html.EscapeString(territoryName)
		return fmt.Sprintf("Вісника <b>%s</b> призначено на територію <b>%s</b> ✅", escapedFullName, escapedTerritoryName)
	}

	MessageTakeTerritoryRequestRejected = func(territoryTitle string) string {
		escapedTitle := html.EscapeString(territoryTitle)
		return fmt.Sprintf("Запит на взяття території <b>%s</b> відхилено ❌", escapedTitle)
	}
	MessageTakeTerritoryRequestRejectedDone = func(fullName string, territoryTitle string) string {
		escapedFullName := html.EscapeString(fullName)
		escapedTitle := html.EscapeString(territoryTitle)
		return fmt.Sprintf("Вісника <b>%s</b> відхилено на територію <b>%s</b> ❌", escapedFullName, escapedTitle)
	}

	MessagePublisherReturnedTerritory = func(fullName string, territoryTitle string) string {
		escapedFullName := html.EscapeString(fullName)
		escapedTitle := html.EscapeString(territoryTitle)
		return fmt.Sprintf("Вісник <b>%s</b> повернув територію <b>%s</b> ✅", escapedFullName, escapedTitle)
	}
	MessageEditTerritoryNote = func(territoryTitle string, currentNote string) string {
		escapedTitle := html.EscapeString(territoryTitle)
		message := fmt.Sprintf("📝 <b>Редагування нотатки для території %s</b>\n\n", escapedTitle)
		if currentNote != "" {
			escapedNote := html.EscapeString(currentNote)
			message += "Поточна нотатка (натисніть, щоб скопіювати):\n<code>" + escapedNote + "</code>\n\n"
		}
		message += "Надішліть нову нотатку  ✍️"
		return message
	}

	MessageTerritoryNotInUse       = "Територія не використовується 🤷"
	MessageTerritoryCannotEditNote = "Ви не можете редагувати нотатку для цієї території 🤷"
	MessageTerritoryNoteSaved      = "Нотатку збережено ✅"
	MessageTerritoryNoteDeleted    = "Нотатку видалено ✅"

	MessageTerritoryReturned = "Територію повернуто ✅"

	MessagePublisherNotFound = "Вісника не знайдено 🤷"
)

type MessageNewJoinRequestOptions struct {
	FirstName string
	LastName  string
	Username  string
}

type MessageTerritoryListTerritoryCaptionOptions struct {
	UserRole        entity.UserRole
	Title           string
	LastTakenAt     time.Time
	Note            string
	InUseByFullName string
}
