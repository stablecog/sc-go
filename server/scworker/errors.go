package scworker

import (
	"fmt"
	"net/http"
)

// Error wrapper for GPU worker requests
type WorkerError struct {
	StatusCode int

	Err error

	ErrDescription string
}

// Suppresses internal errors
func WorkerInternalServerError() *WorkerError {
	return &WorkerError{http.StatusInternalServerError, fmt.Errorf("An unknown error has occured"), ""}
}

func (r *WorkerError) Error() string {
	return r.Err.Error()
}
