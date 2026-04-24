package app_user

import "time"

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Mobile   string `json:"mobile"`
}

type CaptchaRequest struct {
	ID   string `json:"id"`
	Code string `json:"code"`
}

type LoginRequest struct {
	Username string         `json:"username"`
	Password string         `json:"password"`
	Captcha  CaptchaRequest `json:"captcha"`
}

type RequestMeta struct {
	IP        string
	UserAgent string
}

type RegisterResult struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
}

type LoginResult struct {
	Token       string         `json:"token"`
	AccessToken string         `json:"accessToken"`
	ExpiresAt   time.Time      `json:"expires_at"`
	User        LoginUserBrief `json:"user"`
}

type LoginUserBrief struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
}

type LogoutResult struct {
	Logout bool `json:"logout"`
}

type ProfileSaveRequest struct {
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Mobile   string `json:"mobile"`
	Avatar   string `json:"avatar"`
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
	Avatar    string    `json:"avatar"`
	Money     float64   `json:"money"`
	Score     float64   `json:"score"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PasswordChangeResult struct {
	Changed bool `json:"changed"`
}
