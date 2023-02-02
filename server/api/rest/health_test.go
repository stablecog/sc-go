package rest

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Health API, just returns a 200 to know our instance is running
func TestHandleHealth(t *testing.T) {
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Content-Type", "application/json")

	MockController.HandleHealth(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
}
