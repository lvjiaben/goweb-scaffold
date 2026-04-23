package admin_auth

import "time"

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

type LoginResult struct {
	Token     string         `json:"token"`
	ExpiresAt time.Time      `json:"expires_at"`
	User      LoginUserBrief `json:"user"`
}

type LoginUserBrief struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	IsSuper  bool   `json:"is_super"`
}

type CurrentAdmin struct {
	ID       int64
	Username string
	Nickname string
	IsSuper  bool
}

type MeResponse struct {
	ID          int64    `json:"id"`
	Username    string   `json:"username"`
	Nickname    string   `json:"nickname"`
	IsSuper     bool     `json:"is_super"`
	RoleIDs     []int64  `json:"role_ids"`
	AccessCodes []string `json:"access_codes"`
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
