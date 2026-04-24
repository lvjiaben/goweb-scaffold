package system_home

type DashboardResponse struct {
	UserMoney        int64     `json:"user_money"`
	UserScore        int64     `json:"user_score"`
	UserCount        int64     `json:"user_count"`
	UserToday        int64     `json:"user_today"`
	UserStatus       int64     `json:"user_status"`
	AdminCount       int64     `json:"admin_count"`
	UploadCount      int64     `json:"upload_count"`
	UploadTodayCount int64     `json:"upload_today_count"`
	LineChart        LineChart `json:"line_chart"`
}

type LineChart struct {
	XAxis []string   `json:"xAxis"`
	YAxis [3][]int64 `json:"yAxis"`
}
