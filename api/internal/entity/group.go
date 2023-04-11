package entity

type Group struct {
	ID    string `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Title string // can be place name like Kiev, Lviv, etc.
}
