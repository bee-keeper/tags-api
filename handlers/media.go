package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"gorm.io/gorm"
)

// Media - data representation
type Media struct {
	gorm.Model
	Name string `json:"name"`
	Tags []*Tag `json:"tags" gorm:"many2many:media_tags"`
	URL  string `json:"URL" gorm:"unique"`
	File []byte `json:"-" gorm:"-"`
}

// AllMedia - HTTP methods for media operations
func AllMedia(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	if r.URL.Path != "/v1/media" {
		http.NotFound(w, r)
		return
	}

	switch r.Method {

	case http.MethodGet:
		tagIDStr := r.URL.Query().Get("tag")
		var medias []Media

		if tagIDStr != "" {
			tagID, err := strconv.ParseUint(tagIDStr, 10, 32)
			if err != nil {
				http.Error(w, "Invalid tag ID", http.StatusBadRequest)
				return
			}

			// Fetch media associated with the specified tag ID
			if result := db.Joins("JOIN media_tags ON media_tags.media_id = media.id").
				Where("media_tags.tag_id = ?", uint(tagID)).
				Preload("Tags"). // Eager load the tags associated with the media
				Find(&medias); result.Error != nil {
				http.Error(w, "Failed to fetch media", http.StatusInternalServerError)
				return
			}
		} else {

			// Fetch all media if no tag is specified
			if result := db.Preload("Tags").Find(&medias); result.Error != nil {
				http.Error(w, "Failed to fetch media", http.StatusInternalServerError)
				return
			}
		}
		// Encode media as JSON
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(medias); err != nil {
			http.Error(w, "Failed to encode media", http.StatusInternalServerError)
			return
		}

	case http.MethodPost:
		// Parse the multipart form (set a 10MB limit for files)
		err := r.ParseMultipartForm(10 << 20) // 10MB
		if err != nil {
			http.Error(w, "File too large or invalid input", http.StatusBadRequest)
			return
		}

		// Retrieve the name from the form fields
		name := r.FormValue("Name")
		if name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}

		// Retrieve the file from the form
		file, fileHeader, err := r.FormFile("File")
		if err != nil {
			http.Error(w, "File is required", http.StatusBadRequest)
			return
		}
		defer file.Close()

		orgFilename := fileHeader.Filename

		// Clean the name to create a valid filename
		cleanName := sanitizeString(name)
		// Generate a unique filename hash based on the name and the current time
		filename := generateUniqueFilename(cleanName) + "_" + sanitizeString(orgFilename)

		// Save the file to disk
		filePath, err := saveFileToDisk(file, filename)
		if err != nil {
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}

		// Extract tags from the form and parse them
		tagsString := r.FormValue("Tags")
		var tags []Tag
		err = json.Unmarshal([]byte(tagsString), &tags)
		if err != nil {
			http.Error(w, "Invalid tags format", http.StatusBadRequest)
			return
		}

		// Ensure tags exist in the database and create them if necessary
		var dbTags []*Tag
		for _, tag := range tags {
			var existingTag Tag
			if err := db.Where("name = ?", tag.Name).First(&existingTag).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					// If the tag doesn't exist, create it
					newTag := Tag{Name: tag.Name}
					db.Create(&newTag)
					dbTags = append(dbTags, &newTag)
				} else {
					http.Error(w, "Error processing tags", http.StatusInternalServerError)
					return
				}
			} else {
				dbTags = append(dbTags, &existingTag)
			}
		}

		// Create the media record
		newMedia := Media{
			Name: name,
			URL:  filePath, // Store the file path (or URL if needed)
			Tags: dbTags,
		}

		// Save media details (without the file content, only file path) to the database
		if err := db.Create(&newMedia).Error; err != nil {
			http.Error(w, "Error saving media", http.StatusInternalServerError)
			return
		}

		// Respond with the created media (excluding the file content)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(newMedia); err != nil {
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			return
		}

	case http.MethodOptions:
		w.Header().Set("Allow", "GET, POST, OPTIONS")
		w.WriteHeader(http.StatusNoContent)

	default:
		w.Header().Set("Allow", "GET, POST, OPTIONS")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// sanitizeFilename removes all non-ASCII characters, replaces spaces with underscores, and removes any non-alphanumeric characters
func sanitizeString(name string) string {
	// Replace spaces with underscores
	name = regexp.MustCompile(`\s+`).ReplaceAllString(name, "_")
	// Remove all non-ASCII and non-alphanumeric characters
	name = regexp.MustCompile(`[^a-zA-Z0-9_.-]`).ReplaceAllString(name, "")

	return name
}

// generateUniqueFilename creates a unique hash-based filename using SHA256 and the current timestamp
func generateUniqueFilename(name string) string {
	// Create a unique hash using the name and the current timestamp
	hash := sha256.New()
	hash.Write([]byte(name + time.Now().String()))
	return hex.EncodeToString(hash.Sum(nil))
}

// saveFileToDisk writes the file to disk with the given filename and returns the file path
func saveFileToDisk(file io.Reader, filename string) (string, error) {
	// Define the directory to save the file
	uploadDir := "../static/uploads"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return "", err
	}

	// Construct the full file path (you can add a file extension if needed)
	filePath := filepath.Join(uploadDir, filename)

	// Create the file on disk
	outFile, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	// Copy the uploaded file data to the disk file
	if _, err := io.Copy(outFile, file); err != nil {
		return "", err
	}

	return filePath, nil
}
