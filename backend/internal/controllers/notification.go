package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/apperrors"
	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/auth"
	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/models"
)

type NotificationController struct {
	service NotificationService
}

type NotificationService interface {
	List(ctx context.Context, recipientID string) ([]models.Notification, error)
}

func NewNotificationController(service NotificationService) *NotificationController {
	return &NotificationController{service: service}
}

func (c *NotificationController) List(w http.ResponseWriter, r *http.Request) {
	recipientID, err := auth.UserID(r.Context())
	if err != nil {
		apperrors.ErrorHandler(w, r, apperrors.Unauthorized.Wrap(err, "login is required"))
		return
	}

	notifications, err := c.service.List(r.Context(), recipientID)
	if err != nil {
		apperrors.ErrorHandler(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(models.NotificationsResponse{Notifications: notifications}); err != nil {
		apperrors.ErrorHandler(w, r, err)
	}
}
