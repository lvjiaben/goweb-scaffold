package app_user

import (
	"errors"
	"time"

	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
	sharedquery "github.com/lvjiaben/goweb-scaffold/internal/shared/query"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(runtime *bootstrap.Runtime) *Repo {
	return &Repo{db: runtime.DB}
}

func (r *Repo) List(params ListParams) ([]AppUser, int64, error) {
	db := r.db.Model(&AppUser{}).Where("deleted_at IS NULL")
	filters := map[string]any{}
	for key, value := range params.Filters {
		filters[key] = value
	}
	if params.Keyword != "" {
		filters["keyword"] = params.Keyword
	}
	queryResult := sharedquery.Apply(db, sharedquery.Params{
		Page:      params.Page,
		PageSize:  params.PageSize,
		Search:    params.Keyword,
		Filters:   filters,
		SortBy:    params.SortBy,
		SortOrder: params.SortOrder,
	}, sharedquery.Options{
		SearchFields: []string{"username", "nickname", "email", "mobile", "code"},
		LikeFields:   []string{"username", "nickname", "email", "mobile", "code"},
		ExactFields:  []string{"id", "pid", "tid", "status"},
		RangeFields:  []string{"created_at"},
		AllowedSorts: []string{"id", "pid", "tid", "status", "money", "score", "created_at", "updated_at"},
		DefaultSorts: []string{"id DESC"},
	})

	var total int64
	if err := queryResult.Count.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []AppUser
	if err := queryResult.Query.Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *Repo) FindByID(id int64) (AppUser, error) {
	var user AppUser
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&user).Error
	return user, err
}

func (r *Repo) FindByUsername(username string) (AppUser, error) {
	var user AppUser
	err := r.db.Where("username = ? AND deleted_at IS NULL", username).First(&user).Error
	return user, err
}

func (r *Repo) UsernameExists(username string, exceptID int64) (bool, error) {
	var count int64
	db := r.db.Model(&AppUser{}).Where("username = ? AND deleted_at IS NULL", username)
	if exceptID > 0 {
		db = db.Where("id <> ?", exceptID)
	}
	if err := db.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repo) Create(user *AppUser) error {
	return r.db.Create(user).Error
}

func (r *Repo) Update(id int64, updates map[string]any) error {
	return r.db.Model(&AppUser{}).Where("id = ? AND deleted_at IS NULL", id).Updates(updates).Error
}

func (r *Repo) Delete(ids []int64) (int, error) {
	result := r.db.Where("id IN ?", ids).Delete(&AppUser{})
	return int(result.RowsAffected), result.Error
}

func (r *Repo) UpdateMany(ids []int64, updates map[string]any) (int, error) {
	result := r.db.Model(&AppUser{}).Where("id IN ? AND deleted_at IS NULL", ids).Updates(updates)
	return int(result.RowsAffected), result.Error
}

func (r *Repo) CreateSession(session *AppUserSession) error {
	return r.db.Create(session).Error
}

func (r *Repo) DeleteSession(id int64) error {
	return r.db.Where("id = ?", id).Delete(&AppUserSession{}).Error
}

func (r *Repo) UpdateProfile(userID int64, updates map[string]any) error {
	return r.Update(userID, updates)
}

func (r *Repo) UpdatePassword(userID int64, passwordHash string, updatedAt time.Time) error {
	return r.Update(userID, map[string]any{
		"password_hash": passwordHash,
		"updated_at":    updatedAt,
	})
}

func (r *Repo) UpdateMoneyWithLog(userID int64, kind int, amount float64, note string, source string) (float64, float64, error) {
	var before, after float64
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var user AppUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error; err != nil {
			return err
		}
		before = user.Money
		after = before + signedAmount(kind, amount)
		if after < 0 {
			return errors.New("余额不足")
		}
		if err := tx.Model(&AppUser{}).Where("id = ?", userID).Updates(map[string]any{
			"money":      after,
			"updated_at": time.Now(),
		}).Error; err != nil {
			return err
		}
		return tx.Create(&UserMoneyLog{
			UserID:      userID,
			Type:        kind,
			Money:       amount,
			BeforeMoney: before,
			AfterMoney:  after,
			Note:        note,
			Source:      source,
			CreatedAt:   time.Now(),
		}).Error
	})
	return before, after, err
}

func (r *Repo) UpdateScoreWithLog(userID int64, kind int, amount float64, note string, source string) (float64, float64, error) {
	var before, after float64
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var user AppUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error; err != nil {
			return err
		}
		before = user.Score
		after = before + signedAmount(kind, amount)
		if after < 0 {
			return errors.New("积分不足")
		}
		if err := tx.Model(&AppUser{}).Where("id = ?", userID).Updates(map[string]any{
			"score":      after,
			"updated_at": time.Now(),
		}).Error; err != nil {
			return err
		}
		return tx.Create(&UserScoreLog{
			UserID:      userID,
			Type:        kind,
			Score:       amount,
			BeforeScore: before,
			AfterScore:  after,
			Note:        note,
			Source:      source,
			CreatedAt:   time.Now(),
		}).Error
	})
	return before, after, err
}

func (r *Repo) MoneyLogs(params LogListParams) ([]UserMoneyLog, int64, error) {
	db := r.db.Model(&UserMoneyLog{})
	if params.UserID > 0 {
		db = db.Where("user_id = ?", params.UserID)
	}
	var total int64
	if err := db.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []UserMoneyLog
	err := db.Order("id DESC").Offset((params.Page - 1) * params.PageSize).Limit(params.PageSize).Find(&list).Error
	return list, total, err
}

func (r *Repo) ScoreLogs(params LogListParams) ([]UserScoreLog, int64, error) {
	db := r.db.Model(&UserScoreLog{})
	if params.UserID > 0 {
		db = db.Where("user_id = ?", params.UserID)
	}
	var total int64
	if err := db.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []UserScoreLog
	err := db.Order("id DESC").Offset((params.Page - 1) * params.PageSize).Limit(params.PageSize).Find(&list).Error
	return list, total, err
}

func signedAmount(kind int, amount float64) float64 {
	if kind == 2 {
		return -amount
	}
	return amount
}
