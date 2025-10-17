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
	// 使用Cors中间件来允许跨域请求
	r.Use(middlerware.Cors())
	// 添加swagger路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// 注册所有的API相关路有
	configRoute(r)
	// 配置静态文件服务，用于前端页面
	configNoRoute(r)
}

// configRoute 用于配置所有的API接口路有
func configRoute(r *gin.Engine) {
	// 创建一个 /ping 路由组，用于健康检查
	hello := r.Group("/ping")
	{
		// GET /ping，返回 "pong" 字段
		hello.GET("", func(c *gin.Context) {
			c.JSON(200, "pong")
		})
		// POST /ping，用于测试 POST 请求和 JSON 绑定
		hello.POST("", func(c *gin.Context) {
			type Hello struct {
				Name string `json:"name" form:"name"`
			}
			var h Hello
			err := c.ShouldBindJSON(&h)
			if err != nil {
				c.JSON(resp.ERROR, err.Error())
			}
			c.JSON(200, "hello,"+h.Name)
		})
	}

	// 创建一个基础路有，路径为 / ，用于存放不需要认证的接口
	base := r.Group("")
	{
		base.POST("register", defaultUserRouter.Register) // 注册用户接口
		base.POST("login", defaultUserRouter.Login)       // 登陆接口ww
	}

	// 创建 /statis 路由组，用于统计信息，需要JWT认证
	stat := r.Group("/statis")
	// stat.Use(middlerware.JWTAuth())
	{
		stat.GET("today", defaultStatRouter.GetTodayStatistics) // 获取今日统计数据
		stat.GET("week", defaultStatRouter.GetWeekStatistics)   // 获取本周统计数据
		stat.GET("system", defaultStatRouter.GetSystemInfo)     // 获取系统信息
	}

	// 创建 /job 路由组，用于任务管理，需要JWT认证
	job := r.Group("/job")
	// job.Use(middlerware.JWTAuth())
	{
		job.POST("add", defaultJobRouter.CreateOrUpdate) // 添加或更新任务
		job.POST("del", defaultJobRouter.Delete)         // 删除任务
		job.GET("find", defaultJobRouter.FindById)       // 根据ID查找任务
		job.POST("search", defaultJobRouter.Search)      // 搜索任务列表
		job.POST("log", defaultJobRouter.SearchLog)      // 搜索任务日志
		job.POST("once", defaultJobRouter.Once)          // 立即执行一个任务
	}

	// 创建 /user 路由组，用于用户管理，需要JWT认证
	user := r.Group("/user")
	// user.Use(middlerware.JWTAuth())
	{
		user.POST("del", defaultUserRouter.Delete)               // 删除用户
		user.POST("update", defaultUserRouter.Update)            // 更新用户信息
		user.POST("change_pw", defaultUserRouter.ChangePassword) // 修改密码
		user.GET("find", defaultUserRouter.FindById)             // 根据ID查找用户
		user.POST("search", defaultUserRouter.Search)            // 搜索用户列表
	}
	node := r.Group("/node")
	// node.Use(middlerware.JWTAuth())
	{
		node.POST("search", defaultNodeRouter.Search) // 搜索节点列表
		node.POST("del", defaultNodeRouter.Delete)    // 删除节点
	}
	script := r.Group("/script")
	// script.Use(middlerware.JWTAuth())
	{
		script.POST("add", defaultScriptRouter.CreateOrUpdate) // 添加或更新脚本
		script.POST("del", defaultScriptRouter.Delete)         // 删除脚本
		script.GET("find", defaultScriptRouter.FindById)       // 根据ID查找脚本
		script.POST("search", defaultScriptRouter.Search)      // 搜索脚本列表
	}
}

// configNoRoute 配置静态文件服务
func configNoRoute(r *gin.Engine) {
	r.LoadHTMLGlob("./dist/*.html") // npm打包成dist的路径
	r.StaticFile("favicon.ico", "./dist/favicon.ico")
	r.Static("/css", "./dist/css")         // dist里面的静态资源
	r.Static("/fonts", "./dist/fonts")     // dist里面的静态资源
	r.Static("/js", "./dist/js")           // dist里面的静态资源
	r.Static("/img", "./dist/img")         // dist里面的静态资源
	r.StaticFile("/", "./dist/index.html") // 前端网页入口页面
}

// func configNoRoute(r *gin.Engine) {
// 	// 将所有 `./dist` 修改为 `admin/web/dist`
// 	r.LoadHTMLGlob("admin/web/dist/*.html")
// 	r.StaticFile("favicon.ico", "admin/web/dist/favicon.ico")
// 	r.Static("/css", "admin/web/dist/css")
// 	r.Static("/fonts", "admin/web/dist/fonts")
// 	r.Static("/js", "admin/web/dist/js")
// 	r.Static("/img", "admin/web/dist/img")
// 	r.StaticFile("/", "admin/web/dist/index.html")
// }
