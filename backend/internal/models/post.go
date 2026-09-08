package models

import "time"

type Post struct {
	ID        string    `json:"id"`
	AuthorID  string    `json:"authorId"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

type CreatePostRequest struct {
	Content string `json:"content"`
}
