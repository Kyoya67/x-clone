package models

import "time"

type Follow struct {
	FollowerID string    `json:"followerId"`
	FolloweeID string    `json:"followeeId"`
	CreatedAt  time.Time `json:"createdAt"`
}
