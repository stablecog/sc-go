package scworker

import "fmt"

// Error wrapper for GPU worker requests
type WorkerError struct {
	StatusCode int

	Err error
}

func (r *WorkerError) Error() string {
	return fmt.Sprintf("status %d: err %v", r.StatusCode, r.Err)
}
