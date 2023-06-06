// * Requests initiated by logged in users
package requests

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/utils"
)

type QueryVoiceoverFilters struct {
	Order            SortOrder  `json:"order"`
	StartDt          *time.Time `json:"start_dt"`
	EndDt            *time.Time `json:"end_dt"`
	UserID           *uuid.UUID `json:"user_id"`
	OrderBy          OrderBy    `json:"order_by"`
	IsFavorited      *bool      `json:"is_favorited,omitempty"`
	WasAutoSubmitted *bool      `json:"was_auto_submitted,omitempty"`
	PromptID         *uuid.UUID `json:"prompt,omitempty"`
}

// Parse all filters into a QueryGenerationFilters struct
func (filters *QueryVoiceoverFilters) ParseURLQueryParameters(urlValues url.Values) error {
	for key, value := range urlValues {
		// Order
		if key == "order" {
			if strings.ToLower(value[0]) == string(SortOrderAscending) {
				filters.Order = SortOrderAscending
			} else if strings.ToLower(value[0]) == string(SortOrderDescending) {
				filters.Order = SortOrderDescending
			} else {
				return fmt.Errorf("invalid order: '%s' expected '%s' or '%s'", value[0], SortOrderAscending, SortOrderDescending)
			}
		}

		// Start and end date
		if key == "start_dt" {
			startDt, err := utils.ParseIsoTime(value[0])
			if err != nil {
				return fmt.Errorf("invalid start_dt: %s", value[0])
			}
			filters.StartDt = &startDt
		}
		if key == "end_dt" {
			endDt, err := utils.ParseIsoTime(value[0])
			if err != nil {
				return fmt.Errorf("invalid end_dt: %s", value[0])
			}
			filters.EndDt = &endDt
		}
		if key == "order_by" {
			if strings.ToLower(value[0]) == string(OrderByUpdatedAt) {
				filters.OrderBy = OrderByUpdatedAt
			} else if strings.ToLower(value[0]) == string(OrderByCreatedAt) {
				filters.OrderBy = OrderByCreatedAt
			} else {
				return fmt.Errorf("invalid order_by: '%s' expected '%s' or '%s'", value[0], OrderByUpdatedAt, OrderByCreatedAt)
			}
		}

		// Favorited
		if key == "is_favorited" {
			if strings.ToLower(value[0]) == "true" {
				t := true
				filters.IsFavorited = &t
			} else if strings.ToLower(value[0]) == "false" {
				f := false
				filters.IsFavorited = &f
			} else {
				return fmt.Errorf("invalid is_favorited: '%s' expected 'true' or 'false'", value[0])
			}
		}

		// Was auto submitted
		if key == "was_auto_submitted" {
			if strings.ToLower(value[0]) == "true" {
				t := true
				filters.WasAutoSubmitted = &t
			} else if strings.ToLower(value[0]) == "false" {
				f := false
				filters.WasAutoSubmitted = &f
			} else {
				return fmt.Errorf("invalid was_auto_submitted: '%s' expected 'true' or 'false'", value[0])
			}
		}

		// Prompt id
		if key == "prompt_id" {
			parsed, err := uuid.Parse(value[0])
			if err != nil {
				return fmt.Errorf("invalid prompt_id: %s", value[0])
			}
			filters.PromptID = &parsed
		}
	}
	// Descending default
	if filters.Order == "" {
		filters.Order = SortOrderDescending
	}

	// Sort by created_at by default
	if filters.OrderBy == "" {
		filters.OrderBy = OrderByCreatedAt
	}
	return nil
}
