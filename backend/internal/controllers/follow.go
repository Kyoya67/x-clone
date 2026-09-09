package controllers

import (
	"context"
	"net/http"
	"regexp"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/apperrors"
	"github.com/gorilla/mux"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type FollowController struct {
	service FollowService
}

type FollowService interface {
	Follow(ctx context.Context, followerID, followeeID string) error
	Unfollow(ctx context.Context, followerID, followeeID string) error
}

type followAction func(ctx context.Context, followerID, followeeID string) error

func NewFollowController(service FollowService) *FollowController {
	return &FollowController{service: service}
}

func (c *FollowController) Follow(w http.ResponseWriter, r *http.Request) {
	c.handleFollowAction(w, r, c.service.Follow)
}

func (c *FollowController) Unfollow(w http.ResponseWriter, r *http.Request) {
	c.handleFollowAction(w, r, c.service.Unfollow)
}

func (c *FollowController) handleFollowAction(w http.ResponseWriter, r *http.Request, action followAction) {
	followeeID, ok := followeeIDFromRequest(w, r)
	if !ok {
		return
	}

	if err := action(r.Context(), fixedAuthorID, followeeID); err != nil {
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
