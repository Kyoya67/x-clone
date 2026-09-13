package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/apperrors"
	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/auth"
	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/models"
)

type TimelineService interface {
	List(ctx context.Context, userID string, feed models.TimelineFeed) ([]models.TimelinePost, error)
}

type TimelineController struct {
	service TimelineService
}

func NewTimelineController(service TimelineService) *TimelineController {
	return &TimelineController{service: service}
}

func (c *TimelineController) List(w http.ResponseWriter, r *http.Request) {
	feed := models.TimelineFeed(r.URL.Query().Get("feed"))
	if feed == "" {
		feed = models.TimelineFeedForYou
	}

	userID, err := auth.UserID(r.Context())
	if err != nil {
		apperrors.ErrorHandler(w, r, apperrors.Unauthorized.Wrap(err, "login is required"))
		return
	}

	posts, err := c.service.List(r.Context(), userID, feed)
	if err != nil {
		apperrors.ErrorHandler(w, r, err)
		return
	}

	response := models.TimelineResponse{Posts: posts}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		apperrors.ErrorHandler(w, r, err)
	}
}
