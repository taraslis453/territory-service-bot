package entity

type User struct {
	ID                 string `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	JoinCongregationID string // represents in which congeration user wants to join
	CongregationID     string // represents in which congeration user in
	MessengerUserID    string
	MessengerChatID    string
	FullName           string
	Role               UserRole
	Stage              UserStage
}

type UserRole string

const (
	UserRoleAdmin     UserRole = "admin"
	UserRolePublisher UserRole = "publisher"
)

type UserStage string

const (
	UserPublisherStageEnterFullName                   UserStage = "user_publisher_enter_full_name"
	UserPublisherStageEnterCongregationName           UserStage = "user_publisher_enter_congregation_name"
	UserPublisherStageWaitingForAdminApproval         UserStage = "user_publisher_waiting_for_admin_approval"
	UserPublisherStageCongregationJoinRequestRejected UserStage = "user_publisher_congregation_join_request_rejected"
	UserStageSelectActionFromMenu                     UserStage = "user_select_action_from_menu"
	UserAdminStageSendTerritory                       UserStage = "user_admin_send_territory"
	UserStageLeaveTerritoryNote                       UserStage = "user_leave_territory_note"
)
