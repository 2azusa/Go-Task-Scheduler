package middlerware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Cors 中间件用于处理跨域资源共享请求
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求的方法
		method := c.Request.Method
		// 获取请求头的“origin”字段
		origin := c.Request.Header.Get("origin")
		// 设置"Access-Control-Allow-Origin"响应头，允许来自指定源的跨域请求
		c.Header("Access-Control-Allow-Origin", origin)
		// 设置“Access-Control-Allow-Headers”响应头，指定在实际请求中可以使用的HTTP请求头
		c.Header("Access-Control-Allow-Headers", "Content-Type, AccessToken, X-CSRF-Token, Authorization, Token,Authorization, X-User-Id")
		// 设置“Access-Control-Allow-Headers"响应头，指定允许的跨域请求方法
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PUT")
		// 设置”Access-Control-Expose-Header“响应头，指定哪些响应头可以作为响应的一部分暴露给客户端
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		// 设置”Access-Control-Allow-Credentials“响应头，白哦是是否可以将响应暴露给页面
		c.Header("Access-Control-Allow-Credentials", "true")

		// 如果请求方法是预检请求，则中止请求
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}
