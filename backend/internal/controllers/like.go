package controllers

import (
	"context"
	"net/http"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/apperrors"
	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/auth"
	"github.com/gorilla/mux"
)

type LikeController struct {
	service LikeService
}

type LikeService interface {
	Like(ctx context.Context, userID, postID string) error
	Unlike(ctx context.Context, userID, postID string) error
}

type likeAction func(ctx context.Context, userID, postID string) error

func NewLikeController(service LikeService) *LikeController {
	return &LikeController{service: service}
}

func (c *LikeController) Like(w http.ResponseWriter, r *http.Request) {
	c.handleLikeAction(w, r, c.service.Like)
}

func (c *LikeController) Unlike(w http.ResponseWriter, r *http.Request) {
	c.handleLikeAction(w, r, c.service.Unlike)
}

func (c *LikeController) handleLikeAction(w http.ResponseWriter, r *http.Request, action likeAction) {
	postID := mux.Vars(r)["postId"]
	if !uuidPattern.MatchString(postID) {
		apperrors.ErrorHandler(w, r, apperrors.BadParam.Wrap(nil, "postId must be a UUID"))
		return
	}

	userID, err := auth.UserID(r.Context())
	if err != nil {
		apperrors.ErrorHandler(w, r, apperrors.Unauthorized.Wrap(err, "login is required"))
		return
	}

	if err := action(r.Context(), userID, postID); err != nil {
		apperrors.ErrorHandler(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
