package routers

import (
	"database/sql"
	"net/http"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/controllers"
	"github.com/gorilla/mux"
)

func NewRouter(db *sql.DB) http.Handler {
	_ = db
	healthController := controllers.NewHealthController()

	r := mux.NewRouter()
	r.HandleFunc("/health", healthController.Health).Methods(http.MethodGet)
	r.HandleFunc("/docs", controllers.SwaggerUI).Methods(http.MethodGet)
	r.HandleFunc("/openapi.yaml", controllers.OpenAPISpec).Methods(http.MethodGet)

	return r
}
