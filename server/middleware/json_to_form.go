package middleware

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

func JsonToFormMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") == "application/json" {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read request body", http.StatusInternalServerError)
				return
			}

			// Restore the request body
			r.Body = io.NopCloser(r.Body)

			var formData map[string]string
			if err := json.Unmarshal(body, &formData); err != nil {
				http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
				return
			}

			// Populate form values
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Failed to parse form", http.StatusInternalServerError)
				return
			}

			for key, value := range formData {
				r.Form.Set(key, value)
			}
		}

		// Pass to the next middleware/handler
		next.ServeHTTP(w, r)
	})
}
