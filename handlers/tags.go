package handlers

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"
)

// Tag - data representation
type Tag struct {
	gorm.Model
	Name string `json:"name" gorm:"unique"`
}

// Tags - HTTP methods for tag operations
func Tags(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	if r.URL.Path != "/v1/tags" {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		var tags []Tag
		// fetch tags
		if result := db.Find(&tags); result.Error != nil {
			http.Error(w, "Failed to fetch tags", http.StatusInternalServerError)
			return
		}
		// list tags
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(tags); err != nil {
			http.Error(w, "Failed to encode tags", http.StatusInternalServerError)
			return
		}

	case http.MethodPost:
		var tag Tag
		if err := json.NewDecoder(r.Body).Decode(&tag); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}
		// check existing tags
		var existingTag Tag
		if result := db.Where("name = ?", tag.Name).First(&existingTag); result.RowsAffected > 0 {
			http.Error(w, "Tag with this name already exists", http.StatusConflict)
			return
		}
		// create tag
		if result := db.Create(&tag); result.Error != nil {
			http.Error(w, "Failed to create tag", http.StatusInternalServerError)
			return
		}
		// return status
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(tag)

	case http.MethodOptions:
		w.Header().Set("Allow", "GET, POST, OPTIONS")
		w.WriteHeader(http.StatusNoContent)

	default:
		w.Header().Set("Allow", "GET, POST, OPTIONS")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
