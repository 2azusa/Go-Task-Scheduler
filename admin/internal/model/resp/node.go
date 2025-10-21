package resp

import "pulse/common/models"

type (
	RspNodeSearch struct {
		models.Node
		JobCount int `json:"jobCount"`
	}
)
