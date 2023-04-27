package requests

import "github.com/google/uuid"

type BanAction string

const (
	BanActionBan   BanAction = "ban"
	BanActionUnban BanAction = "unban"
)

type BanUsersRequest struct {
	Action  BanAction   `json:"action"`
	UserIDs []uuid.UUID `json:"user_ids"`
}
