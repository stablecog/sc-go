package requests

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

type DeactiveApiTokenRequest struct {
	ID uuid.UUID `json:"id"`
}

type NewTokenRequest struct {
	Name string `json:"name"`
}

// Filters for querying
type ApiTokenType string

const (
	ApiTokenAny ApiTokenType = "any"
	// Only auth client
	ApiTokenClient ApiTokenType = "client"
	// Only created by user manually
	ApiTokenManual ApiTokenType = "manual"
)

type ApiTokenQueryFilters struct {
	ApiTokenType ApiTokenType `json:"api_token_type"`
}

// Parse all filters into a QueryGenerationFilters struct
func (filters *ApiTokenQueryFilters) ParseURLQueryParameters(urlValues url.Values) error {
	for key, value := range urlValues {
		// Type
		if key == "type" {
			if strings.ToLower(value[0]) == string(ApiTokenAny) {
				filters.ApiTokenType = ApiTokenAny
			} else if strings.ToLower(value[0]) == string(ApiTokenClient) {
				filters.ApiTokenType = ApiTokenClient
			} else if strings.ToLower(value[0]) == string(ApiTokenManual) {
				filters.ApiTokenType = ApiTokenManual
			} else {
				return fmt.Errorf("invalid type: '%s' expected '%s', '%s', or '%s'", value[0], ApiTokenAny, ApiTokenClient, ApiTokenManual)
			}
		}
	}

	// Only user-created ones by default
	if filters.ApiTokenType == "" {
		filters.ApiTokenType = ApiTokenManual
	}
	return nil
}
