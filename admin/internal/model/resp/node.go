package resp

import "crony/common/models"

type (
	RspNodeSearch struct {
		models.Node
		JobCount int `json:"job_count"`
	}
)
