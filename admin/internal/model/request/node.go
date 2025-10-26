package request

type (
	ReqNodeSearch struct {
		PageInfo
		IP     string `json:"ip" form:"ip"`         // ip
		UUID   string `json:"uuid" form:"uuid"`     // uuid
		UpTime int64  `json:"up" form:"up"`         // 启动时间
		Status int    `json:"status" form:"status"` // 状态 1 or 2
	}
	ByUUID struct {
		UUID string `json:"uuid" form:"uuid"`
	}
)
