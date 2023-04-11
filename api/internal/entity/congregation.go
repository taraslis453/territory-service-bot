package entity

type Congregation struct {
	ID    string             `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Users []CongregationUser `gorm:"foreignKey:CongregationID"`
}

type CongregationUser struct {
	ID              string `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	CongregationID  string `json:"congregation_id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	MessengerUserID string
	Role            Role
}

type Role string

const (
	RoleAdmin     Role = "admin"
	RolePublisher Role = "publisher"
)
