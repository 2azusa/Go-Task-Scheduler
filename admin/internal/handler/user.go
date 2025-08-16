package handler

import (
	"fmt"
	"pulse/admin/internal/middlerware"
	"pulse/admin/internal/model/request"
	"pulse/admin/internal/model/resp"
	"pulse/admin/internal/service"
	"pulse/common/models"
	"pulse/common/pkg/logger"
	"pulse/common/pkg/utils"
	"time"

	"github.com/gin-gonic/gin"
)

// UserRouter 用户路由结构体
type UserRouter struct{}

// defaultUserRouter 用于创建一个默认路由实例
var defaultUserRouter = new(UserRouter)

// Login 用于处理用户登陆请求
func (u *UserRouter) Login(c *gin.Context) {
	var req request.ReqUserLogin
	// 绑定并验证请求的JSON数据u
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[user_login] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[user_login] request parameter error", c)
		return
	}
	// 调用服务层的登录方法验证用户名和密码
	user, err := service.DefaultUserService.Login(req.UserName, req.Password)
	if err != nil || user.ID == 0 {
		logger.GetLogger().Error(fmt.Sprintf("[user_login] db error: %v", err))
		resp.FailWithMessage(resp.ERROR, "[user_login] username or password is incorrect", c)
		return
	}

	// 创建JWT实例
	j := middlerware.NewJWT()
	// 创建JWT的claims
	claims := j.CreateClaims(middlerware.BaseClaims{
		ID:       user.ID,
		UserName: user.UserName,
	})
	// 生成JWT token
	token, err := j.CreateToken(claims)
	if err != nil {
		resp.FailWithMessage(resp.ErrorTokenGenerate, "获取token失败", c)
		return
	}
	// 返回成功的响应，包含用户信息的token
	resp.OkWithDetailed(resp.RspLogin{
		User:  user,
		Token: token,
	}, "login success", c)
}

// Register 用于处理用户注册请求
func (u *UserRouter) Register(c *gin.Context) {
	var req request.ReqUserRegister
	// 绑定并验证请求的JSONshuju
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[user_register] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[user_register] request parameter error", c)
		return
	}
	// 检查用户名是否存在
	user, err := service.DefaultUserService.FindByUserName(req.UserName)
	if user != nil || err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[user_register] db find by username: %s", req.UserName))
		resp.FailWithMessage(resp.ErrorUserNameExist, "[user_register] the user name has already been used", c)
		return
	}
	// 如果请求中为指定角色，则设置为普通用户
	if req.Role == 0 {
		req.Role = models.RoleNormal
	}
	// 创建用户模型实例
	userModel := &models.User{
		UserName: req.UserName,
		Password: utils.MD5(req.Password),
		Role:     req.Role,
		Email:    req.Email,
		Created:  time.Now().Unix(),
	}
	// 将新用户插入数据库
	insertId, err := userModel.Insert()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[user_register] db insert error: %v", err))
		resp.FailWithMessage(resp.ERROR, "[user_register] db insert error", c)
		return
	}
	userModel.ID = insertId
	// 返回成功的响应，包含新创建的用户信息
	resp.OkWithDetailed(userModel, "register success", c)
}

// Update 处理更新用户信息的请求
func (u *UserRouter) Update(c *gin.Context) {
	var req models.User
	// 绑定并验证请求的JSON数据
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[update_user] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[update_user] request parameter error", c)
		return
	}
	// 设置更新时间
	req.Updated = time.Now().Unix()
	// 调用模型方法更新用户信息
	err := req.Update()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[update_user] db update error: %v", err))
		resp.FailWithMessage(resp.ERROR, "[update_user] db update error", c)
		return
	}
	// 返回成功的响应消息
	resp.OkWithMessage("update success", c)
}

// Delete 处理删除用户的请求
func (u *UserRouter) Delete(c *gin.Context) {
	var req request.ByIDS
	// 绑定并验证包含ID列表的请求JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[delete_user] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[delete_user] request parmaeter error", c)
		return
	}
	// 遍历ID列表，逐个删除用户
	for _, id := range req.IDS {
		userModel := models.User{ID: id}
		err := userModel.Delete()
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("[delete_user] db error: %v", err))
		}
	}
	// 返回成功的响应消息
	resp.OkWithMessage("delete success", c)
}

// ChangePassword 用于处理修改密码的请求
func (u *UserRouter) ChangePassword(c *gin.Context) {
	var req request.ReqChangePassword
	// 绑定并验证请求的JSON数据
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[change_password] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[change_password] request parameter error", c)
		return
	}
	// 调用服务层方法修改密码
	// 从gin的上下文中获取当前登录用户的ID
	err := service.DefaultUserService.ChangePassword(middlerware.GetUserInfo(c).ID, req.Password, req.NewPassword)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[change_password] db error: %v", err))
		resp.FailWithMessage(resp.ERROR, "[change_password] db error", c)
		return
	}
	// 返回成功的响应消息
	resp.OkWithMessage("update success", c)
}

// FindById 用于根据ID查找用户
func (u *UserRouter) FindById(c *gin.Context) {
	var req request.ByID
	// 绑定并验证URL查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[find_user] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[find_user] request parameter error", c)
		return
	}
	if req.ID == 0 {
		req.ID = middlerware.GetUserInfo(c).ID
	}
	userModel := models.User{ID: req.ID}
	// 调用模型方法根据ID查找用户
	err := userModel.FindById()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[find_user] db error: %v", err))
		resp.FailWithMessage(resp.ERROR, "[find_user] db error", c)
		return
	}
	// 返回成功的响应，包含查找到的用户信息
	resp.OkWithDetailed(userModel, "finc_success", c)
}

// Search 用于根据条件搜索用户
func (u *UserRouter) Search(c *gin.Context) {
	var req request.ReqUserSearch
	// 绑定并验证请求的JSON数据
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[serach_user] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[serach_user] request parameter error", c)
		return
	}
	// 检查并设置分页参数的默认值
	req.Check()

	// 调用服务层方法进行搜索
	users, total, err := service.DefaultUserService.Search(&req)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[search_user] db error: %v", err))
		resp.FailWithMessage(resp.ERROR, "[search_user] db error", c)
		return
	}
	// 返回成功的响应
	resp.OkWithDetailed(resp.PageResult{
		List:     users,
		Total:    total,
		PageSize: req.PageSize,
		Page:     req.Page,
	}, "search success", c)
}
