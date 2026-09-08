package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/routers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Database initialization will be added when the database layer is introduced.
	var db *sql.DB

	addr := ":" + port
	log.Printf("backend server listening on %s", addr)
	if err := http.ListenAndServe(addr, routers.NewRouter(db)); err != nil {
		log.Fatal(err)
	}
}
