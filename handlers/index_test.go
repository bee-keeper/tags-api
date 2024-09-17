package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bee-keeper/tags-api/utils"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setup() *gorm.DB {
	db, err := gorm.Open(postgres.Open(utils.GetDSN()), &gorm.Config{})
	if err != nil {
		panic("Test DB connection failed")
	}
	db.AutoMigrate(&Tag{}, &Media{})
	return db
}

func teardown(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		panic("Failed to get raw DB connection")
	}
	// Clear join table
	if err := db.Exec("DELETE FROM media_tags").Error; err != nil {
		panic("Failed to delete records from join table media_tags: " + err.Error())
	}
	// Delete Media
	if err := db.Unscoped().Where("1 = 1").Delete(&Media{}).Error; err != nil {
		panic("Failed to delete records from media table: " + err.Error())
	}
	// Delete Tags
	if err := db.Unscoped().Where("1 = 1").Delete(&Tag{}).Error; err != nil {
		panic("Failed to delete records from tags table: " + err.Error())
	}

	sqlDB.Close()
}

func TestIndex(t *testing.T) {
	db := setup()
	defer teardown(db)

	req, err := http.NewRequest("GET", "/v1", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	mux := http.NewServeMux()
	mux.HandleFunc("/v1", func(w http.ResponseWriter, r *http.Request) {
		Index(w, r, db)
	})
	mux.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	assert.Equal(t, "With more time I'd return a detailed API spec here.", recorder.Body.String())
}

func TestIndexOptionsHandler(t *testing.T) {
	db := setup()
	defer teardown(db)

	req, err := http.NewRequest(http.MethodOptions, "/v1", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1", func(w http.ResponseWriter, r *http.Request) {
		Index(w, r, db)
	})
	mux.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNoContent, recorder.Code)
	allowHeader := recorder.Header().Get("Allow")
	assert.Equal(t, "GET, OPTIONS", allowHeader)
}

func TestIndexDefaultOptionsHandler(t *testing.T) {
	db := setup()
	defer teardown(db)

	req, err := http.NewRequest(http.MethodPut, "/v1", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1", func(w http.ResponseWriter, r *http.Request) {
		Index(w, r, db)
	})
	mux.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusMethodNotAllowed, recorder.Code)
	allowHeader := recorder.Header().Get("Allow")
	assert.Equal(t, "GET, OPTIONS", allowHeader)
	assert.Contains(t, recorder.Body.String(), "method not allowed")
}
