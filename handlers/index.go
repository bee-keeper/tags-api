package handlers

import (
	"net/http"

	"gorm.io/gorm"
)

// Index - API root
func Index(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	if r.URL.Path != "/v1" {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		w.WriteHeader(http.StatusOK)
		// https://github.com/swaggo/swag
		w.Write([]byte("With more time I'd return a detailed API spec here."))

	case http.MethodOptions:
		w.Header().Set("Allow", "GET, OPTIONS")
		w.WriteHeader(http.StatusNoContent)

	default:
		w.Header().Set("Allow", "GET, OPTIONS")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
