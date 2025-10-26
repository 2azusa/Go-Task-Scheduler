package handler

import (
	"pulse/admin/internal/middlerware"
	"pulse/admin/internal/model/resp"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// RegisterRouters 是注册所有路有的入口点
func RegisterRouters(r *gin.Engine) {
	r.Use(middlerware.Cors())                                            // 使用Cors中间件来允许跨域请求
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler)) // 添加swagger路由
	apiGroup := r.Group("/api")
	configRoute(apiGroup) // 注册所有的API相关路由
	configNoRoute(r)      // 配置静态文件服务，用于前端页面
}

func configNoRoute(r *gin.Engine) {
	r.Static("/assets", "dist/assets")
	r.StaticFile("/favicon.ico", "dist/favicon.ico")
	r.NoRoute(func(c *gin.Context) {
		c.File("dist/index.html")
	})
}

type Hello struct {
	Name string `json:"name" form:"name"`
}

func configRoute(r *gin.RouterGroup) {
	// 创建一个 /ping 路由组，用于健康检查
	hello := r.Group("/ping")
	{
		// 心跳检测
		hello.GET("", func(c *gin.Context) {
			c.JSON(200, "pong")
		})
		// 带名称的心跳检测
		hello.POST("", func(c *gin.Context) {
			var h Hello
			err := c.ShouldBindJSON(&h)
			if err != nil {
				c.JSON(resp.ERROR, err.Error())
			}
			c.JSON(200, "hello,"+h.Name)
		})
	}

	base := r.Group("")
	{
		base.POST("register", defaultUserRouter.Register) // 注册用户接口
		base.POST("login", defaultUserRouter.Login)       // 登陆接口
	}

	stat := r.Group("/statis")
	stat.Use(middlerware.JwtToken())
	{
		stat.GET("today", defaultStatRouter.GetTodayStatistics) // 获取今日统计数据
		stat.GET("week", defaultStatRouter.GetWeekStatistics)   // 获取本周统计数据
		stat.POST("system", defaultStatRouter.GetSystemInfo)    // 获取系统信息
	}

	job := r.Group("/job")
	job.Use(middlerware.JwtToken())
	{
		job.POST("add", defaultJobRouter.CreateOrUpdate) // 添加或更新任务
		job.POST("del", defaultJobRouter.Delete)         // 删除任务
		job.POST("find", defaultJobRouter.FindById)      // 根据ID查找任务
		job.POST("search", defaultJobRouter.Search)      // 搜索任务列表
		job.POST("log", defaultJobRouter.SearchLog)      // 搜索任务日志
		job.POST("once", defaultJobRouter.Once)          // 立即执行一个任务
		job.POST("kill", defaultJobRouter.Kill)          // 强行终止一个任务
	}

	user := r.Group("/user")
	user.Use(middlerware.JwtToken())
	{
		user.POST("del", defaultUserRouter.Delete)               // 删除用户
		user.POST("update", defaultUserRouter.Update)            // 更新用户信息
		user.POST("change_pw", defaultUserRouter.ChangePassword) // 修改密码
		user.POST("find", defaultUserRouter.FindById)            // 根据ID查找用户
		user.POST("search", defaultUserRouter.Search)            // 搜索用户列表
	}

	node := r.Group("/node")
	node.Use(middlerware.JwtToken())
	{
		node.POST("search", defaultNodeRouter.Search) // 搜索节点列表
		node.POST("del", defaultNodeRouter.Delete)    // 删除节点
	}

	script := r.Group("/script")
	script.Use(middlerware.JwtToken())
	{
		script.POST("add", defaultScriptRouter.CreateOrUpdate) // 添加或更新脚本
		script.POST("del", defaultScriptRouter.Delete)         // 删除脚本
		script.POST("find", defaultScriptRouter.FindById)      // 根据ID查找脚本
		script.POST("search", defaultScriptRouter.Search)      // 搜索脚本列表
	}
}
