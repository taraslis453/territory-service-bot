package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
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

// NOTE: using short names because of telegram button query limit of 64 bytes
const approvePublisherJoinRequestButtonUnique = "ap"
const rejectPublisherJoinRequestButtonUnique = "rp"
const territoryGroupButtonUnique = "-tg"
const takeTerritoryButtonUnique = "-tt"
const returnTerritoryButtonUnique = "-rt"
const approveTerritoryTakeButtonUnique = "-att"
const rejectTerritoryTakeButtonUnique = "-rtt"
const leaveTerritoryNoteButtonUnique = "-ltn"

const messengerIDContextKey = "messengerID"

func (s *botService) HandleStart(c tb.Context, b *tb.Bot) error {
	logger := s.logger.
		Named("HandleStart")

	user, err := s.storages.User.GetUser(&GetUserFilter{
		MessengerUserID: fmt.Sprint(c.Sender().ID),
	})
	if err != nil {
		logger.Error("failed to get user by telegram id", "error", err)
		return err
	}
	if user == nil {
		logger.Info("user not found")

		_, err := s.storages.User.CreateUser(&entity.User{
			MessengerUserID:    fmt.Sprint(c.Sender().ID),
			MessengerChatID:    fmt.Sprint(c.Chat().ID),
			Stage:              entity.UserPublisherStageEnterFullName,
			JoinCongregationID: c.Message().Payload,
		})
		if err != nil {
			logger.Error("failed to create user", "error", err)
			return err
		}

		return c.Send(MessageEnterFullName)
	}
	if user.FullName == "" && user.Stage == entity.UserPublisherStageEnterFullName {
		logger.Info("user full name not set")
		return c.Send(MessageEnterFullName)
	}
	if user.Stage == entity.UserPublisherStageEnterCongregationName {
		logger.Info("user waiting for admin approval")
		return c.Send(MessageWaitingForAdminApproval)
	}

	user.Stage = entity.UserStageSelectActionFromMenu
	_, err = s.storages.User.UpdateUser(user)
	if err != nil {
		logger.Error("failed to update user", "error", err)
		return err
	}

	c.Set(messengerIDContextKey, user.MessengerChatID)
	return s.RenderMenu(c, b)
}

func (s *botService) HandleMessage(c tb.Context, b *tb.Bot) error {
	logger := s.logger.
		Named("HandleMessage")

	user, err := s.storages.User.GetUser(&GetUserFilter{
		MessengerUserID: fmt.Sprint(c.Sender().ID),
	})
	if err != nil {
		logger.Error("failed to get user by telegram id", "error", err)
		return err
	}
	if user == nil {
		logger.Info("user not found")
		_, err := s.storages.User.CreateUser(&entity.User{
			MessengerUserID: fmt.Sprint(c.Sender().ID),
			MessengerChatID: fmt.Sprint(c.Chat().ID),
			Stage:           entity.UserPublisherStageEnterCongregationName,
		})
		if err != nil {
			logger.Error("failed to create user", "error", err)
			return err
		}

		return c.Send(MessageEnterCongregationName)
	}
	logger.Info("user found")

	switch c.Message().Text {
	case entity.ViewTerritoryListButton:
		return s.handleViewTerritoryGroupList(c, user)
	case entity.ViewMyTerritoryListButton:
		return s.handleViewMyTerritoryList(c, user)
	case entity.AddTerritoryButton:
		return s.handleAddTerritory(c, user)
	}

	switch user.Stage {
	case entity.UserPublisherStageEnterFullName:
		return s.handlePublisherFullName(c, b, user)
	case entity.UserPublisherStageEnterCongregationName:
		return s.handleCongregationPublisherJoinRequest(c, b, handleCongregationPublisherJoinRequestOptions{
			User:             user,
			CongregationName: c.Message().Text,
		})
	case entity.UserPublisherStageWaitingForAdminApproval:
		return c.Send(MessageWaitingForAdminApproval)
	case entity.UserAdminStageSendTerritory:
		return s.sendAddTerritoryInstruction(c)
	case entity.UserStageLeaveTerritoryNote:
		territoryID := strings.Replace(c.Message().ReplyTo.Entities[0].URL, "tg://btn/", "", -1)
		return s.handleLeaveTerritoryNoteMessage(c, user, territoryID, c.Message().Text)
	default:
		c.Set(messengerIDContextKey, user.MessengerChatID)
		return s.RenderMenu(c, b)
	}
}

func (s *botService) handlePublisherFullName(c tb.Context, b *tb.Bot, user *entity.User) error {
	logger := s.logger.
		Named("handlePublisherFullName").
		With("fullName", c.Message().Text)

	user.FullName = c.Message().Text
	if user.JoinCongregationID != "" {
		return s.handleCongregationPublisherJoinRequest(c, b, handleCongregationPublisherJoinRequestOptions{
			User:           user,
			CongregationID: user.JoinCongregationID,
		})
	}
	user.Stage = entity.UserPublisherStageEnterCongregationName
	_, err := s.storages.User.UpdateUser(user)
	if err != nil {
		logger.Error("failed to update user", "error", err)
		return err
	}

	return c.Send(MessageEnterCongregationName)
}

type handleCongregationPublisherJoinRequestOptions struct {
	User             *entity.User
	CongregationName string
	CongregationID   string
}

func (s *botService) handleCongregationPublisherJoinRequest(c tb.Context, b *tb.Bot, options handleCongregationPublisherJoinRequestOptions) error {
	logger := s.logger.
		Named("handleCongregationPublisherJoinRequest").
		With("congregationName", c.Message().Text)

	congregation, err := s.storages.Congregation.GetCongregation(&GetCongregationFilter{
		Name: options.CongregationName,
		ID:   options.CongregationID,
	})
	if err != nil {
		logger.Error("failed to get congregation by name", "error", err)
		return err
	}
	if congregation == nil {
		logger.Info("congregation not found")
		return c.Send(MessageCongregationNotFound)
	}

	admins, err := s.storages.User.ListUsers(&ListUsersFilter{
		CongregationID: congregation.ID,
		Role:           entity.UserRoleAdmin,
	})
	if err != nil {
		logger.Error("failed to get admin user by congregation id", "err", err)
		return err
	}
	if len(admins) == 0 {
		logger.Info("admins not found")
		return c.Send(MessageCongregationAdminNotFound)
	}

	requestActionStateID := uuid.New().String()
	var messages []entity.AdminMessage
	for _, admin := range admins {
		// NOTE using this because we can't pass more than 64 bytes in callback data
		// https://github.com/nmlorg/metabot/issues/1
		message := fmt.Sprintf("<a href=\"tg://btn/%s/%s\">\u200b</a> %s", options.User.ID, requestActionStateID, MessageNewJoinRequest(&MessageNewJoinRequestOptions{
			FirstName: c.Sender().FirstName,
			LastName:  c.Sender().LastName,
			Username:  c.Sender().Username,
		}))
		sentMessage, err := b.Send(&recepient{chatID: admin.MessengerChatID}, message, &tb.ReplyMarkup{
			InlineKeyboard: [][]tb.InlineButton{
				{
					tb.InlineButton{
						Unique: approvePublisherJoinRequestButtonUnique,
						Text:   entity.ApprovePublisherButton,
					},
					tb.InlineButton{
						Unique: rejectPublisherJoinRequestButtonUnique,
						Text:   entity.RejectPublisherButton,
					},
				},
			},
		}, tb.ModeHTML)
		if err != nil {
			logger.Error("failed to send message to admin", "err", err)
			return err
		}
		messages = append(messages, entity.AdminMessage{
			MessageID: fmt.Sprint(sentMessage.ID),
			ChatID:    fmt.Sprint(sentMessage.Chat.ID),
		})
	}
	logger = logger.With("messages", messages)

	createdActionState, err := s.storages.Chat.CreateRequestActionState(&entity.RequestActionState{
		ID:            requestActionStateID,
		AdminMessages: messages,
	})
	if err != nil {
		logger.Error("failed to create request action state", "err", err)
		return err
	}
	logger = logger.With("createdActionState", createdActionState)

	options.User.Stage = entity.UserPublisherStageWaitingForAdminApproval
	_, err = s.storages.User.UpdateUser(options.User)
	if err != nil {
		logger.Error("failed to update user", "err", err)
		return err
	}

	return c.Send(MessageCongregationJoinRequestSent(congregation.Name), &tb.ReplyMarkup{}, tb.ModeMarkdown)
}

type recepient struct {
	chatID string
}

func (r *recepient) Recipient() string {
	return r.chatID
}

func (s *botService) RenderMenu(c tb.Context, b *tb.Bot) error {
	logger := s.logger.
		Named("RenderMenu")

	messengerID := fmt.Sprint(c.Sender().ID)

	if c.Get(messengerIDContextKey) != nil {
		messengerID = fmt.Sprint(c.Get(messengerIDContextKey))
	}
	logger = logger.With("messengerID", messengerID)

	user, err := s.storages.User.GetUser(&GetUserFilter{
		MessengerUserID: messengerID,
	})
	if err != nil {
		logger.Error("failed to get user by telegram id", "err", err)
		return err
	}
	if user == nil {
		logger.Info("user not found")
		return c.Send(MessageUserNotFound)
	}
	if user.Role == "" {
		logger.Info("user not joined to congregation")
		return c.Send(MessageEnterCongregationName)
	}

	buttons := [][]tb.ReplyButton{
		{tb.ReplyButton{Text: entity.ViewTerritoryListButton}},
		{tb.ReplyButton{Text: entity.ViewMyTerritoryListButton}},
	}
	if user.Role == entity.UserRoleAdmin {
		buttons = append(buttons, []tb.ReplyButton{
			{Text: entity.AddTerritoryButton},
		})
	}
	logger = logger.With("buttons", buttons)
	logger.Info("successfully rendered menu buttons")

	_, err = b.Send(&recepient{chatID: messengerID}, MessageHowCanIHelpYou, &tb.SendOptions{
		ReplyMarkup: &tb.ReplyMarkup{
			ReplyKeyboard: buttons,
		},
	},
	)
	if err != nil {
		logger.Error("failed to send menu", "err", err)
		return err
	}

	return nil
}

func (s *botService) HandleInlineButton(c tb.Context, b *tb.Bot) error {
	logger := s.logger.
		Named("HandleInlineButton")

	var err error
	defer func() {
		err = c.Respond() // respond to the callback to remove the loading state
	}()
	if err != nil {
		return err
	}

	user, err := s.storages.User.GetUser(&GetUserFilter{
		MessengerUserID: fmt.Sprint(c.Sender().ID),
	})
	if err != nil {
		logger.Error("failed to get user by messenger user id", "err", err)
		return err
	}
	if user == nil {
		logger.Info("user not found")
		return c.Send(MessageUserNotFound)
	}

	data := c.Data()
	data = strings.Replace(data, "\f", "", -1)

	switch {
	case strings.Contains(data, approvePublisherJoinRequestButtonUnique):
		publisherID := strings.Split(strings.Replace(c.Message().Entities[0].URL, "tg://btn/", "", -1), "/")[0]
		requestActionStateID := strings.Replace(c.Message().Entities[0].URL, fmt.Sprintf("tg://btn/%s/", publisherID), "", -1)
		return s.handleApprovePublisherJoinRequest(c, b, user, publisherID, requestActionStateID)
	case strings.Contains(data, rejectPublisherJoinRequestButtonUnique):
		publisherID := strings.Split(strings.Replace(c.Message().Entities[0].URL, "tg://btn/", "", -1), "/")[0]
		requestActionStateID := strings.Replace(c.Message().Entities[0].URL, fmt.Sprintf("tg://btn/%s/", publisherID), "", -1)
		return s.handleRejectPublisherJoinRequest(c, b, user, publisherID, requestActionStateID)
	case strings.Contains(data, territoryGroupButtonUnique):
		groupName := strings.Replace(data, territoryGroupButtonUnique, "", -1)
		return s.handleViewTerritoriesList(c, user, groupName)
	case strings.Contains(data, takeTerritoryButtonUnique):
		territoryID := strings.Replace(data, takeTerritoryButtonUnique, "", -1)
		return s.handleTakeTerritoryRequest(c, b, user, territoryID)
	case strings.Contains(data, approveTerritoryTakeButtonUnique):
		parts := strings.Split(strings.TrimPrefix(c.Message().CaptionEntities[0].URL, "tg://btn/"), "/")
		return s.handleApproveTerritoryTakeRequest(c, b, parts[0], parts[1], parts[2])
	case strings.Contains(data, rejectTerritoryTakeButtonUnique):
		parts := strings.Split(strings.TrimPrefix(c.Message().CaptionEntities[0].URL, "tg://btn/"), "/")
		return s.handleRejectTerritoryTakeRequest(c, b, parts[0], parts[1], parts[2])
	case strings.Contains(data, leaveTerritoryNoteButtonUnique):
		territoryID := strings.Replace(data, leaveTerritoryNoteButtonUnique, "", -1)
		return s.handleLeaveTerritoryNoteRequest(c, user, territoryID)
	case strings.Contains(data, returnTerritoryButtonUnique):
		territoryID := strings.Replace(data, returnTerritoryButtonUnique, "", -1)
		return s.handleReturnTerritoryRequest(c, b, user, territoryID)
	default:
		return fmt.Errorf("unknown button: %s", data)
	}
}

func (s *botService) handleAddTerritory(c tb.Context, user *entity.User) error {
	logger := s.logger.
		Named("handleAddTerritory")

	if user.Role != entity.UserRoleAdmin {
		logger.Info("user is not admin")
		return c.Send(MessageUserIsNotAdmin)
	}

	user.Stage = entity.UserAdminStageSendTerritory
	_, err := s.storages.User.UpdateUser(user)
	if err != nil {
		logger.Error("failed to update user", "err", err)
		return err
	}

	return s.sendAddTerritoryInstruction(c)
}

func (s *botService) handleViewTerritoryGroupList(c tb.Context, user *entity.User) error {
	logger := s.logger.
		Named("handleViewTerritoryGroupList")

	var showAvailableTerritories *bool
	if user.Role != entity.UserRoleAdmin {
		logger.Info("user is not admin")
		showAvailableTerritories = &[]bool{true}[0]
	}

	territories, err := s.storages.Congregation.ListTerritories(&ListTerritoriesFilter{
		CongregationID: user.CongregationID,
		Available:      showAvailableTerritories,
	})
	if err != nil {
		logger.Error("failed to list territories", "err", err)
		return err
	}
	if len(territories) == 0 {
		logger.Info("no territories found")
		return c.Send(MessageNoTerritoriesFound)
	}

	var groupIDs []string
	for _, territory := range territories {
		groupIDs = append(groupIDs, territory.GroupID)
	}

	groups, err := s.storages.Congregation.ListTerritoryGroups(&ListTerritoryGroupsFilter{
		IDs: groupIDs,
	})
	if err != nil {
		logger.Error("failed to list groups", "err", err)
		return err
	}
	if len(groups) == 0 {
		logger.Error("no groups found")
		return nil
	}

	groupIDTitles := make(map[string]string)
	for _, group := range groups {
		groupIDTitles[group.ID] = group.Title
	}

	countAvailableTerritoriesInGroups := make(map[string]int)
	for _, territory := range territories {
		groupTitle, ok := groupIDTitles[territory.GroupID]
		if !ok {
			logger.Warn("group title not found", "group_id", territory.GroupID)
			continue
		}
		countAvailableTerritoriesInGroups[groupTitle]++
	}

	var buttons [][]tb.InlineButton
	for groupName, territoriesCount := range countAvailableTerritoriesInGroups {
		buttons = append(buttons, []tb.InlineButton{
			{
				Unique: groupName + territoryGroupButtonUnique,
				Text:   groupName + " (" + strconv.Itoa(territoriesCount) + ")",
			},
		})
	}

	return c.Send(MessageTerritoryList, &tb.SendOptions{
		ReplyMarkup: &tb.ReplyMarkup{
			InlineKeyboard: buttons,
		},
	}, tb.ModeMarkdown)
}

func (s *botService) handleViewMyTerritoryList(c tb.Context, user *entity.User) error {
	logger := s.logger.
		Named("handleViewMyTerritoryList").
		With("user", user)

	territories, err := s.storages.Congregation.ListTerritories(&ListTerritoriesFilter{
		CongregationID: user.CongregationID,
		InUseByUserID:  user.ID,
	})
	if err != nil {
		logger.Error("failed to list territories", "err", err)
		return err
	}
	if len(territories) == 0 {
		logger.Info("no territories found")
		return c.Send(MessageNoTerritoriesFound)
	}

	for _, territory := range territories {
		var notes []string
		for _, note := range territory.Notes {
			notes = append(notes, note.Text)
		}
		caption := MessageMyTerritoryListTerritoryCaption(territory.Title, territory.LastTakenAt, notes)

		var sendObject interface{}
		if territory.FileType == entity.CongregationTerritoryFileTypePhoto {
			sendObject = &tb.Photo{File: tb.File{
				FileID: territory.FileID,
			},
				Caption: caption,
			}
		} else if territory.FileType == entity.CongregationTerritoryFileTypeDocument {
			sendObject = &tb.Document{File: tb.File{
				FileID: territory.FileID,
			},
				Caption: caption,
			}
		} else {
			logger.Error("unknown file type", "file_type", territory.FileType)
			continue
		}

		err := c.Send(sendObject, &tb.SendOptions{
			ReplyMarkup: &tb.ReplyMarkup{
				InlineKeyboard: [][]tb.InlineButton{
					{
						{
							Unique: territory.ID + leaveTerritoryNoteButtonUnique,
							Text:   entity.LeaveTerritoryNoteButton,
						},
					},
					{
						{
							Unique: territory.ID + returnTerritoryButtonUnique,
							Text:   entity.ReturnTerritoryButton,
						},
					},
				},
			},
		}, tb.ModeMarkdown)
		if err != nil {
			logger.Error("failed to send photo", "err", err)
			return err
		}
	}

	return nil
}

func (s *botService) sendAddTerritoryInstruction(c tb.Context) error {
	return c.Send(MessageAddTerritoryInstruction, &tb.SendOptions{}, tb.ModeMarkdown)
}

func (s *botService) handleLeaveTerritoryNoteRequest(c tb.Context, user *entity.User, territoryID string) error {
	logger := s.logger.
		Named("handleLeaveTerritoryNoteRequest").
		With("territoryID", territoryID)

	territory, err := s.storages.Congregation.GetTerritory(&GetTerritoryFilter{
		ID: territoryID,
	})
	if err != nil {
		logger.Error("failed to get territory", "err", err)
		return err
	}
	if territory == nil {
		logger.Info("territory not found")
		return c.Send(MessageTerritoryNotFound)
	}

	user.Stage = entity.UserStageLeaveTerritoryNote
	_, err = s.storages.User.UpdateUser(user)
	if err != nil {
		logger.Error("failed to update user", "err", err)
		return err
	}

	message := fmt.Sprintf("<a href=\"tg://btn/%s\">\u200b</a> %s", territory.ID, MessageLeaveTerritoryNote(territory.Title))
	return c.Send(message, &tb.SendOptions{
		ReplyMarkup: &tb.ReplyMarkup{
			ForceReply: true,
		},
	}, tb.ModeHTML)
}

func (s *botService) handleLeaveTerritoryNoteMessage(c tb.Context, user *entity.User, territoryID, note string) error {
	logger := s.logger.
		Named("handleLeaveTerritoryNoteMessage").
		With("user", user, "territoryID", territoryID, "note", note)

	territory, err := s.storages.Congregation.GetTerritory(&GetTerritoryFilter{
		ID: territoryID,
	})
	if err != nil {
		logger.Error("failed to get territory", "err", err)
		return err
	}
	if territory == nil {
		logger.Info("territory not found")
		return c.Send(MessageTerritoryNotFound)
	}

	if territory.InUseByUserID == nil {
		logger.Info("territory not in use")
		return c.Send(MessageTerritoryNotInUse)
	}
	if *territory.InUseByUserID != user.ID {
		logger.Info("territory not in use by user")
		return c.Send(MessageTerritoryCannotLeaveNote)
	}

	_, err = s.storages.Congregation.AddTerritoryNote(&entity.CongregationTerritoryNote{
		TerritoryID: territory.ID,
		UserID:      user.ID,
		Text:        note,
	})
	if err != nil {
		logger.Error("failed to add territory note", "err", err)
		return err
	}

	user.Stage = entity.UserStageSelectActionFromMenu
	_, err = s.storages.User.UpdateUser(user)
	if err != nil {
		logger.Error("failed to update user", "err", err)
		return err
	}

	return c.Send(MessageTerritoryNoteSaved, tb.ModeMarkdown)
}

func (s botService) handleReturnTerritoryRequest(c tb.Context, b *tb.Bot, user *entity.User, territoryID string) error {
	logger := s.logger.
		Named("handleReturnTerritoryRequest").
		With(user, "user", territoryID, territoryID)

	territory, err := s.storages.Congregation.GetTerritory(&GetTerritoryFilter{
		ID: territoryID,
	})
	if err != nil {
		logger.Error("failed to get territory", "err", err)
		return err
	}
	if territory == nil {
		logger.Info("territory not found")
		return c.Send(MessageTerritoryNotFound)
	}

	territory.InUseByUserID = nil
	territory.LastTakenAt = time.Now()

	_, err = s.storages.Congregation.UpdateTerritory(territory)
	if err != nil {
		logger.Error("failed to update territory", "err", err)
		return err
	}

	if user.Role == entity.UserRolePublisher {
		admins, err := s.storages.User.ListUsers(&ListUsersFilter{
			CongregationID: user.CongregationID,
			Role:           entity.UserRoleAdmin,
		})
		if err != nil {
			logger.Error("failed to get admin", "err", err)
			return err
		}
		if len(admins) == 0 {
			logger.Info("admin not found")
			return c.Send(MessageCongregationAdminNotFound)
		}

		for _, admin := range admins {
			_, err = b.Send(&recepient{
				chatID: admin.MessengerChatID,
			}, MessagePublisherReturnedTerritory(user.FullName, territory.Title), tb.ModeMarkdown)
			if err != nil {
				logger.Error("failed to send message", "err", err)
				return err
			}
		}
	}

	message := c.Message()
	message.ReplyMarkup = nil
	message.Caption = message.Caption + "\n\n" + MessageTerritoryReturned
	_, err = b.EditCaption(&editable{
		chatID:    message.Chat.ID,
		messageID: fmt.Sprintf("%d", message.ID),
	}, MessageTerritoryReturned)
	if err != nil {
		logger.Error("failed to edit message", "err", err)
		return err
	}

	return nil
}

func (s *botService) handleApprovePublisherJoinRequest(c tb.Context, b *tb.Bot, admin *entity.User, publisherID string, requestActionStateID string) error {
	logger := s.logger.
		Named("handleApprovePublisherJoinRequest").
		With("publisherID", publisherID)

	publisher, err := s.storages.User.GetUser(&GetUserFilter{
		ID: publisherID,
	})
	if err != nil {
		logger.Error("failed to get publisher", "err", err)
		return err
	}
	if publisher == nil {
		logger.Info("publisher not found")
		return c.Send(MessagePublisherNotFound)
	}

	publisher.CongregationID = admin.CongregationID
	publisher.Stage = entity.UserStageSelectActionFromMenu
	publisher.Role = entity.UserRolePublisher

	_, err = s.storages.User.UpdateUser(publisher)
	if err != nil {
		logger.Error("failed to update publisher", "err", err)
		return err
	}

	_, err = b.Send(&recepient{chatID: publisher.MessengerChatID}, MessageCongregationJoinRequestApproved)
	if err != nil {
		logger.Error("failed to send message to publisher", "err", err)
		return err
	}

	// Render menu for publisher after approving request from admin
	c.Set(messengerIDContextKey, publisher.MessengerUserID)
	err = s.RenderMenu(c, b)
	if err != nil {
		logger.Error("failed to render menu for publisher", "err", err)
	}

	requestActionState, err := s.storages.Chat.GetRequestActionState(requestActionStateID)
	if err != nil {
		logger.Error("failed to get request action state", "err", err)
		return err
	}

	for _, message := range requestActionState.AdminMessages {
		chatID, err := strconv.ParseInt(message.ChatID, 10, 64)
		if err != nil {
			logger.Error("failed to parse chat id", "err", err)
			return err
		}

		_, err = b.Edit(&editable{
			chatID:    chatID,
			messageID: message.MessageID,
		}, MessageCongregationJoinRequestApprovedDone(publisher.FullName), tb.ModeMarkdown)
		if err != nil {
			logger.Error("failed to edit message", "err", err)
			return err
		}
	}

	err = s.storages.Chat.DeleteRequestActionState(requestActionStateID)
	if err != nil {
		logger.Error("failed to delete request action state", "err", err)
		return err
	}

	return nil
}

func (s *botService) handleRejectPublisherJoinRequest(c tb.Context, b *tb.Bot, admin *entity.User, publisherID string, requestActionStateID string) error {
	logger := s.logger.
		Named("handleRejectPublisherJoinRequest").
		With("publisherID", publisherID)

	publisher, err := s.storages.User.GetUser(&GetUserFilter{
		ID: publisherID,
	})
	if err != nil {
		logger.Error("failed to get publisher", "err", err)
		return err
	}
	if publisher == nil {
		logger.Info("publisher not found")
		return c.Send(MessagePublisherNotFound)
	}

	publisher.Stage = entity.UserPublisherStageCongregationJoinRequestRejected

	_, err = s.storages.User.UpdateUser(publisher)
	if err != nil {
		logger.Error("failed to update publisher", "err", err)
		return err
	}

	_, err = b.Send(&recepient{chatID: publisher.MessengerChatID}, MessageCongregationJoinRequestRejected, tb.ModeMarkdown)
	if err != nil {
		logger.Error("failed to send message to publisher", "err", err)
		return err
	}

	requestActionState, err := s.storages.Chat.GetRequestActionState(requestActionStateID)
	if err != nil {
		logger.Error("failed to get request action state", "err", err)
		return err
	}

	for _, message := range requestActionState.AdminMessages {
		chatID, err := strconv.ParseInt(message.ChatID, 10, 64)
		if err != nil {
			logger.Error("failed to parse chat id", "err", err)
			return err
		}
		_, err = b.Edit(&editable{
			chatID:    chatID,
			messageID: message.MessageID,
		}, MessageCongregationJoinRequestRejectedDone(publisher.FullName), tb.ModeMarkdown)
		if err != nil {
			logger.Error("failed to edit message", "err", err)
			return err
		}
	}

	return nil
}

type editable struct {
	chatID    int64
	messageID string
}

func (e *editable) MessageSig() (string, int64) {
	return e.messageID, e.chatID
}

func (s *botService) handleViewTerritoriesList(c tb.Context, user *entity.User, groupName string) error {
	logger := s.logger.
		Named("handleViewTerritoriesList")

	// FIXME: we should not use create method
	group, err := s.storages.Congregation.GetOrCreateCongregationTerritoryGroup(&GetOrCreateCongregationTerritoryGroupOptions{
		CongregationID: user.CongregationID,
		Title:          groupName,
	})
	if err != nil {
		logger.Error("failed to get or create territory group", "err", err)
		return err
	}

	listTerritoriesFilter := &ListTerritoriesFilter{
		CongregationID: user.CongregationID,
		GroupID:        group.ID,
		SortBy:         "last_taken_at asc",
	}

	if user.Role == entity.UserRolePublisher {
		listTerritoriesFilter.Available = &[]bool{true}[0]
	}

	territories, err := s.storages.Congregation.ListTerritories(listTerritoriesFilter)
	if err != nil {
		logger.Error("failed to list territories", "err", err)
		return err
	}

	for _, territory := range territories {
		var sendOptions tb.SendOptions

		var inUseByFullName string
		if territory.InUseByUserID != nil {

			publisher, err := s.storages.User.GetUser(&GetUserFilter{
				ID: *territory.InUseByUserID,
			})
			if err != nil {
				logger.Error("failed to get publisher", "err", err)
				return err
			}
			if publisher != nil {
				inUseByFullName = publisher.FullName
			}
		}

		var notes []string
		if len(territory.Notes) > 0 {
			for _, note := range territory.Notes {
				notes = append(notes, note.Text)
			}
		}

		caption := MessageTerritoryListTerritoryCaption(MessageTerritoryListTerritoryCaptionOptions{
			UserRole:        user.Role,
			Title:           territory.Title,
			LastTakenAt:     territory.LastTakenAt,
			Notes:           notes,
			InUseByFullName: inUseByFullName,
		})

		var sendObject interface{}
		if territory.FileType == entity.CongregationTerritoryFileTypePhoto {
			sendObject = &tb.Photo{File: tb.File{
				FileID: territory.FileID,
			},
				Caption: caption,
			}
		} else if territory.FileType == entity.CongregationTerritoryFileTypeDocument {
			sendObject = &tb.Document{File: tb.File{
				FileID: territory.FileID,
			},
				Caption: caption,
			}
		} else {
			logger.Error("unknown file type", "file_type", territory.FileType)
			continue
		}

		if territory.InUseByUserID == nil {
			button := tb.InlineButton{
				Unique: territory.ID + takeTerritoryButtonUnique,
				Text:   fmt.Sprintf("%s %s", entity.TakeTerritoryButton, territory.Title),
			}

			keyboard := [][]tb.InlineButton{{button}}
			markup := &tb.ReplyMarkup{InlineKeyboard: keyboard}
			sendOptions.ReplyMarkup = markup
		}
		err := c.Send(sendObject, &sendOptions, tb.ModeMarkdown)
		if err != nil {
			logger.Error("failed to send territory", "err", err)
			return err
		}
	}

	return nil
}

func (s *botService) handleTakeTerritoryRequest(c tb.Context, b *tb.Bot, user *entity.User, territoryID string) error {
	logger := s.logger.
		Named("handleTakeTerritoryRequest")

	territory, err := s.storages.Congregation.GetTerritory(&GetTerritoryFilter{
		ID: territoryID,
	})
	if err != nil {
		logger.Error("failed to get territory", "err", err)
		return err
	}
	if territory == nil {
		logger.Info("territory not found")
		return c.Send(MessageTerritoryNotFound)
	}
	if territory.InUseByUserID != nil {
		logger.Info("territory is not available")
		return c.Send(MessageTerritoryNotAvailable)
	}

	admins, err := s.storages.User.ListUsers(&ListUsersFilter{
		CongregationID: user.CongregationID,
		Role:           entity.UserRoleAdmin,
	})
	if err != nil {
		logger.Error("failed to get admin user by congregation id", "err", err)
		return err
	}
	if len(admins) == 0 {
		logger.Info("admin user not found")
		return c.Send(MessageCongregationAdminNotFound)
	}

	requestActionStateID := uuid.New().String()
	var messages []entity.AdminMessage

	for _, admin := range admins {
		message := fmt.Sprintf("<a href=\"tg://btn/%s/%s/%s\">\u200b</a> %s", user.ID, territoryID, requestActionStateID, MessageTakeTerritoryRequest(user, territory.Title))
		var sendObject interface{}
		if territory.FileType == entity.CongregationTerritoryFileTypePhoto {
			sendObject = &tb.Photo{File: tb.File{
				FileID: territory.FileID,
			},
				Caption: message,
			}
		} else if territory.FileType == entity.CongregationTerritoryFileTypeDocument {
			sendObject = &tb.Document{File: tb.File{
				FileID: territory.FileID,
			},
				Caption: message,
			}
		} else {
			logger.Error("unknown file type", "file_type", territory.FileType)
			continue
		}
		sentMessage, err := b.Send(&recepient{chatID: admin.MessengerChatID},
			sendObject,
			&tb.ReplyMarkup{
				InlineKeyboard: [][]tb.InlineButton{
					{
						tb.InlineButton{
							Unique: approveTerritoryTakeButtonUnique,
							Text:   entity.ApproveTakeTerritoryButton,
						},
						tb.InlineButton{
							Unique: rejectTerritoryTakeButtonUnique,
							Text:   entity.RejectTakeTerritoryButton,
						},
					},
				},
			}, tb.ModeHTML)
		if err != nil {
			logger.Error("failed to send message to admin", "err", err)
			return err
		}

		messages = append(messages, entity.AdminMessage{
			MessageID: fmt.Sprint(sentMessage.ID),
			ChatID:    fmt.Sprint(sentMessage.Chat.ID),
		})
	}
	logger = logger.With("messages", messages)

	createdActionState, err := s.storages.Chat.CreateRequestActionState(&entity.RequestActionState{
		ID:            requestActionStateID,
		AdminMessages: messages,
	})
	if err != nil {
		logger.Error("failed to create request action state", "err", err)
		return err
	}
	logger = logger.With("createdActionState", createdActionState)

	messageID := c.Callback().Message.ID
	_, err = b.EditCaption(&editable{
		chatID:    c.Callback().Message.Chat.ID,
		messageID: fmt.Sprintf("%d", messageID),
	}, MessageTakeTerritoryRequestSent, tb.ModeMarkdown)
	if err != nil {
		logger.Error("failed to edit message", "err", err)
		return err
	}

	return nil
}

func (s *botService) handleApproveTerritoryTakeRequest(c tb.Context, b *tb.Bot, publisherID string, territoryID string, requestActionStateID string) error {
	logger := s.logger.
		Named("handleApproveTerritoryTakeRequest").
		With("publisherID", publisherID, "territoryID", territoryID, "requestActionStateID", requestActionStateID)

	publisher, err := s.storages.User.GetUser(&GetUserFilter{
		ID: publisherID,
	})
	if err != nil {
		logger.Error("failed to get publisher user", "err", err)
		return err
	}
	if publisher == nil {
		logger.Info("publisher user not found")
		return c.Send(MessagePublisherNotFound)
	}

	territory, err := s.storages.Congregation.GetTerritory(&GetTerritoryFilter{
		ID: territoryID,
	})
	if err != nil {
		logger.Error("failed to get territory", "err", err)
		return err
	}
	if territory == nil {
		logger.Info("territory not found")
		return c.Send(MessageTerritoryNotFound)
	}

	if territory.InUseByUserID != nil {
		logger.Info("territory is not available")
		return c.Send(MessageTerritoryNotAvailable)
	}

	territory.InUseByUserID = &publisherID
	territory.LastTakenAt = time.Now()
	territory, err = s.storages.Congregation.UpdateTerritory(territory)
	if err != nil {
		logger.Error("failed to update territory", "err", err)
		return err
	}

	var notes []string
	for _, note := range territory.Notes {
		notes = append(notes, note.Text)
	}

	message := MessageTakeTerritoryRequestApproved(territory.Title, notes)
	_, err = b.Send(&recepient{chatID: publisher.MessengerChatID}, message, tb.ModeMarkdown)
	if err != nil {
		logger.Error("failed to send message to user", "err", err)
		return err
	}

	requestActionState, err := s.storages.Chat.GetRequestActionState(requestActionStateID)
	if err != nil {
		logger.Error("failed to get request action state", "err", err)
		return err
	}

	for _, message := range requestActionState.AdminMessages {
		chatID, err := strconv.ParseInt(message.ChatID, 10, 64)
		if err != nil {
			logger.Error("failed to parse chat id", "err", err)
			return err
		}

		_, err = b.EditCaption(&editable{
			chatID:    chatID,
			messageID: message.MessageID,
		}, MessageTakeTerritoryRequestApprovedDone(publisher.FullName, territory.Title), tb.ModeMarkdown)
		if err != nil {
			logger.Error("failed to edit message", "err", err)
			return err
		}
	}

	err = s.storages.Chat.DeleteRequestActionState(requestActionStateID)
	if err != nil {
		logger.Error("failed to delete request action state", "err", err)
		return err
	}

	return nil
}

func (s *botService) handleRejectTerritoryTakeRequest(c tb.Context, b *tb.Bot, publisherID string, territoryID string, requestActionStateID string) error {
	logger := s.logger.
		Named("handleRejectTerritoryTakeRequest").
		With("publisherID", publisherID, "territoryID", territoryID, "requestActionStateID", requestActionStateID)

	publisher, err := s.storages.User.GetUser(&GetUserFilter{
		ID: publisherID,
	})
	if err != nil {
		logger.Error("failed to get publisher user", "err", err)
		return err
	}
	if publisher == nil {
		logger.Info("publisher user not found")
		return c.Send(MessagePublisherNotFound)
	}

	territory, err := s.storages.Congregation.GetTerritory(&GetTerritoryFilter{
		ID: territoryID,
	})
	if err != nil {
		logger.Error("failed to get territory", "err", err)
		return err
	}
	if territory == nil {
		logger.Info("territory not found")
		return c.Send(MessageTerritoryNotFound)
	}

	message := MessageTakeTerritoryRequestRejected(territory.Title)
	_, err = b.Send(&recepient{chatID: publisher.MessengerChatID}, message, tb.ModeMarkdown)
	if err != nil {
		logger.Error("failed to send message to user", "err", err)
		return err
	}

	requestActionState, err := s.storages.Chat.GetRequestActionState(requestActionStateID)
	if err != nil {
		logger.Error("failed to get request action state", "err", err)
		return err
	}

	for _, message := range requestActionState.AdminMessages {
		chatID, err := strconv.ParseInt(message.ChatID, 10, 64)
		if err != nil {
			logger.Error("failed to parse chat id", "err", err)
			return err
		}

		_, err = b.EditCaption(&editable{
			chatID:    chatID,
			messageID: message.MessageID,
		}, MessageTakeTerritoryRequestRejectedDone(publisher.FullName, territory.Title), tb.ModeMarkdown)
		if err != nil {
			logger.Error("failed to edit message", "err", err)
			return err
		}
	}

	err = s.storages.Chat.DeleteRequestActionState(requestActionStateID)
	if err != nil {
		logger.Error("failed to delete request action state", "err", err)
		return err
	}

	return nil
}

func (s *botService) HandleImageUpload(c tb.Context, b *tb.Bot) error {
	logger := s.logger.
		Named("HandleImageUpload")

	user, err := s.storages.User.GetUser(&GetUserFilter{
		MessengerUserID: fmt.Sprint(c.Sender().ID),
	})
	if err != nil {
		logger.Error("failed to get user by messenger user id", "err", err)
		return err
	}
	if user == nil {
		logger.Info("user not found")
		return c.Send(MessageUserNotFound)
	}
	if user.Role != entity.UserRoleAdmin {
		logger.Info("user is not admin")
		return nil
	}

	congregation, err := s.storages.Congregation.GetCongregation(&GetCongregationFilter{
		ID: user.CongregationID,
	})
	if err != nil {
		logger.Error("failed to get congregation by messenger user id", "err", err)
		return err
	}
	if congregation == nil {
		logger.Info("congregation not found")
		return c.Send(MessageCongregationNotFound)
	}

	msg := c.Message()
	fileID := msg.Photo.FileID
	caption := msg.Caption
	if caption == "" || !strings.Contains(caption, "_") {
		return s.sendAddTerritoryInstruction(c)
	}

	split := strings.Split(caption, "_") // Klevan_123-а
	groupName := split[0]                // Klevan
	territoryName := split[1]            // 123-а

	group, err := s.storages.Congregation.GetOrCreateCongregationTerritoryGroup(&GetOrCreateCongregationTerritoryGroupOptions{
		CongregationID: congregation.ID,
		Title:          groupName,
	})
	if err != nil {
		logger.Error("failed to create or get congregation territory group", "err", err)
		return err
	}

	territory, err := s.storages.Congregation.GetTerritory(&GetTerritoryFilter{
		CongregationID: congregation.ID,
		Title:          territoryName,
		GroupID:        group.ID,
	})
	if err != nil {
		logger.Error("failed to get territory", "err", err)
		return err
	}
	if territory != nil {
		logger.Info("territory already exists")
		return c.Send(MessageTerritoryExistsInGroup(territoryName, groupName), &tb.SendOptions{}, tb.ModeMarkdown)
	}

	territory, err = s.storages.Congregation.CreateTerritory(&entity.CongregationTerritory{
		CongregationID: congregation.ID,
		GroupID:        group.ID,
		Title:          territoryName,
		FileID:         fileID,
		FileType:       entity.CongregationTerritoryFileTypePhoto,
	})
	if err != nil {
		logger.Error("failed to create territory", "err", err)
		return err
	}
	logger.With("territory", territory)

	logger.Info("successfully handled image upload")
	return c.Send(fmt.Sprintf("Територія %s успішно додана в групу %s!", territoryName, groupName), &tb.SendOptions{}, tb.ModeMarkdown)
}

func (s *botService) HandleDocumentUpload(c tb.Context, b *tb.Bot) error {
	logger := s.logger.
		Named("HandleDocumentUpload")

	user, err := s.storages.User.GetUser(&GetUserFilter{
		MessengerUserID: fmt.Sprint(c.Sender().ID),
	})
	if err != nil {
		logger.Error("failed to get user by messenger user id", "err", err)
		return err
	}
	if user == nil {
		logger.Info("user not found")
		return c.Send(MessageUserNotFound)
	}
	if user.Role != entity.UserRoleAdmin {
		logger.Info("user is not admin")
		return nil
	}

	congregation, err := s.storages.Congregation.GetCongregation(&GetCongregationFilter{
		ID: user.CongregationID,
	})
	if err != nil {
		logger.Error("failed to get congregation by messenger user id", "err", err)
		return err
	}
	if congregation == nil {
		logger.Info("congregation not found")
		return c.Send(MessageCongregationNotFound)
	}

	msg := c.Message()
	fileID := msg.Document.FileID
	caption := msg.Caption
	if caption == "" || !strings.Contains(caption, "_") {
		return s.sendAddTerritoryInstruction(c)
	}

	split := strings.Split(caption, "_") // Klevan_123-а
	groupName := split[0]                // Klevan
	territoryName := split[1]            // 123-а

	group, err := s.storages.Congregation.GetOrCreateCongregationTerritoryGroup(&GetOrCreateCongregationTerritoryGroupOptions{
		CongregationID: congregation.ID,
		Title:          groupName,
	})
	if err != nil {
		logger.Error("failed to create or get congregation territory group", "err", err)
		return err
	}

	territory, err := s.storages.Congregation.GetTerritory(&GetTerritoryFilter{
		CongregationID: congregation.ID,
		Title:          territoryName,
		GroupID:        group.ID,
	})
	if err != nil {
		logger.Error("failed to get territory", "err", err)
		return err
	}
	if territory != nil {
		logger.Info("territory already exists")
		return c.Send(MessageTerritoryExistsInGroup(territoryName, groupName), &tb.SendOptions{}, tb.ModeMarkdown)
	}

	territory, err = s.storages.Congregation.CreateTerritory(&entity.CongregationTerritory{
		CongregationID: congregation.ID,
		GroupID:        group.ID,
		Title:          territoryName,
		FileID:         fileID,
		FileType:       entity.CongregationTerritoryFileTypeDocument,
	})
	if err != nil {
		logger.Error("failed to create territory", "err", err)
		return err
	}
	logger.With("territory", territory)

	logger.Info("successfully handled image upload")
	return c.Send(fmt.Sprintf("Територія %s успішно додана в групу %s!", territoryName, groupName), &tb.SendOptions{}, tb.ModeMarkdown)
}
