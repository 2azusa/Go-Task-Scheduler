package resp

type (
	RspSystemStatistics struct {
		NormalNodeCount    int64 `json:"normalNodeCount"`
		FailNodeCount      int64 `json:"failNodeCount"`
		JobExcSuccessCount int64 `json:"jobExcSuccessCount"`
		JobRunningCount    int64 `json:"jobRunningCount"`
		JobExcFailCount    int64 `json:"jobExcFailCount"`
	}
	DateCount struct {
		Date  string `json:"date" gorm:"column:date"`
		Count string `json:"count" gorm:"column:count"`
	}
	RspDateCount struct {
		SuccessDateCount []DateCount `json:"successDateCount"`
		FailDateCount    []DateCount `json:"failDateCount"`
	}
)
