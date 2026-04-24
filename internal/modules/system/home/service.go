package system_home

import (
	"time"

	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
)

type Service struct {
	repo *Repo
}

func NewService(runtime *bootstrap.Runtime) *Service {
	return &Service{repo: NewRepo(runtime)}
}

func (s *Service) Dashboard(days int) (DashboardResponse, error) {
	if days != 30 {
		days = 7
	}
	today := truncateDay(time.Now())
	userCount, err := s.repo.CountAppUsers()
	if err != nil {
		return DashboardResponse{}, err
	}
	userToday, err := s.repo.CountAppUsersCreatedSince(today)
	if err != nil {
		return DashboardResponse{}, err
	}
	userStatus, err := s.repo.CountDisabledAppUsers()
	if err != nil {
		return DashboardResponse{}, err
	}
	adminCount, err := s.repo.CountAdmins()
	if err != nil {
		return DashboardResponse{}, err
	}
	uploadCount, err := s.repo.CountUploads()
	if err != nil {
		return DashboardResponse{}, err
	}
	uploadToday, err := s.repo.CountUploadsCreatedSince(today)
	if err != nil {
		return DashboardResponse{}, err
	}
	return DashboardResponse{
		UserMoney:        0,
		UserScore:        0,
		UserCount:        userCount,
		UserToday:        userToday,
		UserStatus:       userStatus,
		AdminCount:       adminCount,
		UploadCount:      uploadCount,
		UploadTodayCount: uploadToday,
		LineChart:        emptyLineChart(days),
	}, nil
}

func emptyLineChart(days int) LineChart {
	xAxis := make([]string, 0, days)
	for i := days - 1; i >= 0; i-- {
		xAxis = append(xAxis, truncateDay(time.Now()).AddDate(0, 0, -i).Format("01-02"))
	}
	return LineChart{
		XAxis: xAxis,
		YAxis: [3][]int64{
			make([]int64, days),
			make([]int64, days),
			make([]int64, days),
		},
	}
}

func truncateDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}
