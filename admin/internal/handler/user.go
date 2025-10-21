package handler

import (
	"errors"
	"fmt"
	"net/http"
	"pulse/admin/internal/middlerware"
	"pulse/admin/internal/model/request"
	"pulse/admin/internal/model/resp"
	"pulse/admin/internal/service"
	"pulse/common/models"
	"pulse/common/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserRouter 用户路由结构体
type UserRouter struct{}

var defaultUserRouter = new(UserRouter)

// @Summary 用户登录
// @Description 验证用户并返回token
// @Tags users
// @Accept  json
// @Produce  json
// @Param   body  body   request.ReqUserLogin  true  "Login Credentials"
// @Success 200 {object} resp.RspLogin "Successful login"
// @Failure 400 {object} resp.Response "Bad request"
// @Failure 500 {object} resp.Response "Internal server error"
// @Router /login [post]
func (u *UserRouter) Login(c *gin.Context) {
	var req request.ReqUserLogin
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[user_login] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[user_login] request parameter error", c)
		return
	}
	user, err := service.DefaultUserService.Login(req.UserName, req.Password)
	if err != nil || user.ID == 0 {
		resp.FailWithMessage(resp.ERROR, "[user_login] username or password is incorrect", c)
		return
	}

	token, err := middlerware.SetToken(user.ID, user.UserName)
	if err != nil {
		resp.FailWithMessage(resp.ErrorTokenGenerate, "[user_login] failed to get token", c)
		return
	}

	resp.OkWithDetailed(resp.RspLogin{
		User:  user,
		Token: token,
	}, "login success", c)
}

// @Summary 注册新用户
// @Description 创建新的用户帐户
// @Tags users
// @Accept  json
// @Produce  json
// @Param   body  body   request.ReqUserRegister  true  "User registration details"
// @Success 200 {object} models.User "Successfully registered user"
// @Failure 400 {object} resp.Response "Bad request"
// @Failure 500 {object} resp.Response "Internal server error"
// @Router /register [post]
func (u *UserRouter) Register(c *gin.Context) {
	var req request.ReqUserRegister
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[user_register] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[user_register] request parameter error", c)
		return
	}
	user, err := service.DefaultUserService.FindByUserName(req.UserName)
	if user != nil || err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[user_register] db find by username: %s", req.UserName))
		resp.FailWithMessage(resp.ErrorUserNameExist, "[user_register] the user name has already been used", c)
		return
	}
	if req.Role == 0 {
		req.Role = models.RoleNormal
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[user_register] failed to hash password: %v", err))
		resp.FailWithMessage(resp.ERROR, "[user_register] failed to hash password", c)
		return
	}

	userModel := &models.User{
		UserName: req.UserName,
		Password: string(hashedPassword),
		Role:     req.Role,
		Email:    req.Email,
		Created:  time.Now().Unix(),
	}
	insertId, err := userModel.Insert()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[user_register] db insert error: %v", err))
		resp.FailWithMessage(resp.ERROR, "[user_register] db insert error", c)
		return
	}
	userModel.ID = insertId
	resp.OkWithDetailed(userModel, "register success", c)
}

// @Summary 更新用户信息
// @Description 更新用户的个人信息
// @Tags users
// @Accept  json
// @Produce  json
// @Param   body  body   models.User  true  "User information to update"
// @Success 200 {object} resp.Response "Update success"
// @Failure 400 {object} resp.Response "Bad request"
// @Failure 500 {object} resp.Response "Internal server error"
// @Router /user/update [post]
func (u *UserRouter) Update(c *gin.Context) {
	var req request.ReqUserUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[update_user] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[update_user] request parameter error", c)
		return
	}

	userId, _ := c.Get("userID")
	// todo if !exists return 身份验证失败

	err := service.DefaultUserService.UpdateUser(userId.(int), &req)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[update_user] db update error: %v", err))
		resp.FailWithMessage(resp.ERROR, "[update_user] db update error", c)
		return
	}
	resp.OkWithMessage("update success", c)
}

// @Summary 删除用户
// @Description 根据ID删除一个或多个用户
// @Tags users
// @Accept  json
// @Produce  json
// @Param   body  body   request.ByIDS  true  "List of user IDs"
// @Success 200 {object} resp.Response "Delete success"
// @Failure 400 {object} resp.Response "Bad request"
// @Router /user/del [post]
func (u *UserRouter) Delete(c *gin.Context) {
	var req request.ByIDS
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[delete_user] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[delete_user] request parmaeter error", c)
		return
	}
	for _, id := range req.IDS {
		userModel := models.User{ID: id}
		err := userModel.Delete()
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("[delete_user] db error: %v", err))
		}
	}
	resp.OkWithMessage("delete success", c)
}

// @Summary 更改用户密码
// @Description 更改当前登录用户的密码
// @Tags users
// @Accept  json
// @Produce  json
// @Param   body  body   request.ReqChangePassword  true  "Password change request"
// @Success 200 {object} resp.Response "Update success"
// @Failure 400 {object} resp.Response "Bad request"
// @Failure 500 {object} resp.Response "Internal server error"
// @Router /user/change_pw [post]

// func (u *UserRouter) ChangePassword(c *gin.Context) {
// 	var req request.ReqChangePassword
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		logger.GetLogger().Error(fmt.Sprintf("[change_password] request parameter error: %s", err.Error()))
// 		resp.FailWithMessage(resp.ErrorRequestParameter, "[change_password] request parameter error", c) // 1001
// 		return
// 	}

//		userId, _ := c.Get("userID")
//		err := service.DefaultUserService.ChangePassword(userId.(int), req.Password, req.NewPassword)
//		if err != nil {
//			logger.GetLogger().Error(fmt.Sprintf("[change_password] db error: %v", err))
//			resp.FailWithMessage(resp.ERROR, "[change_password] db error", c) // 1000
//			return
//		}
//		resp.OkWithMessage("update success", c)
//	}
func (u *UserRouter) ChangePassword(c *gin.Context) {
	var req request.ReqChangePassword
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[change_password] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[change_password] request parameter error", c)
		return
	}

	// --- 关键的安全检查 ---
	// 1. 检查 context 中是否存在 "userID"
	val, exists := c.Get("userID")
	if !exists {
		logger.GetLogger().Error("[change_password] failed to get userID from context. Token may be missing or invalid.")
		// 明确告诉前端需要认证
		resp.FailWithMessage(resp.ERROR, "user not authenticated", c)
		return
	}

	// 2. 检查 "userID" 的类型是否正确
	userId, ok := val.(int)
	if !ok {
		logger.GetLogger().Error("[change_password] userID in context is not of expected type (int)")
		resp.FailWithMessage(resp.ERROR, "internal server error: invalid user id type", c)
		return
	}
	// --- 安全检查结束 ---

	// 调用我们已经确认是安全的 Service 层代码
	err := service.DefaultUserService.ChangePassword(userId, req.Password, req.NewPassword)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[change_password] service error: %v", err))
		// 将 Service 层返回的具体错误信息返回给前端
		resp.FailWithMessage(resp.ERROR, err.Error(), c)
		return
	}
	resp.OkWithMessage("update success", c)
}

// @Summary 根据ID查找用户
// @Description 根据用户ID检索单个用户的详细信息。如果未提供ID，则检索当前用户的详细信息
// @Tags users
// @Produce  json
// @Param   id query int false "User ID"
// @Success 200 {object} models.User "Find success"
// @Failure 400 {object} resp.Response "Bad request"
// @Failure 500 {object} resp.Response "Internal server error"
// @Router /user/find [get]
func (u *UserRouter) FindById(c *gin.Context) {

	// // --- 开始调试 ---
	// // 1. 读取请求体
	// bodyBytes, err := io.ReadAll(c.Request.Body)
	// if err != nil {
	// 	logger.GetLogger().Error(fmt.Sprintf("[find_job] failed to read request body: %s", err.Error()))
	// 	resp.FailWithMessage(resp.ERROR, "[find_job] failed to read request body", c)
	// 	return
	// }
	// // 2. 将读取的 body 写回，以便 ShouldBindJSON 还能使用
	// c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// // 3. 打印所有诊断信息到你的日志
	// logger.GetLogger().Info("================ DEBUGGING REQUEST ================")
	// logger.GetLogger().Info(fmt.Sprintf("Request Method: %s", c.Request.Method))
	// logger.GetLogger().Info(fmt.Sprintf("Content-Type Header: %s", c.GetHeader("Content-Type")))
	// logger.GetLogger().Info(fmt.Sprintf("Raw Request Body: %s", string(bodyBytes)))
	// logger.GetLogger().Info("===================================================")
	// // --- 结束调试 ---

	var req request.ByID
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[find_user] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[find_user] request parameter error", c)
		return
	}

	userId, _ := c.Get("userID")
	if req.ID == 0 {
		req.ID = userId.(int)
	}

	userModel := models.User{ID: req.ID}
	err := userModel.FindById()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.GetLogger().Info(fmt.Sprintf("[find_user] 未找到 ID 为 %d 的用户", req.ID))
			// 以 404 Not Found 状态码响应
			c.JSON(http.StatusNotFound, gin.H{
				"code": 404,
				"msg":  fmt.Sprintf("未找到 ID 为 %d 的用户", req.ID),
			})
			return
		}
		logger.GetLogger().Error(fmt.Sprintf("[find_user] db error: %v", err))
		resp.FailWithMessage(resp.ERROR, "[find_user] db error", c)
		return
	}
	resp.OkWithDetailed(userModel, "find success", c)
}

// @Summary 搜索用户
// @Description 根据指定条件搜索用户
// @Tags users
// @Accept  json
// @Produce  json
// @Param   body  body   request.ReqUserSearch  true  "Search criteria"
// @Success 200 {object} resp.PageResult "Search results"
// @Failure 400 {object} resp.Response "Bad request"
// @Failure 500 {object} resp.Response "Internal server error"
// @Router /user/search [post]
func (u *UserRouter) Search(c *gin.Context) {
	var req request.ReqUserSearch
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[serach_user] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[serach_user] request parameter error", c)
		return
	}

	req.Check()

	users, total, err := service.DefaultUserService.Search(&req)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[search_user] db error: %v", err))
		resp.FailWithMessage(resp.ERROR, "[search_user] db error", c)
		return
	}
	resp.OkWithDetailed(resp.PageResult{
		List:     users,
		Total:    total,
		PageSize: req.PageSize,
		Page:     req.Page,
	}, "search success", c)
}
