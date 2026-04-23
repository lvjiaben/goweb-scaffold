package app_user_auth

import "time"

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
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

type validationError string

func (e validationError) Error() string {
	return string(e)
}

func isValidationError(err error) bool {
	_, ok := err.(validationError)
	return ok
}
