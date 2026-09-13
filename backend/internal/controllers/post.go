package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/apperrors"
	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/auth"
	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/models"
)

type PostController struct {
	service PostService
}

type PostService interface {
	Create(ctx context.Context, authorID string, request models.CreatePostRequest) (models.Post, error)
}

func NewPostController(service PostService) *PostController {
	return &PostController{service: service}
}

func (c *PostController) Create(w http.ResponseWriter, r *http.Request) {
	var request models.CreatePostRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		errCode := apperrors.ReqBodyDecodeFailed
		message := "request body must be valid JSON"
		if strings.HasPrefix(err.Error(), "json: unknown field ") {
			errCode = apperrors.UnknownField
			message = "request contains an unknown field"
		}
		apperrors.ErrorHandler(w, r, errCode.Wrap(err, message))
		return
	}

	authorID, err := auth.UserID(r.Context())
	if err != nil {
		apperrors.ErrorHandler(w, r, apperrors.Unauthorized.Wrap(err, "login is required"))
		return
	}

	post, err := c.service.Create(r.Context(), authorID, request)
	if err != nil {
		apperrors.ErrorHandler(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(post)
}
