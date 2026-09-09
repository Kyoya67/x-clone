package routers

import (
	"database/sql"
	"net/http"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/controllers"
	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/repositories"
	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/services"
	"github.com/gorilla/mux"
)

func NewRouter(db *sql.DB) http.Handler {
	_ = db
	healthController := controllers.NewHealthController()
	postRepository := repositories.NewPostRepository(db)
	postService := services.NewPostService(postRepository)
	postController := controllers.NewPostController(postService)
	followRepository := repositories.NewFollowRepository(db)
	followService := services.NewFollowService(followRepository)
	followController := controllers.NewFollowController(followService)

	r := mux.NewRouter()
	r.HandleFunc("/health", healthController.Health).Methods(http.MethodGet)
	r.HandleFunc("/docs", controllers.SwaggerUI).Methods(http.MethodGet)
	r.HandleFunc("/openapi.yaml", controllers.OpenAPISpec).Methods(http.MethodGet)
	r.HandleFunc("/posts", postController.Create).Methods(http.MethodPost)
	r.HandleFunc("/me/following", followController.ListFollowing).Methods(http.MethodGet)
	r.HandleFunc("/users/{userId}/follow", followController.Follow).Methods(http.MethodPut)
	r.HandleFunc("/users/{userId}/follow", followController.Unfollow).Methods(http.MethodDelete)

	return r
}
