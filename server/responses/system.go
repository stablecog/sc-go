package responses

import "github.com/stablecog/sc-go/shared"

type ChangeSystemBackendResponse struct {
	Backend shared.BackendType `json:"backend"`
	Error   string             `json:"error,omitempty"`
}

type SystemStatusResponse struct {
	Backend shared.BackendType `json:"backend"`
}
