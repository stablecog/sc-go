package responses

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

var UnableToParseJsonError = ErrorResponse{
	Error: "json_parse_error",
}

func ErrUnableToParseJson(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusBadRequest)
	render.JSON(w, r, &UnableToParseJsonError)
}

var UnauthorizedError = ErrorResponse{
	Error: "Unauthorized",
}

func ErrUnauthorized(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusUnauthorized)
	render.JSON(w, r, &UnauthorizedError)
}

var InsufficientCredits = ErrorResponse{
	Error: "insufficient_credits",
}

var InsufficientCreditsErr = errors.New(InsufficientCredits.Error)

func ErrInsufficientCredits(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusBadRequest)
	render.JSON(w, r, &InsufficientCredits)
}

func ErrNotFound(w http.ResponseWriter, r *http.Request, errorText string) {
	render.Status(r, http.StatusNotFound)
	render.JSON(w, r, &ErrorResponse{
		Error: errorText,
	})
}

func ErrBadRequest(w http.ResponseWriter, r *http.Request, errorText string) {
	render.Status(r, http.StatusBadRequest)
	render.JSON(w, r, &ErrorResponse{
		Error: errorText,
	})
}

func ErrInternalServerError(w http.ResponseWriter, r *http.Request, errorText string) {
	render.Status(r, http.StatusInternalServerError)
	render.JSON(w, r, &ErrorResponse{
		Error: errorText,
	})
}

func ErrMethodNotAllowed(w http.ResponseWriter, r *http.Request, errorText string) {
	render.Status(r, http.StatusMethodNotAllowed)
	render.JSON(w, r, &ErrorResponse{
		Error: errorText,
	})
}
