package handler

import (
	"crony/admin/internal/model/request"
	"crony/admin/internal/model/resp"
	"crony/admin/internal/service"
	"crony/common/models"
	"crony/common/pkg/logger"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

type ScriptRouter struct{}

var defaultScriptRouter = new(ScriptRouter)

// CreateOrUpdate 用于创建或更新脚本
func (s *ScriptRouter) CreateOrUpdate(c *gin.Context) {
	var req models.Script
	// 将请求的JSON绑定到req
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[create_script] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[creata_script] requestparameter error", c)
		return
	}
	var err error
	t := time.Now()
	if req.ID > 0 {
		// update
		req.Updated = t.Unix() // 设置更新时间为当前时间的Unix时间戳
		err = req.Update()     // 调用Update方法写入更新写入数据库
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("[update_script] into db error: %s", err.Error()))
			resp.FailWithMessage(resp.ERROR, "[update_script] into db id error", c)
			return
		}
	} else {
		// create
		req.Created = t.Unix() // 设置创建时间为当前时间的Unix时间戳
		_, err = req.Insert()  // 调用Insert方法将新脚本插入数据库
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("[create_script] insert script into db error: %s", err.Error()))
			resp.FailWithMessage(resp.ERROR, "[create_script] insert script into db error", c)
			return
		}
	}
	// 返回成功响应
	resp.OkWithDetailed(req, "operate success", c)
}

// Delete 用于删除脚本
func (s *ScriptRouter) Delete(c *gin.Context) {
	var req request.ByIDS
	// 将请求的JSON绑定到req
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[delete_script] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[delete_script] request parameter error", c)
		return
	}
	// 遍历请求中提供的所有ID
	for _, id := range req.IDS {
		script := models.Script{ID: id}
		// 根据ID查找脚本是否存在
		err := script.FindById()
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("[delete_script] find script by id: %d error: %s", id, err.Error()))
			continue
		}
		// 从数据库中删除脚本
		err = script.Delete()
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("[delete_script] into db error: %s", err.Error()))
			continue
		}
	}
	// 返回成功响应
	resp.OkWithMessage("delete success", c)
}

// 用于根据ID查找单个脚本
func (s *ScriptRouter) FindById(c *gin.Context) {
	var req request.ByID
	// 将URL查询参数中的ID绑定到req
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[find_script] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[find_script] request parameter error", c)
		return
	}
	script := models.Script{ID: req.ID}
	// 根据ID从数据库中查找脚本
	err := script.FindById()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[find_script] find script by id: %d error: %s", req.ID, err.Error()))
		resp.FailWithMessage(resp.ERROR, "[find_script] find script by id error", c)
		return
	}
	// 返回包含脚本的成功响应
	resp.OkWithDetailed(script, "find success", c)
}

// Search 用于搜索脚本列表
func (s *ScriptRouter) Search(c *gin.Context) {
	var req request.ReqScriptSearch
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[search_script] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[search_script] request parameter error", c)
		return
	}
	req.Check()
	scripts, total, err := service.DefaultScriptService.Search(&req)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[search_script] search script error: %s", err.Error()))
		resp.FailWithMessage(resp.ERROR, "[search_script] search script error", c)
		return
	}
	resp.OkWithDetailed(resp.PageResult{
		List:     scripts,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, "search success", c)
}
