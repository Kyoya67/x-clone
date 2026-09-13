package models

import "time"

type User struct {
	ID                string    `json:"id"`
	Handle            string    `json:"handle"`
	DisplayName       string    `json:"displayName"`
	Bio               string    `json:"bio"`
	CreatedAt         time.Time `json:"createdAt"`
	NeedsProfileSetup bool      `json:"needsProfileSetup"`
}

type UpdateUserProfileRequest struct {
	Handle      string `json:"handle"`
	DisplayName string `json:"displayName"`
	Bio         string `json:"bio"`
}
