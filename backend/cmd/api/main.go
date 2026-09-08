package main

import (
	"database/sql"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/routers"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://app:app@localhost:5432/app"
	}
	databaseURL = withSSLMode(databaseURL, os.Getenv("DATABASE_SSL_MODE"))
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	addr := ":" + port
	log.Printf("backend server listening on %s", addr)
	if err := http.ListenAndServe(addr, routers.NewRouter(db)); err != nil {
		log.Fatal(err)
	}
}

func withSSLMode(databaseURL, sslMode string) string {
	if sslMode == "" {
		sslMode = "disable"
	}

	parsedURL, err := url.Parse(databaseURL)
	if err != nil {
		log.Fatalf("invalid DATABASE_URL: %v", err)
	}
	query := parsedURL.Query()
	query.Set("sslmode", sslMode)
	parsedURL.RawQuery = query.Encode()
	return parsedURL.String()
}
