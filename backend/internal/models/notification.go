package models

import "time"

type NotificationType string

const (
	NotificationTypeFollow NotificationType = "follow"
	NotificationTypeLike   NotificationType = "like"
)

type NotificationActor struct {
	ID          string `json:"id"`
	Handle      string `json:"handle"`
	DisplayName string `json:"displayName"`
}

type Notification struct {
	ID        string            `json:"id"`
	Type      NotificationType  `json:"type"`
	Actor     NotificationActor `json:"actor"`
	PostID    *string           `json:"postId,omitempty"`
	CreatedAt time.Time         `json:"createdAt"`
}

type NotificationsResponse struct {
	Notifications []Notification `json:"notifications"`
}
