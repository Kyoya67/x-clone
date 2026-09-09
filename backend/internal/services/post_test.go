package services

import (
	"context"
	"errors"
	"testing"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/models"
)

type fakePostRepository struct {
	called   bool
	authorID string
	content  string
	post     models.Post
	err      error
}

func (f *fakePostRepository) Create(_ context.Context, authorID, content string) (models.Post, error) {
	f.called = true
	f.authorID = authorID
	f.content = content
	return f.post, f.err
}

func TestPostServiceCreate(t *testing.T) {
	repository := &fakePostRepository{post: models.Post{ID: "post-1"}}
	service := NewPostService(repository)

	post, err := service.Create(context.Background(), "author-1", models.CreatePostRequest{Content: "  hello  "})
	if err != nil {
		t.Fatal(err)
	}
	if !repository.called || repository.authorID != "author-1" || repository.content != "hello" {
		t.Fatalf("unexpected repository call: %+v", repository)
	}
	if post.ID != "post-1" {
		t.Fatalf("unexpected post: %+v", post)
	}
}

func TestPostServiceRejectsInvalidContent(t *testing.T) {
	repository := &fakePostRepository{}
	service := NewPostService(repository)

	_, err := service.Create(context.Background(), "author-1", models.CreatePostRequest{Content: "   "})
	if !errors.Is(err, ErrInvalidPostContent) {
		t.Fatalf("expected invalid content error, got %v", err)
	}
	if repository.called {
		t.Fatal("repository should not be called")
	}
}
