package storage

import (

	// third party
	"fmt"

	"github.com/taraslis453/territory-service-bot/internal/entity"
	"github.com/taraslis453/territory-service-bot/internal/service"
	"github.com/taraslis453/territory-service-bot/pkg/database"
	"gorm.io/gorm"
)

type congregationStorage struct {
	database.Database
}

var _ service.CongregationStorage = (*congregationStorage)(nil)

func NewCongregationStorage(database database.Database) *congregationStorage {
	return &congregationStorage{database}
}

func (r *congregationStorage) GetCongregation(filter *service.GetCongregationFilter) (*entity.Congregation, error) {
	stmt := r.Instance()
	if filter.ID != "" {
		stmt = stmt.Where(&entity.Congregation{ID: filter.ID})
	}
	if filter.Name != "" {
		stmt = stmt.Where(&entity.Congregation{Name: filter.Name})
	}

	congregation := entity.Congregation{}
	err := stmt.
		Take(&congregation).
		Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &congregation, nil
}

func (r *congregationStorage) GetOrCreateCongregationTerritoryGroup(options *service.CreateOrGetCongregationTerritoryGroupOptions) (*entity.CongregationTerritoryGroup, error) {
	territoryGroup := entity.CongregationTerritoryGroup{}
	err := r.Instance().
		Where(&entity.CongregationTerritoryGroup{CongregationID: options.CongregationID, Title: options.Title}).
		Take(&territoryGroup).
		Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// crete new territory group
			err = r.Instance().Create(&entity.CongregationTerritoryGroup{
				CongregationID: options.CongregationID,
				Title:          options.Title,
			}).Error
			if err != nil {
				return nil, err
			}
			err = r.Instance().
				Where(&entity.CongregationTerritoryGroup{CongregationID: options.CongregationID, Title: options.Title}).
				Take(&territoryGroup).
				Error
			if err != nil {
				return nil, err
			}

			return &territoryGroup, nil
		}
		return nil, err
	}

	return &territoryGroup, nil
}

func (r *congregationStorage) CreateTerritory(territory *entity.CongregationTerritory) (*entity.CongregationTerritory, error) {
	err := r.Instance().Create(territory).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create territory: %w", err)
	}

	err = r.Instance().
		Where(&entity.CongregationTerritory{Title: territory.Title}).
		Take(&territory).
		Error
	if err != nil {
		return nil, fmt.Errorf("failed to get created territory: %w", err)
	}

	return territory, nil
}

func (r *congregationStorage) GetTerritory(filter *service.GetTerritoryFilter) (*entity.CongregationTerritory, error) {
	stmt := r.Instance()
	if filter.CongregationID != "" {
		stmt = stmt.Where(&entity.CongregationTerritory{CongregationID: filter.CongregationID})
	}
	if filter.Title != "" {
		stmt = stmt.Where(&entity.CongregationTerritory{Title: filter.Title})
	}

	territory := entity.CongregationTerritory{}
	err := stmt.
		Take(&territory).
		Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &territory, nil
}
