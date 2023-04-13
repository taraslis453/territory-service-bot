package service

import "github.com/taraslis453/territory-service-bot/internal/entity"

type Storages struct {
	User         UserStorage
	Congregation CongregationStorage
}

type UserStorage interface {
	CreateUser(*entity.User) (*entity.User, error)
	GetUser(filter *GetUserFilter) (*entity.User, error)
	UpdateUser(*entity.User) (*entity.User, error)
}

type GetUserFilter struct {
	MessengerUserID string
	CongregationID  string
	Role            entity.UserRole
}

type CongregationStorage interface {
	GetCongregation(filter *GetCongregationFilter) (*entity.Congregation, error)
	GetOrCreateCongregationTerritoryGroup(options *CreateOrGetCongregationTerritoryGroupOptions) (*entity.CongregationTerritoryGroup, error)
	CreateTerritory(*entity.CongregationTerritory) (*entity.CongregationTerritory, error)
	GetTerritory(filter *GetTerritoryFilter) (*entity.CongregationTerritory, error)
	ListTerritories(filter *ListTerritoriesFilter) ([]entity.CongregationTerritory, error)
	ListTerritoryGroups(filter *ListTerritoryGroupsFilter) ([]entity.CongregationTerritoryGroup, error)
}

type GetCongregationFilter struct {
	ID   string
	Name string
}

type CreateOrGetCongregationTerritoryGroupOptions struct {
	CongregationID string
	Title          string
}

type GetTerritoryFilter struct {
	CongregationID string
	Title          string
	GroupID        string
}

type ListTerritoriesFilter struct {
	CongregationID string
	GroupID        string
	Available      *bool
}

type ListTerritoryGroupsFilter struct {
	CongregationID string
	IDs            []string
}
