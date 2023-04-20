package service

import (
	"fmt"
	"strconv"
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

// NOTE: using short names because of telegram button query limit of 64 bytes
const approvePublisherJoinRequestButtonUnique = "ap"
const rejectPublisherJoinRequestButtonUnique = "rp"
const territoryGroupButtonUnique = "-tg"
const takeTerritoryButtonUnique = "-tt"
const approveTerritoryTakeButtonPrefix = "-att"

func (s *botService) HandleStart(c tb.Context) error {
	logger := s.logger.
		Named("HandleStart").
		With(c)

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
			Stage:           entity.UserPublisherStageEnterFullName,
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

	return s.RenderMenu(c)
}

func (s *botService) HandleMessage(c tb.Context, b *tb.Bot) error {
	logger := s.logger.
		Named("HandleMessage").
		With(c)

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

	switch user.Stage {
	case entity.UserPublisherStageEnterFullName:
		return s.handlePublisherFullName(c, user)
	case entity.UserPublisherStageEnterCongregationName:
		return s.handleCongregationPublisherJoinRequest(c, b, user)
	case entity.UserPublisherStageWaitingForAdminApproval:
		return c.Send(MessageWaitingForAdminApproval)
	case entity.UserAdminStageSendTerritory:
		return s.sendAddTerritoryInstruction(c)
	default:
		return s.RenderMenu(c)
	}
}

func (s *botService) handlePublisherFullName(c tb.Context, user *entity.User) error {
	logger := s.logger.
		Named("handlePublisherFullName").
		With(c).
		With("fullName", c.Message().Text)

	user.FullName = c.Message().Text
	user.Stage = entity.UserPublisherStageEnterCongregationName
	_, err := s.storages.User.UpdateUser(user)
	if err != nil {
		logger.Error("failed to update user", "error", err)
		return err
	}

	return c.Send(MessageEnterCongregationName)
}

func (s *botService) handleCongregationPublisherJoinRequest(c tb.Context, b *tb.Bot, user *entity.User) error {
	logger := s.logger.
		Named("handleCongregationPublisherJoinRequest").
		With("congregationName", c.Message().Text)

	congregation, err := s.storages.Congregation.GetCongregation(&GetCongregationFilter{
		Name: c.Message().Text,
	})
	if err != nil {
		logger.Error("failed to get congregation by name", "error", err)
		return err
	}
	if congregation == nil {
		logger.Info("congregation not found")
		return c.Send(MessageCongregationNotFound)
	}

	admin, err := s.storages.User.GetUser(&GetUserFilter{
		CongregationID: congregation.ID,
		Role:           entity.UserRoleAdmin,
	})
	if err != nil {
		logger.Error("failed to get admin user by congregation id", "err", err)
		return err
	}
	if admin == nil {
		logger.Info("admin user not found")
		return c.Send(MessageCongregationAdminNotFound)
	}

	// NOTE using this because we can't pass more than 64 bytes in callback data
	// https://github.com/nmlorg/metabot/issues/1
	message := fmt.Sprintf("<a href=\"tg://btn/%s\">\u200b</a> %s", user.ID, MessageNewJoinRequest(&MessageNewJoinRequestOptions{
		FirstName: c.Sender().FirstName,
		LastName:  c.Sender().LastName,
		Username:  c.Sender().Username,
	}))
	_, err = b.Send(&recepient{chatID: admin.MessengerChatID}, message, &tb.ReplyMarkup{
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

	user.Stage = entity.UserPublisherStageWaitingForAdminApproval
	_, err = s.storages.User.UpdateUser(user)
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

func (s *botService) RenderMenu(c tb.Context) error {
	logger := s.logger.
		Named("RenderMenu").
		With(c)

	user, err := s.storages.User.GetUser(&GetUserFilter{
		MessengerUserID: fmt.Sprint(c.Sender().ID),
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

	var buttons [][]tb.InlineButton
	if user.Role == entity.UserRoleAdmin {
		buttons = [][]tb.InlineButton{
			{tb.InlineButton{Unique: entity.ViewTerritoryListButton, Text: entity.ViewTerritoryListButton}},
			{tb.InlineButton{Unique: entity.AddTerritoryButton, Text: entity.AddTerritoryButton}},
		}
	}
	if user.Role == entity.UserRolePublisher {
		buttons = [][]tb.InlineButton{
			{tb.InlineButton{Unique: entity.ViewTerritoryListButton, Text: entity.ViewTerritoryListButton}},
		}
	}
	logger.With("buttons", buttons)
	logger.Info("successfully rendered menu buttons")

	return c.Send(MessageHowCanIHelpYou, &tb.SendOptions{
		ReplyMarkup: &tb.ReplyMarkup{
			InlineKeyboard: buttons,
		},
	})
}

func (s *botService) HandleButton(c tb.Context, b *tb.Bot) error {
	logger := s.logger.
		Named("HandleButton").
		With(c)

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
	case data == entity.AddTerritoryButton:
		return s.handleAddTerritory(c, user)
	case data == entity.ViewTerritoryListButton:
		return s.handleViewTerritoryGroupList(c, user)
	case strings.Contains(data, approvePublisherJoinRequestButtonUnique):
		publisherID := strings.Replace(c.Message().Entities[0].URL, "tg://btn/", "", -1)
		return s.handleApprovePublisherJoinRequest(c, b, user, publisherID)
	case strings.Contains(data, rejectPublisherJoinRequestButtonUnique):
		publisherID := strings.Replace(c.Message().Entities[0].URL, "tg://btn/", "", -1)
		return s.handleRejectPublisherJoinRequest(c, b, user, publisherID)
	case strings.Contains(data, territoryGroupButtonUnique):
		groupName := strings.Replace(data, territoryGroupButtonUnique, "", -1)
		return s.handleViewTerritoriesList(c, user, groupName)
	case strings.Contains(data, takeTerritoryButtonUnique):
		territoryID := strings.Replace(data, takeTerritoryButtonUnique, "", -1)
		return s.handleTakeTerritoryRequest(c, b, user, territoryID)
		// TODO: implement approve and reject buttons handlers for approve user and approve territory take
	default:
		return fmt.Errorf("unknown button: %s", data)
	}
}

func (s *botService) handleAddTerritory(c tb.Context, user *entity.User) error {
	logger := s.logger.
		Named("handleAddTerritory").
		With(c)

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
		Named("handleViewTerritoryList").
		With(c)

	var showAvailableTerritories *bool
	if user.Role != entity.UserRoleAdmin {
		logger.Info("user is not admin")
		showAvailableTerritories = &[]bool{false}[0]
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

func (s *botService) sendAddTerritoryInstruction(c tb.Context) error {
	return c.Send(MessageAddTerritoryInstruction, &tb.SendOptions{}, tb.ModeMarkdown)
}

func (s *botService) handleApprovePublisherJoinRequest(c tb.Context, b *tb.Bot, admin *entity.User, publisherID string) error {
	logger := s.logger.
		Named("handleApprovePublisherJoinRequest").
		With(c).
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

	messageID := c.Callback().Message.ID
	_, err = b.Edit(&editable{
		chatID:    c.Callback().Message.Chat.ID,
		messageID: fmt.Sprintf("%d", messageID),
	}, MessageCongregationJoinRequestApprovedDone(publisher.FullName))
	if err != nil {
		logger.Error("failed to edit message", "err", err)
		return err
	}

	return nil
}

func (s *botService) handleRejectPublisherJoinRequest(c tb.Context, b *tb.Bot, admin *entity.User, publisherID string) error {
	logger := s.logger.
		Named("handleRejectPublisherJoinRequest").
		With(c).
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

	_, err = b.Send(&recepient{chatID: publisher.MessengerChatID}, MessageCongregationJoinRequestRejected)
	if err != nil {
		logger.Error("failed to send message to publisher", "err", err)
		return err
	}

	messageID := c.Callback().Message.ID
	_, err = b.Edit(&editable{
		chatID:    c.Callback().Message.Chat.ID,
		messageID: fmt.Sprintf("%d", messageID),
	}, MessageCongregationJoinRequestRejectedDone(publisher.FullName))
	if err != nil {
		logger.Error("failed to edit message", "err", err)
		return err
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
		Named("handleViewTerritoriesList").
		With(c)

	// FIXME: we should not use create method
	group, err := s.storages.Congregation.GetOrCreateCongregationTerritoryGroup(&GetOrCreateCongregationTerritoryGroupOptions{
		CongregationID: user.CongregationID,
		Title:          groupName,
	})
	if err != nil {
		logger.Error("failed to get or create territory group", "err", err)
		return err
	}

	territories, err := s.storages.Congregation.ListTerritories(&ListTerritoriesFilter{
		CongregationID: user.CongregationID,
		GroupID:        group.ID,
	})
	if err != nil {
		logger.Error("failed to list territories", "err", err)
		return err
	}

	for _, territory := range territories {
		photo := tb.Photo{
			File: tb.File{
				FileID: territory.FileID,
			},
			Caption: territory.Title,
		}

		// Create inline button
		button := tb.InlineButton{
			Unique: territory.ID + takeTerritoryButtonUnique,
			Text:   fmt.Sprintf("Взяти %s", territory.Title),
		}

		keyboard := [][]tb.InlineButton{{button}}
		markup := &tb.ReplyMarkup{InlineKeyboard: keyboard}

		err := c.Send(&photo, &tb.SendOptions{
			ReplyMarkup: markup,
		}, tb.ModeMarkdown)
		if err != nil {
			logger.Error("failed to send photo", "err", err)
			return err
		}
	}

	return nil
}

func (s *botService) handleTakeTerritoryRequest(c tb.Context, b *tb.Bot, user *entity.User, territoryID string) error {
	logger := s.logger.
		Named("handleTakeTerritoryRequest").
		With(c)

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
	if !*territory.IsAvailable {
		logger.Info("territory is not available")
		return c.Send(MessageTerritoryNotAvailable)
	}

	admin, err := s.storages.User.GetUser(&GetUserFilter{
		CongregationID: user.CongregationID,
		Role:           entity.UserRoleAdmin,
	})
	if err != nil {
		logger.Error("failed to get admin user by congregation id", "err", err)
		return err
	}
	if admin == nil {
		logger.Info("admin user not found")
		return c.Send(MessageCongregationAdminNotFound)
	}

	message := MessageTakeTerritoryRequest(user, territory.Title)
	_, err = b.Send(&recepient{chatID: admin.MessengerChatID}, message, &tb.ReplyMarkup{
		InlineKeyboard: [][]tb.InlineButton{
			{
				tb.InlineButton{
					Unique: entity.ApproveTakeTerritoryButton,
					Text:   entity.ApproveTakeTerritoryButton,
				},
				tb.InlineButton{
					Unique: entity.RejectTakeTerritoryButton,
					Text:   entity.RejectTakeTerritoryButton,
				},
			},
		},
	})
	if err != nil {
		logger.Error("failed to send message to admin", "err", err)
		return err
	}

	return c.Send(MessageTakeTerritoryRequestSent)
}

func (s *botService) handleApproveTerritoryTake(c tb.Context, b *tb.Bot, user *entity.User, publisherID string, territoryID string) error {
	logger := s.logger.
		Named("handleApproveTerritoryTake").
		With(c)

	// FIXME: now we pass user as admin, but we should pass publisher user

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

	notAvailable := false
	territory.IsAvailable = &notAvailable
	territory.InUseByUserID = &publisherID
	territory, err = s.storages.Congregation.UpdateTerritory(territory)
	if err != nil {
		logger.Error("failed to update territory", "err", err)
		return err
	}

	message := MessageTakeTerritoryRequestApproved(territory.Title)
	// TODO: test if sends message to publisher
	err = c.Send(&recepient{chatID: publisher.MessengerChatID}, message)
	if err != nil {
		logger.Error("failed to send message to user", "err", err)
		return err
	}

	admin, err := s.storages.User.GetUser(&GetUserFilter{
		CongregationID: user.CongregationID,
		Role:           entity.UserRoleAdmin,
	})
	if err != nil {
		logger.Error("failed to get admin user by congregation id", "err", err)
		return err
	}
	if admin == nil {
		logger.Info("admin user not found")
		return c.Send(MessageCongregationAdminNotFound)
	}

	// TODO: update message with buttons approve/reject to change it to ✅ so admin can see who he approved already
	return c.Send("Temp message")
}

func (s *botService) HandleImageUpload(c tb.Context) error {
	logger := s.logger.
		Named("HandleImageUpload").
		With(c)

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

	available := true
	territory, err = s.storages.Congregation.CreateTerritory(&entity.CongregationTerritory{
		CongregationID: congregation.ID,
		GroupID:        group.ID,
		Title:          territoryName,
		FileID:         fileID,
		IsAvailable:    &available,
	})
	if err != nil {
		logger.Error("failed to create territory", "err", err)
		return err
	}
	logger.With("territory", territory)

	logger.Info("successfully handled image upload")
	return c.Send(fmt.Sprintf("Територія %s успішно додана в групу %s!", territoryName, groupName), &tb.SendOptions{}, tb.ModeMarkdown)
}
