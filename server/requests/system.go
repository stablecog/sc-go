package requests

import "github.com/stablecog/sc-go/shared"

type ChangeSystemBackendRequest struct {
	Backend shared.BackendType `json:"backend"`
}
