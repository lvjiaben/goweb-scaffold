package app_user

import "time"

type CurrentUser struct {
	ID           int64
	Username     string
	PasswordHash string
	Nickname     string
	Email        string
	Mobile       string
	Status       int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type ProfileSaveRequest struct {
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Mobile   string `json:"mobile"`
}

type PasswordChangeRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type ProfileResponse struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Nickname  string    `json:"nickname"`
	Email     string    `json:"email"`
	Mobile    string    `json:"mobile"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SaveResult struct {
	ID int64 `json:"id"`
}

type PasswordChangeResult struct {
	Changed bool `json:"changed"`
}
