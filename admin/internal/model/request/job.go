package request

import (
	"encoding/json"
	"pulse/common/models"
)

type (
	ReqJobSearch struct {
		PageInfo
		ID     int            `json:"id" form:"id"`
		Name   string         `json:"name" form:"name"`
		RunOn  string         `json:"runOn" form:"runOn"`
		Type   models.JobType `json:"jobType" form:"type"`
		Status *int           `json:"status" form:"status"`
	}
	ReqJobLogSearch struct {
		PageInfo
		Name     string `json:"name" form:"name"`
		JobId    int    `json:"jobId" form:"jobId"`
		NodeUUID string `json:"nodeUuid" form:"nodeUuid"`
		Success  *bool  `json:"success" form:"success"`
	}
	ReqJobUpdate struct {
		*models.Job
		Allocation int `json:"allocation" form:"allocation" binding:"required"`
	}
	ReqJobOnce struct {
		JobId    int    `json:"jobId" form:"jobId"`
		NodeUUID string `json:"nodeUuid" form:"nodeUuid"`
	}
	ReqJobKill struct {
		JobId    int    `json:"jobId" form:"jobId"`
		NodeUUID string `json:"nodeUuid" form:"nodeUuid"`
	}
)

func (r *ReqJobUpdate) Valid() error {
	// default automatic assignment
	if r.Allocation == 0 {
		r.Allocation = models.AutoAllocation
	}
	notifyTo, _ := json.Marshal(r.NotifyToArray)
	r.NotifyTo = notifyTo
	scriptID, _ := json.Marshal(r.ScriptIDArray)
	r.ScriptID = scriptID
	return r.Check()
}
