package rest

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/render"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/utils"
)

// HTTP Get - all operations for user
// Takes query paramers for pagination
// per_page: number of generations to return
// cursor: cursor for pagination, it is an iso time string in UTC
func (c *RestAPI) HandleQueryOperations(w http.ResponseWriter, r *http.Request) {
	var user *ent.User
	if user = c.GetUserIfAuthenticated(w, r); user == nil {
		return
	}

	// Validate query parameters
	perPage := DEFAULT_PER_PAGE
	var err error
	if perPageStr := r.URL.Query().Get("per_page"); perPageStr != "" {
		perPage, err = strconv.Atoi(perPageStr)
		if err != nil {
			responses.ErrBadRequest(w, r, "per_page must be an integer", "")
			return
		} else if perPage < 1 || perPage > MAX_PER_PAGE {
			responses.ErrBadRequest(w, r, fmt.Sprintf("per_page must be between 1 and %d", MAX_PER_PAGE), "")
			return
		}
	}

	var cursor *time.Time
	if cursorStr := r.URL.Query().Get("cursor"); cursorStr != "" {
		cursorTime, err := utils.ParseIsoTime(cursorStr)
		if err != nil {
			responses.ErrBadRequest(w, r, "cursor must be a valid iso time string", "")
			return
		}
		cursor = &cursorTime
	}

	operations, err := c.Repo.QueryUserOperations(user.ID, perPage, cursor)
	if err != nil {
		log.Error("Error getting operations for user", "err", err)
		responses.ErrInternalServerError(w, r, "Error getting operations")
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, operations)
}
