package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateMedia(t *testing.T) {
	db := setup()
	defer teardown(db)

	recorder := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/media", func(w http.ResponseWriter, r *http.Request) {
		AllMedia(w, r, db)
	})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("Name", "media1")
	_ = writer.WriteField("Tags", `[{"Name":"tag1"}, {"Name":"tag2"}]`)
	file, err := os.Open("../static/tests/bg.png")
	if err != nil {
		t.Fatalf("could not open test file: %v", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("File", filepath.Base(file.Name()))
	if err != nil {
		t.Fatalf("could not create form file: %v", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		t.Fatalf("could not copy file content: %v", err)
	}
	writer.Close()

	req := httptest.NewRequest("POST", "/v1/media", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	mux.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	var respMedia Media
	err = json.Unmarshal(recorder.Body.Bytes(), &respMedia)
	if err != nil {
		t.Fatalf("could not unmarshal response body: %v", err)
	}

	assert.Equal(t, "media1", respMedia.Name, "The 'name' field should be 'media1'")
	pattern := `^../static/uploads/.+_bg\.png$`
	re, err := regexp.Compile(pattern)
	if err != nil {
		t.Fatalf("Failed to compile regex: %v", err)
	}
	matched := re.MatchString(respMedia.URL)
	assert.True(t, matched, "The 'URL' field should match the pattern '^../static/uploads/.+_bg\\.png$'")

	expectedNames := map[string]bool{
		"tag1": false,
		"tag2": false,
	}

	for _, tag := range respMedia.Tags {
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

func TestGetMediaByTag(t *testing.T) {
	db := setup()
	defer teardown(db)

	var tag1, tag2, tag3 Tag
	db.FirstOrCreate(&tag1, Tag{Name: "tag1"})
	db.FirstOrCreate(&tag2, Tag{Name: "tag2"})
	db.FirstOrCreate(&tag3, Tag{Name: "tag3"})

	result := db.Create([]Media{
		{Name: "media1", URL: "../static/uploads/fgggff5fgf_bg.png", Tags: []*Tag{&tag2}},
		{Name: "media2", URL: "../static/uploads/fgggttggf_bg.png", Tags: []*Tag{&tag3}},
		{Name: "media3", URL: "../static/uploads/fgeerggf_bg.png", Tags: []*Tag{&tag3, &tag1}},
	})
	if result.Error != nil {
		fmt.Printf("Failed to create media: %v\n", result.Error)
	}

	recorder := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/media", func(w http.ResponseWriter, r *http.Request) {
		AllMedia(w, r, db)
	})

	req := httptest.NewRequest("GET", "/v1/media?tag="+strconv.FormatUint(uint64(tag2.ID), 10), nil)
	mux.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// make sure only media1 is returned

	var mediaResp []Media
	err := json.Unmarshal(recorder.Body.Bytes(), &mediaResp)
	if err != nil {
		t.Fatalf("could not unmarshal response body: %v", err)
	}

	assert.Len(t, mediaResp, 1, "Expected mediaResp to contain exactly 1 records")
	assert.Equal(t, mediaResp[0].Name, "media1", "Expected media to be called media1")

	assert.Len(t, mediaResp[0].Tags, 1, "Expected mediaResp to contain exactly 1 tag")
	assert.Equal(t, mediaResp[0].Tags[0].Name, "tag2", "Expected tag to be tag2")

}

func TestListMedia(t *testing.T) {
	db := setup()
	defer teardown(db)

	var tag1 Tag
	db.FirstOrCreate(&tag1, Tag{Name: "tag1"})

	result := db.Create([]Media{
		{Name: "media1", URL: "../static/uploads/fgggff5fgf_bg.png", Tags: []*Tag{&tag1}},
	})
	if result.Error != nil {
		fmt.Printf("Failed to create media: %v\n", result.Error)
	}

	recorder := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/media", func(w http.ResponseWriter, r *http.Request) {
		AllMedia(w, r, db)
	})

	req := httptest.NewRequest("GET", "/v1/media", nil)
	mux.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var mediaResp []Media
	err := json.Unmarshal(recorder.Body.Bytes(), &mediaResp)
	if err != nil {
		t.Fatalf("could not unmarshal response body: %v", err)
	}
	assert.Len(t, mediaResp, 1, "Expected mediaResp to contain exactly 1 record")
	assert.Equal(t, "media1", mediaResp[0].Name, "The 'name' field should be 'media1'")
}

func TestGetMediaByNonExistantTag(t *testing.T) {
	db := setup()
	defer teardown(db)

	var tag1, tag2, tag3 Tag
	db.FirstOrCreate(&tag1, Tag{Name: "tag1"})
	db.FirstOrCreate(&tag2, Tag{Name: "tag2"})
	db.FirstOrCreate(&tag3, Tag{Name: "tag3"})

	result := db.Create([]Media{
		{Name: "media1", URL: "../static/uploads/fgggff5fgf_bg.png", Tags: []*Tag{&tag1, &tag2, &tag3}},
	})
	if result.Error != nil {
		fmt.Printf("Failed to create media: %v\n", result.Error)
	}

	recorder := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/media", func(w http.ResponseWriter, r *http.Request) {
		AllMedia(w, r, db)
	})

	req := httptest.NewRequest("GET", "/v1/media?tag=909345", nil)
	mux.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var mediaResp []Media
	err := json.Unmarshal(recorder.Body.Bytes(), &mediaResp)
	if err != nil {
		t.Fatalf("could not unmarshal response body: %v", err)
	}
	assert.Len(t, mediaResp, 0, "Expected mediaResp to be empty")
}

func TestMediaOptionsHandler(t *testing.T) {
	db := setup()
	defer teardown(db)

	req, err := http.NewRequest(http.MethodOptions, "/v1/media", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/media", func(w http.ResponseWriter, r *http.Request) {
		AllMedia(w, r, db)
	})
	mux.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNoContent, recorder.Code)
	allowHeader := recorder.Header().Get("Allow")
	assert.Equal(t, "GET, POST, OPTIONS", allowHeader)
}

func TestMediaDefaultHandlerMethod(t *testing.T) {
	db := setup()
	defer teardown(db)

	req, err := http.NewRequest(http.MethodPut, "/v1/media", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/media", func(w http.ResponseWriter, r *http.Request) {
		AllMedia(w, r, db)
	})
	mux.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusMethodNotAllowed, recorder.Code)
	allowHeader := recorder.Header().Get("Allow")
	assert.Equal(t, "GET, POST, OPTIONS", allowHeader)
	assert.Contains(t, recorder.Body.String(), "method not allowed")
}
