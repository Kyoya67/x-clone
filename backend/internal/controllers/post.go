package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/apperrors"
	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/models"
)

const fixedAuthorID = "00000000-0000-0000-0000-000000000001"

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
	if err := decoder.Decode(&request); err != nil {
		apperrors.ErrorHandler(w, r, apperrors.ReqBodyDecodeFailed.Wrap(err, "request body must be valid JSON"))
		return
	}

	post, err := c.service.Create(r.Context(), fixedAuthorID, request)
	if err != nil {
		apperrors.ErrorHandler(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(post)
}
