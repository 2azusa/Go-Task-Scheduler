package resp

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type (
	PageResult struct {
		List     any   `json:"list"`
		Total    int64 `json:"total"`
		Page     int   `json:"page"`
		PageSize int   `json:"pageSize"`
	}
	Response struct {
		Code int    `json:"code"`
		Data any    `json:"data"`
		Msg  string `json:"msg"`
	}
)

const (
	SUCCESS = 200
	ERROR   = 1000

	ErrorRequestParameter = 1001
	ErrorJobFormat        = 1002
	ErrorTokenGenerate    = 1003
	ErrorUserNameExist    = 1004
)

func Result(code int, data any, msg string, c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Data: data,
		Msg:  msg,
	})
}

func Ok(c *gin.Context) {
	Result(SUCCESS, map[string]any{}, "operation success", c)
}

func OkWithMessage(message string, c *gin.Context) {
	Result(SUCCESS, map[string]any{}, message, c)
}

func OkWithData(data any, c *gin.Context) {
	Result(SUCCESS, data, "opertaion success", c)
}

func OkWithDetailed(data any, message string, c *gin.Context) {
	Result(SUCCESS, data, message, c)
}

func FailWithMessage(code int, message string, c *gin.Context) {
	Result(code, map[string]any{}, message, c)
}

func FailWithCode(code int, c *gin.Context) {
	Result(code, map[string]any{}, "operation failed", c)
}

func FailWithDetailed(code int, data any, message string, c *gin.Context) {
	Result(code, data, message, c)
}
