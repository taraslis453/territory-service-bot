package storage

import (
	"fmt"

	// third party
	"github.com/taraslis453/territory-service-bot/internal/entity"
	"github.com/taraslis453/territory-service-bot/internal/service"
	"github.com/taraslis453/territory-service-bot/pkg/database"
	"gorm.io/gorm"
)

type userStorage struct {
	database.Database
}

var _ service.UserStorage = (*userStorage)(nil)

func NewUserStorage(database database.Database) *userStorage {
	return &userStorage{database}
}

func (r *userStorage) CreateUser(user *entity.User) (*entity.User, error) {
	err := r.Instance().Create(user).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	err = r.Instance().
		Where(&entity.User{MessengerUserID: user.MessengerUserID}).
		Take(user).
		Error
	if err != nil {
		return nil, fmt.Errorf("failed to get created user: %w", err)
	}

	return user, nil
}

func (r *userStorage) GetUser(filter *service.GetUserFilter) (*entity.User, error) {
	stmt := r.Instance()
	if filter.ID != "" {
		stmt = stmt.Where(&entity.User{ID: filter.ID})
	}
	if filter.MessengerUserID != "" {
		stmt = stmt.Where(&entity.User{MessengerUserID: filter.MessengerUserID})
	}
	if filter.CongregationID != "" {
		stmt = stmt.Where(&entity.User{CongregationID: filter.CongregationID})
	}
	if filter.Role != "" {
		stmt = stmt.Where(&entity.User{Role: filter.Role})
	}

	user := entity.User{}
	err := stmt.
		Take(&user).
		Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *userStorage) ListUsers(filter *service.ListUsersFilter) ([]entity.User, error) {
	stmt := r.Instance()
	if filter.CongregationID != "" {
		stmt = stmt.Where(&entity.User{CongregationID: filter.CongregationID})
	}
	if filter.Role != "" {
		stmt = stmt.Where(&entity.User{Role: filter.Role})
	}

	users := make([]entity.User, 0)
	err := stmt.
		Find(&users).
		Error
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userStorage) UpdateUser(user *entity.User) (*entity.User, error) {
	err := r.Instance().
		Where(&entity.User{ID: user.ID}).
		Updates(user).
		Error
	if err != nil {
		return nil, err
	}

	return user, nil
}
