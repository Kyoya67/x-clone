package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/apperrors"
	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/auth"
	"github.com/gorilla/mux"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type FollowController struct {
	service FollowService
}

type FollowService interface {
	Follow(ctx context.Context, followerID, followeeID string) error
	Unfollow(ctx context.Context, followerID, followeeID string) error
	ListFolloweeIDs(ctx context.Context, followerID string) ([]string, error)
}

type followAction func(ctx context.Context, followerID, followeeID string) error

type followingResponse struct {
	UserIDs []string `json:"userIds"`
}

func NewFollowController(service FollowService) *FollowController {
	return &FollowController{service: service}
}

func (c *FollowController) Follow(w http.ResponseWriter, r *http.Request) {
	c.handleFollowAction(w, r, c.service.Follow)
}

func (c *FollowController) Unfollow(w http.ResponseWriter, r *http.Request) {
	c.handleFollowAction(w, r, c.service.Unfollow)
}

func (c *FollowController) ListFollowing(w http.ResponseWriter, r *http.Request) {
	followerID, err := auth.UserID(r.Context())
	if err != nil {
		apperrors.ErrorHandler(w, r, apperrors.Unauthorized.Wrap(err, "login is required"))
		return
	}
	followeeIDs, err := c.service.ListFolloweeIDs(r.Context(), followerID)
	if err != nil {
		apperrors.ErrorHandler(w, r, err)
		return
	}

	response := followingResponse{UserIDs: followeeIDs}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		apperrors.ErrorHandler(w, r, err)
	}
}

func (c *FollowController) handleFollowAction(w http.ResponseWriter, r *http.Request, action followAction) {
	followeeID, ok := followeeIDFromRequest(w, r)
	if !ok {
		return
	}

	followerID, err := auth.UserID(r.Context())
	if err != nil {
		apperrors.ErrorHandler(w, r, apperrors.Unauthorized.Wrap(err, "login is required"))
		return
	}

	if err := action(r.Context(), followerID, followeeID); err != nil {
		apperrors.ErrorHandler(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func followeeIDFromRequest(w http.ResponseWriter, r *http.Request) (string, bool) {
	followeeID := mux.Vars(r)["userId"]
	if !uuidPattern.MatchString(followeeID) {
		apperrors.ErrorHandler(w, r, apperrors.BadParam.Wrap(nil, "userId must be a UUID"))
		return "", false
	}
	return followeeID, true
}
