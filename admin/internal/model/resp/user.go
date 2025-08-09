package resp

import "crony/common/models"

type (
	RspLogin struct {
		User  *models.User `json:"user"`
		Token string       `json:"token"`
	}
	RspUser struct {
		ID       int    `json:"id"`
		UserName string `json:"username"`
		Email    string `json:"email"`
		Role     int    `json:"role"`
		Status   int    `json:"statues"`
		Created  int64  `json:"created"`
		Updated  int64  `json:"updated"`
	}
)
