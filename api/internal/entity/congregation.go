package entity

import "time"

type Congregation struct {
	ID     string                       `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Name   string                       `gorm:"uniqueIndex"`
	Groups []CongregationTerritoryGroup `gorm:"foreignkey:CongregationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type CongregationTerritoryGroup struct {
	ID             string `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	CongregationID string `gorm:"type:uuid;index"`
	Title          string // can be place name like Kiev, Lviv, etc.
}

type CongregationTerritory struct {
	ID             string `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	CongregationID string `gorm:"type:uuid;index"`
	Title          string `gorm:"index"`
	GroupID        string
	FileID         string
	FileType       CongregationTerritoryFileType
	InUseByUserID  *string `gorm:"index"`
	// NOTE: when user takes territory, we update this field and when user returns territory, we update this field
	LastTakenAt time.Time
	Notes       []CongregationTerritoryNote `gorm:"foreignkey:TerritoryID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type CongregationTerritoryFileType string

var (
	CongregationTerritoryFileTypePhoto    CongregationTerritoryFileType = "photo"
	CongregationTerritoryFileTypeDocument CongregationTerritoryFileType = "document"
)

type CongregationTerritoryNote struct {
	ID          string `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	TerritoryID string `gorm:"type:uuid;index"`
	UserID      string `gorm:"type:uuid;index"`
	Text        string
	CreatedAt   time.Time
}
