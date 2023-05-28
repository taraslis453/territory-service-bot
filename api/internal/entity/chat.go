package entity

import "github.com/taraslis453/territory-service-bot/pkg/database/datatypes"

const (
	AddTerritoryButton         = "🌍 Додати територію"
	ViewTerritoryListButton    = "🔍 Пошук територій"
	ViewMyTerritoryListButton  = "🗂️ Мої території"
	ApprovePublisherButton     = "✅ Прийняти"
	RejectPublisherButton      = "❌ Відхилити"
	ApproveTakeTerritoryButton = "✅ Прийняти"
	RejectTakeTerritoryButton  = "❌ Відхилити "
	TakeTerritoryButton        = "🗺️ Взяти територію "
	LeaveTerritoryNoteButton   = "📝 Залишити нотатку "
	ReturnTerritoryButton      = "🔄 Повернути територію"
)

// We suppose that we can have multiple admins.
type RequestActionState struct {
	ID string `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	// Keep messages id for each request in order to syncronize state of actions (approved, rejected, etc.)
	AdminMessages datatypes.Slice[AdminMessage]
}

type AdminMessage struct {
	ChatID    string
	MessageID string
}
