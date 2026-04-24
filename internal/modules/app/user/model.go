package app_user

import (
	"time"

	sharedmodel "github.com/lvjiaben/goweb-scaffold/internal/shared/model"
)

type AppUser struct {
	sharedmodel.BaseModel
	PID           int64   `gorm:"column:pid" json:"pid"`
	TID           int64   `gorm:"column:tid" json:"tid"`
	Username      string  `gorm:"column:username" json:"username"`
	PasswordHash  string  `gorm:"column:password_hash" json:"-"`
	Nickname      string  `gorm:"column:nickname" json:"nickname"`
	Email         string  `gorm:"column:email" json:"email"`
	Mobile        string  `gorm:"column:mobile" json:"mobile"`
	Avatar        string  `gorm:"column:avatar" json:"avatar"`
	Code          string  `gorm:"column:code" json:"code"`
	Status        int     `gorm:"column:status" json:"status"`
	StatusText    string  `gorm:"column:status_text" json:"status_text"`
	Money         float64 `gorm:"column:money" json:"money"`
	Score         float64 `gorm:"column:score" json:"score"`
	WechatUnionID string  `gorm:"column:wechat_unionid" json:"wechat_unionid"`
	WechatOpenID  string  `gorm:"column:wechat_openid" json:"wechat_openid"`
	Version       int     `gorm:"column:version" json:"version"`
}

func (AppUser) TableName() string { return "app_user" }

type AppUserSession struct {
	sharedmodel.BaseModel
	AppUserID  int64     `gorm:"column:app_user_id" json:"app_user_id"`
	ExpiresAt  time.Time `gorm:"column:expires_at" json:"expires_at"`
	LastSeenAt time.Time `gorm:"column:last_seen_at" json:"last_seen_at"`
	UserAgent  string    `gorm:"column:user_agent" json:"user_agent"`
	IP         string    `gorm:"column:ip" json:"ip"`
}

func (AppUserSession) TableName() string { return "app_user_session" }

type UserMoneyLog struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	UserID      int64     `gorm:"column:user_id" json:"user_id"`
	Type        int       `gorm:"column:type" json:"type"`
	Money       float64   `gorm:"column:money" json:"money"`
	BeforeMoney float64   `gorm:"column:before_money" json:"before_money"`
	AfterMoney  float64   `gorm:"column:after_money" json:"after_money"`
	Note        string    `gorm:"column:note" json:"note"`
	Source      string    `gorm:"column:source" json:"source"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
}

func (UserMoneyLog) TableName() string { return "user_money_log" }

type UserScoreLog struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	UserID      int64     `gorm:"column:user_id" json:"user_id"`
	Type        int       `gorm:"column:type" json:"type"`
	Score       float64   `gorm:"column:score" json:"score"`
	BeforeScore float64   `gorm:"column:before_score" json:"before_score"`
	AfterScore  float64   `gorm:"column:after_score" json:"after_score"`
	Note        string    `gorm:"column:note" json:"note"`
	Source      string    `gorm:"column:source" json:"source"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
}

func (UserScoreLog) TableName() string { return "user_score_log" }
