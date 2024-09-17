package main

import (
	"log"
	"net/http"

	"github.com/bee-keeper/tags-api/handlers"
	"github.com/bee-keeper/tags-api/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(postgres.Open(utils.GetDSN()), &gorm.Config{})

	db.AutoMigrate(&handlers.Tag{}, &handlers.Media{})

	if err != nil {
		panic("DB connection failed")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1", func(w http.ResponseWriter, r *http.Request) { handlers.Index(w, r, db) })
	mux.HandleFunc("/v1/tags", func(w http.ResponseWriter, r *http.Request) { handlers.Tags(w, r, db) })
	mux.HandleFunc("/v1/media", func(w http.ResponseWriter, r *http.Request) { handlers.AllMedia(w, r, db) })

	apiErr := http.ListenAndServe(":8080", mux)
	log.Fatal(apiErr)

}
