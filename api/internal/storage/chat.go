package storage

import (
	"fmt"

	"github.com/taraslis453/territory-service-bot/internal/entity"
	"github.com/taraslis453/territory-service-bot/internal/service"
	"github.com/taraslis453/territory-service-bot/pkg/database"
)

type chatStorage struct {
	database.Database
}

var _ service.ChatStorage = (*chatStorage)(nil)

func NewChatStorage(database database.Database) *chatStorage {
	return &chatStorage{database}
}

func (s *chatStorage) CreateRequestActionState(requestActionState *entity.RequestActionState) (*entity.RequestActionState, error) {
	err := s.Instance().Create(requestActionState).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create territory: %w", err)
	}

	err = s.Instance().
		Where(&entity.RequestActionState{ID: requestActionState.ID}).
		Take(&requestActionState).
		Error
	if err != nil {
		return nil, fmt.Errorf("failed to get created territory: %w", err)
	}

	return requestActionState, nil

}

func (s *chatStorage) DeleteRequestActionState(id string) error {
	// NOTE: using hard delete because we don't need to keep this data
	err := s.Instance().Exec("DELETE FROM request_action_states WHERE id = ?", id).Error
	if err != nil {
		return fmt.Errorf("failed to delete request action state: %w", err)
	}

	return nil
}

func (s *chatStorage) GetRequestActionState(id string) (*entity.RequestActionState, error) {
	requestActionState := entity.RequestActionState{}
	err := s.Instance().
		Where(&entity.RequestActionState{ID: id}).
		Take(&requestActionState).
		Error
	if err != nil {
		return nil, fmt.Errorf("failed to get request action state: %w", err)
	}

	return &requestActionState, nil
}
