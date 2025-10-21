package request

type ReqScriptSearch struct {
	PageInfo
	ID   int    `json:"id" form:"id"`
	Name string `json:"name" form:"name"`
}

type ReqScriptUpdate struct {
	ID      *int    `json:"id" form:"id"`
	Name    *string `json:"name" form:"name"`
	Command *string `json:"command" form:"command"`
}
