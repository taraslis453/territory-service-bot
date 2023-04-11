package entity

type Territory struct {
	ID        string `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Title     string
	GroupID   string
	ImageID   string
	Available *bool
}
