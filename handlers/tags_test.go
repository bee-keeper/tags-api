package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateTag(t *testing.T) {
	db := setup()
	defer teardown(db)

	requestBody := []byte(`{"Name":"tag1"}`)
	req, err := http.NewRequest("POST", "/v1/tags", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/tags", func(w http.ResponseWriter, r *http.Request) {
		Tags(w, r, db)
	})
	mux.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	var respTag Tag
	err = json.Unmarshal(recorder.Body.Bytes(), &respTag)
	if err != nil {
		t.Fatalf("could not unmarshal response body: %v", err)
	}

	assert.Equal(t, "tag1", respTag.Name, "The 'name' field should be 'tag1'")
}

func TestInvalidTag(t *testing.T) {
	db := setup()
	defer teardown(db)

	requestBody := []byte(`{"Name":tag1"}`)
	req, err := http.NewRequest("POST", "/v1/tags", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/tags", func(w http.ResponseWriter, r *http.Request) {
		Tags(w, r, db)
	})
	mux.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}

	assert.Equal(t, "Invalid input\n", recorder.Body.String())
}

func TestInvalidRoute(t *testing.T) {
	db := setup()
	defer teardown(db)

	requestBody := []byte(`{"Name":tag1"}`)
	req, err := http.NewRequest("POST", "/v1/tags/fggffh", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/tags", func(w http.ResponseWriter, r *http.Request) {
		Tags(w, r, db)
	})
	mux.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNotFound)
	}
}

func TestDupeCreateTag(t *testing.T) {
	db := setup()
	defer teardown(db)

	result := db.Create(&Tag{Name: "tag1"})
	if result.Error != nil {
		fmt.Printf("Failed to create tag: %v\n", result.Error)
	}

	requestBody := []byte(`{"Name":"tag1"}`)
	req, err := http.NewRequest("POST", "/v1/tags", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/tags", func(w http.ResponseWriter, r *http.Request) {
		Tags(w, r, db)
	})
	mux.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusConflict {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusConflict)
	}
	assert.Equal(t, "Tag with this name already exists\n", recorder.Body.String())
}

func TestListTags(t *testing.T) {
	db := setup()
	defer teardown(db)

	result := db.Create([]Tag{
		{Name: "tag1"},
		{Name: "tag2"},
		{Name: "tag3"},
	})
	if result.Error != nil {
		fmt.Printf("Failed to create tags: %v\n", result.Error)
	}

	req, err := http.NewRequest("GET", "/v1/tags", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/tags", func(w http.ResponseWriter, r *http.Request) {
		Tags(w, r, db)
	})
	mux.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var tags []Tag
	err = json.Unmarshal(recorder.Body.Bytes(), &tags)
	if err != nil {
		t.Fatalf("could not unmarshal response body: %v", err)
	}

	expectedNames := map[string]bool{
		"tag1": false,
		"tag2": false,
		"tag3": false,
	}

	for _, tag := range tags {
		if _, exists := expectedNames[tag.Name]; exists {
			expectedNames[tag.Name] = true
		}
	}

	for name, found := range expectedNames {
		if !found {
			t.Fatalf("Expected tag %s not found", name)
		}
	}
}

func TestOptionsHandler(t *testing.T) {
	db := setup()
	defer teardown(db)

	req, err := http.NewRequest(http.MethodOptions, "/v1/tags", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/tags", func(w http.ResponseWriter, r *http.Request) {
		Tags(w, r, db)
	})
	mux.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNoContent, recorder.Code)
	allowHeader := recorder.Header().Get("Allow")
	assert.Equal(t, "GET, POST, OPTIONS", allowHeader)
}

func TestDefaultHandlerMethod(t *testing.T) {
	db := setup()
	defer teardown(db)

	req, err := http.NewRequest(http.MethodPut, "/v1/tags", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/tags", func(w http.ResponseWriter, r *http.Request) {
		Tags(w, r, db)
	})
	mux.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusMethodNotAllowed, recorder.Code)
	allowHeader := recorder.Header().Get("Allow")
	assert.Equal(t, "GET, POST, OPTIONS", allowHeader)
	assert.Contains(t, recorder.Body.String(), "method not allowed")
}
