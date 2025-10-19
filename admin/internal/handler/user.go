package handler

import (
	"fmt"
	"pulse/admin/internal/middlerware"
	"pulse/admin/internal/model/request"
	"pulse/admin/internal/model/resp"
	"pulse/admin/internal/service"
	"pulse/common/models"
	"pulse/common/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// UserRouter 用户路由结构体
type UserRouter struct{}

// defaultUserRouter 用于创建一个默认路由实例
var defaultUserRouter = new(UserRouter)

// @Summary User Login
// @Description Authenticates a user and returns a token.
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

// @Summary Register a new user
// @Description Creates a new user account.
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

// Update 处理更新用户信息的请求
// @Summary Update user information
// @Description Updates a user's information.
// @Tags users
// @Accept  json
// @Produce  json
// @Param   body  body   models.User  true  "User information to update"
// @Success 200 {object} resp.Response "Update success"
// @Failure 400 {object} resp.Response "Bad request"
// @Failure 500 {object} resp.Response "Internal server error"
// @Router /user/update [post]
func (u *UserRouter) Update(c *gin.Context) {
	var req models.User
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[update_user] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[update_user] request parameter error", c)
		return
	}
	req.Updated = time.Now().Unix()
	err := req.Update()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[update_user] db update error: %v", err))
		resp.FailWithMessage(resp.ERROR, "[update_user] db update error", c)
		return
	}
	resp.OkWithMessage("update success", c)
}

// @Summary Delete users
// @Description Deletes one or more users by their IDs.
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

// @Summary Change user password
// @Description Changes the password for the currently logged-in user.
// @Tags users
// @Accept  json
// @Produce  json
// @Param   body  body   request.ReqChangePassword  true  "Password change request"
// @Success 200 {object} resp.Response "Update success"
// @Failure 400 {object} resp.Response "Bad request"
// @Failure 500 {object} resp.Response "Internal server error"
// @Router /user/change_pw [post]
func (u *UserRouter) ChangePassword(c *gin.Context) {
	var req request.ReqChangePassword
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[change_password] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[change_password] request parameter error", c)
		return
	}

	userId, _ := c.Get("userID")
	err := service.DefaultUserService.ChangePassword(userId.(int), req.Password, req.NewPassword)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[change_password] db error: %v", err))
		resp.FailWithMessage(resp.ERROR, "[change_password] db error", c)
		return
	}
	resp.OkWithMessage("update success", c)
}

// @Summary Find a user by ID
// @Description Retrieves the details of a single user by their ID. If no ID is provided, it retrieves the current user's details.
// @Tags users
// @Produce  json
// @Param   id query int false "User ID"
// @Success 200 {object} models.User "Find success"
// @Failure 400 {object} resp.Response "Bad request"
// @Failure 500 {object} resp.Response "Internal server error"
// @Router /user/find [get]
func (u *UserRouter) FindById(c *gin.Context) {
	var req request.ByID
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[find_user] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[find_user] request parameter error", c)
		return
	}

	userId, _ := c.Get("userId")
	if req.ID == 0 {
		req.ID = userId.(int)
	}

	userModel := models.User{ID: req.ID}
	err := userModel.FindById()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[find_user] db error: %v", err))
		resp.FailWithMessage(resp.ERROR, "[find_user] db error", c)
		return
	}
	resp.OkWithDetailed(userModel, "find success", c)
}

// @Summary Search for users
// @Description Searches for users based on specified criteria.
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
