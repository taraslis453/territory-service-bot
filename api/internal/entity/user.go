package entity

type User struct {
	ID              string `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	CongregationID  string
	MessengerUserID string
	MessengerChatID string
	Role            UserRole
	Stage           UserStage
}

type UserRole string

const (
	UserRoleAdmin     UserRole = "admin"
	UserRolePublisher UserRole = "publisher"
)

type UserStage string

const (
	UserPublisherStageEnterCongregationName   UserStage = "user_publisher_enter_congregation_name"
	UserPublisherStageWaitingForAdminApproval UserStage = "user_publisher_waiting_for_admin_approval"
	UserStageSelectActionFromMenu             UserStage = "user_select_action_from_menu"
	UserAdminStageSendTerritory               UserStage = "user_admin_send_territory"
)
