package app_user

import (
	"strings"
	"time"

	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo *Repo
}

func NewService(runtime *bootstrap.Runtime) *Service {
	return &Service{repo: NewRepo(runtime)}
}

func (s *Service) Profile(user CurrentUser) ProfileResponse {
	return ProfileResponse{
		ID:        user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Email:     user.Email,
		Mobile:    user.Mobile,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func (s *Service) SaveProfile(userID int64, req ProfileSaveRequest) (SaveResult, error) {
	req.Nickname = strings.TrimSpace(req.Nickname)
	req.Email = strings.TrimSpace(req.Email)
	req.Mobile = strings.TrimSpace(req.Mobile)
	if req.Nickname == "" {
		return SaveResult{}, validationError("nickname is required")
	}
	if err := s.repo.UpdateProfile(userID, map[string]any{
		"nickname":   req.Nickname,
		"email":      req.Email,
		"mobile":     req.Mobile,
		"updated_at": time.Now(),
	}); err != nil {
		return SaveResult{}, err
	}
	return SaveResult{ID: userID}, nil
}

func (s *Service) ChangePassword(user CurrentUser, req PasswordChangeRequest) (PasswordChangeResult, error) {
	req.OldPassword = strings.TrimSpace(req.OldPassword)
	req.NewPassword = strings.TrimSpace(req.NewPassword)
	if req.OldPassword == "" {
		return PasswordChangeResult{}, validationError("old_password is required")
	}
	if req.NewPassword == "" {
		return PasswordChangeResult{}, validationError("new_password is required")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)) != nil {
		return PasswordChangeResult{}, validationError("old password is incorrect")
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return PasswordChangeResult{}, err
	}
	if err := s.repo.UpdatePassword(user.ID, string(passwordHash), time.Now()); err != nil {
		return PasswordChangeResult{}, err
	}
	return PasswordChangeResult{Changed: true}, nil
}

type validationError string

func (e validationError) Error() string {
	return string(e)
}

func isValidationError(err error) bool {
	_, ok := err.(validationError)
	return ok
}
