package requests

type ScheduleForDeleteAction string

const (
	DeleteAction   ScheduleForDeleteAction = "schedule-for-deletion"
	UndeleteAction ScheduleForDeleteAction = "cancel-deletion"
)

type DeleteUserRequest struct {
	Action ScheduleForDeleteAction `json:"action"`
}
