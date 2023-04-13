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
			Stage:           entity.UserPublisherStageEnterCongregationName,
		})
		if err != nil {
			logger.Error("failed to create user", "error", err)
			return err
		}

		return c.Send(MessageEnterCongregationName)
	}
	if user.CongregationID == "" && user.Stage == entity.UserPublisherStageEnterCongregationName {
		logger.Info("user not joined to congregation")
		return c.Send(MessageEnterCongregationName)
	}
	if user.Stage == entity.UserPublisherStageWaitingForAdminApproval {
		logger.Info("user waiting for admin approval")
		return c.Send(MessageWaitingForAdminApproval)
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
	case entity.UserPublisherStageEnterCongregationName:
		return s.handleCongregationJoinRequest(c, b)
	case entity.UserAdminStageSendTerritory:
		return s.sendAddTerritoryInstruction(c)
	default:
		// TODO: handle other cases
		return nil
	}
}

func (s *botService) handleCongregationJoinRequest(c tb.Context, b *tb.Bot) error {
	logger := s.logger.
		Named("handleCongregationJoinRequest").
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

	message := MessageNewJoinRequest(&MessageNewJoinRequestOptions{
		FirstName: c.Sender().FirstName,
		LastName:  c.Sender().LastName,
		Username:  c.Sender().Username,
	})
	_, err = b.Send(&recepient{chatID: admin.MessengerChatID}, message, &tb.ReplyMarkup{
		InlineKeyboard: [][]tb.InlineButton{
			{
				tb.InlineButton{
					Unique: entity.ApprovePublisherButton,
					Text:   entity.ApprovePublisherButton,
				},
				tb.InlineButton{
					Unique: entity.RejectPublisherButton,
					Text:   entity.RejectPublisherButton,
				},
			},
		},
	})
	if err != nil {
		logger.Error("failed to send message to admin", "err", err)
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

	var buttons [][]string
	if user.Role == entity.UserRoleAdmin {
		buttons = [][]string{
			{entity.ViewTerritoryListButton},
			{entity.AddTerritoryButton},
		}
	}
	if user.Role == entity.UserRolePublisher {
		buttons = [][]string{
			{entity.ViewTerritoryListButton},
		}
	}
	logger.With("buttons", buttons)
	logger.Info("successfully rendered menu buttons")

	return c.Send(MessageHowCanIHelpYou, &tb.SendOptions{
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

func (s *botService) HandleButton(c tb.Context) error {
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

	switch data {
	case entity.AddTerritoryButton:
		return s.handleAddTerritory(c, user)
	case entity.ViewTerritoryListButton:
		return s.handleViewTerritoryList(c, user)
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

func (s *botService) handleViewTerritoryList(c tb.Context, user *entity.User) error {
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

	var buttons [][]string
	for groupName, territoriesCount := range countAvailableTerritoriesInGroups {
		buttons = append(buttons, []string{fmt.Sprintf("%s (%d)", groupName, territoriesCount)})
	}

	return c.Send(MessageTerritoryList, &tb.SendOptions{
		ReplyMarkup: s.renderInlineKeyboard(buttons),
	}, tb.ModeMarkdown)
}

func (s *botService) sendAddTerritoryInstruction(c tb.Context) error {
	return c.Send(MessageAddTerritoryInstruction, &tb.SendOptions{}, tb.ModeMarkdown)
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

	group, err := s.storages.Congregation.GetOrCreateCongregationTerritoryGroup(&CreateOrGetCongregationTerritoryGroupOptions{
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

// NOTE HOWTO: send files
// c.Send(&tb.Photo{File: tb.File{
// 	FileID: c.Message().Photo.FileID,
// }})
