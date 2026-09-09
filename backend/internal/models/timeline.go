package models

import "time"

type TimelineFeed string

const (
	TimelineFeedForYou    TimelineFeed = "for-you"
	TimelineFeedFollowing TimelineFeed = "following"
)

type TimelineAuthor struct {
	ID          string `json:"id"`
	Handle      string `json:"handle"`
	DisplayName string `json:"displayName"`
}

type TimelinePost struct {
	ID        string         `json:"id"`
	Content   string         `json:"content"`
	CreatedAt time.Time      `json:"createdAt"`
	Author    TimelineAuthor `json:"author"`
}

type TimelineResponse struct {
	Posts []TimelinePost `json:"posts"`
}
