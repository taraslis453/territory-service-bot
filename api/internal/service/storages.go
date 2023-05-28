package service

import "github.com/taraslis453/territory-service-bot/internal/entity"

type Storages struct {
	User         UserStorage
	Congregation CongregationStorage
	Chat         ChatStorage
}

type UserStorage interface {
	CreateUser(*entity.User) (*entity.User, error)
	GetUser(filter *GetUserFilter) (*entity.User, error)
	UpdateUser(*entity.User) (*entity.User, error)
	ListUsers(filter *ListUsersFilter) ([]entity.User, error)
}

type GetUserFilter struct {
	ID              string
	MessengerUserID string
	CongregationID  string
	Role            entity.UserRole
}

type ListUsersFilter struct {
	CongregationID string
	Role           entity.UserRole
}

type CongregationStorage interface {
	GetCongregation(filter *GetCongregationFilter) (*entity.Congregation, error)
	GetOrCreateCongregationTerritoryGroup(options *GetOrCreateCongregationTerritoryGroupOptions) (*entity.CongregationTerritoryGroup, error)
	CreateTerritory(*entity.CongregationTerritory) (*entity.CongregationTerritory, error)
	GetTerritory(filter *GetTerritoryFilter) (*entity.CongregationTerritory, error)
	ListTerritories(filter *ListTerritoriesFilter) ([]entity.CongregationTerritory, error)
	ListTerritoryGroups(filter *ListTerritoryGroupsFilter) ([]entity.CongregationTerritoryGroup, error)
	UpdateTerritory(territory *entity.CongregationTerritory) (*entity.CongregationTerritory, error)
	AddTerritoryNote(territory *entity.CongregationTerritoryNote) (*entity.CongregationTerritoryNote, error)
}

type GetCongregationFilter struct {
	ID   string
	Name string
}

type GetOrCreateCongregationTerritoryGroupOptions struct {
	CongregationID string
	Title          string
}

type GetTerritoryFilter struct {
	ID             string
	CongregationID string
	Title          string
	GroupID        string
}

type ListTerritoriesFilter struct {
	CongregationID string
	GroupID        string
	Available      *bool
	InUseByUserID  string
	SortBy         string
}

type ListTerritoryGroupsFilter struct {
	CongregationID string
	IDs            []string
}

type ChatStorage interface {
	CreateRequestActionState(*entity.RequestActionState) (*entity.RequestActionState, error)
	GetRequestActionState(id string) (*entity.RequestActionState, error)
	DeleteRequestActionState(id string) error
}
