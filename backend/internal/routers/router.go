package routers

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/auth"
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
	timelineRepository := repositories.NewTimelineRepository(db)
	timelineService := services.NewTimelineService(timelineRepository)
	timelineController := controllers.NewTimelineController(timelineService)
	notificationRepository := repositories.NewNotificationRepository(db)
	notificationService := services.NewNotificationService(notificationRepository)
	notificationController := controllers.NewNotificationController(notificationService)
	followRepository := repositories.NewFollowRepository(db)
	followService := services.NewFollowService(followRepository, notificationService)
	followController := controllers.NewFollowController(followService)
	likeRepository := repositories.NewLikeRepository(db)
	likeService := services.NewLikeService(likeRepository, notificationService)
	likeController := controllers.NewLikeController(likeService)
	userRepository := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepository)
	authController := controllers.NewAuthController(authConfigFromEnv(), userService)

	r := mux.NewRouter()
	r.HandleFunc("/health", healthController.Health).Methods(http.MethodGet)
	r.HandleFunc("/docs", controllers.SwaggerUI).Methods(http.MethodGet)
	r.HandleFunc("/openapi.yaml", controllers.OpenAPISpec).Methods(http.MethodGet)
	r.HandleFunc("/auth/login", authController.Login).Methods(http.MethodGet)
	r.HandleFunc("/auth/callback", authController.Callback).Methods(http.MethodGet)
	r.HandleFunc("/auth/logout", authController.Logout).Methods(http.MethodPost)

	protected := r.NewRoute().Subrouter()
	protected.Use(authController.Middleware)
	protected.HandleFunc("/auth/me", authController.Me).Methods(http.MethodGet)
	protected.HandleFunc("/auth/me", authController.UpdateMe).Methods(http.MethodPatch)
	protected.HandleFunc("/posts", postController.Create).Methods(http.MethodPost)
	protected.HandleFunc("/timeline", timelineController.List).Methods(http.MethodGet)
	protected.HandleFunc("/notifications", notificationController.List).Methods(http.MethodGet)
	protected.HandleFunc("/me/following", followController.ListFollowing).Methods(http.MethodGet)
	protected.HandleFunc("/users/{userId}/follow", followController.Follow).Methods(http.MethodPut)
	protected.HandleFunc("/users/{userId}/follow", followController.Unfollow).Methods(http.MethodDelete)
	protected.HandleFunc("/posts/{postId}/like", likeController.Like).Methods(http.MethodPut)
	protected.HandleFunc("/posts/{postId}/like", likeController.Unlike).Methods(http.MethodDelete)

	return r
}

func authConfigFromEnv() controllers.AuthConfig {
	secureCookie := os.Getenv("AUTH_COOKIE_SECURE") != "false"
	return controllers.AuthConfig{
		Issuer:        os.Getenv("AUTH_ISSUER"),
		AuthorizeURL:  os.Getenv("AUTH_AUTHORIZE_URL"),
		TokenURL:      os.Getenv("AUTH_TOKEN_URL"),
		ClientID:      os.Getenv("AUTH_CLIENT_ID"),
		ClientSecret:  os.Getenv("AUTH_CLIENT_SECRET"),
		RedirectURL:   os.Getenv("AUTH_REDIRECT_URL"),
		PostLoginURL:  envOrDefault("AUTH_POST_LOGIN_URL", "/"),
		SessionSecret: os.Getenv("AUTH_SESSION_SECRET"),
		Cookie: auth.CookieConfig{
			Domain:     os.Getenv("AUTH_COOKIE_DOMAIN"),
			Secure:     secureCookie,
			NamePrefix: os.Getenv("AUTH_COOKIE_NAME_PREFIX"),
		},
	}
}

func envOrDefault(name, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	return value
}
