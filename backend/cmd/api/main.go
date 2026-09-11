package main

import (
	"database/sql"
	"errors"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/routers"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL, err := databaseConnectionURL()
	if err != nil {
		log.Fatal(err)
	}
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatal("cannot initialize database connection")
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatal("database connection failed; check network, TLS and credentials")
	}

	addr := ":" + port
	log.Printf("backend server listening on %s", addr)
	log.Printf("Swagger UI available at http://localhost:%s/docs", port)
	if err := http.ListenAndServe(addr, routers.NewRouter(db)); err != nil {
		log.Fatal(err)
	}
}

// ECSではDB_USER/DB_PASSWORDをSecretのJSONキーから注入する。
// ローカル開発の既存DATABASE_URLも引き続き利用できる。
func databaseConnectionURL() (string, error) {
	host, user, password := os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD")
	if host != "" || user != "" || password != "" {
		port, err := strconv.Atoi(os.Getenv("DB_PORT"))
		if host == "" || user == "" || password == "" || err != nil || port < 1 || port > 65535 {
			return "", errors.New("DB_HOST, DB_PORT, DB_USER and DB_PASSWORD are required")
		}
		u := url.URL{Scheme: "postgres", Host: net.JoinHostPort(host, strconv.Itoa(port)), Path: "/app", User: url.UserPassword(user, password)}
		query := url.Values{"sslmode": {"verify-full"}, "sslrootcert": {"/app/certs/rds-ca-bundle.pem"}, "connect_timeout": {"10"}}
		u.RawQuery = query.Encode()
		return u.String(), nil
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://app:app@localhost:5432/app"
	}
	return withSSLMode(databaseURL, os.Getenv("DATABASE_SSL_MODE"))
}

func withSSLMode(databaseURL, sslMode string) (string, error) {
	if sslMode == "" {
		sslMode = "disable"
	}

	parsedURL, err := url.Parse(databaseURL)
	if err != nil {
		return "", errors.New("invalid DATABASE_URL")
	}
	query := parsedURL.Query()
	query.Set("sslmode", sslMode)
	parsedURL.RawQuery = query.Encode()
	return parsedURL.String(), nil
}
