package handler

import (
	"fmt"
	"pulse/admin/internal/model/request"
	"pulse/admin/internal/model/resp"
	"pulse/admin/internal/service"
	"pulse/common/models"
	"pulse/common/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
)

// ScriptRouter 处理脚本相关的路由
type ScriptRouter struct{}

var defaultScriptRouter = new(ScriptRouter)

// CreateOrUpdate 用于创建或更新脚本
// @Summary Create or update a script
// @Description Creates a new script or updates an existing one.
// @Tags script
// @Accept  json
// @Produce  json
// @Param   body  body   models.Script  true  "Script details"
// @Success 200 {object} models.Script "Operation success"
// @Failure 400 {object} resp.Response "Bad request"
// @Failure 500 {object} resp.Response "Internal server error"
// @Router /script/add [post]
func (s *ScriptRouter) CreateOrUpdate(c *gin.Context) {
	var req models.Script
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[create_script] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[creata_script] requestparameter error", c)
		return
	}
	var err error
	t := time.Now()
	if req.ID > 0 {
		req.Updated = t.Unix()
		err = req.Update()
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("[update_script] into db error: %s", err.Error()))
			resp.FailWithMessage(resp.ERROR, "[update_script] into db id error", c)
			return
		}
	} else {
		req.Created = t.Unix()
		_, err = req.Insert()
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("[create_script] insert script into db error: %s", err.Error()))
			resp.FailWithMessage(resp.ERROR, "[create_script] insert script into db error", c)
			return
		}
	}
	resp.OkWithDetailed(req, "operate success", c)
}

// @Summary Delete scripts
// @Description Deletes one or more scripts by their IDs.
// @Tags script
// @Accept  json
// @Produce  json
// @Param   body  body   request.ByIDS  true  "List of script IDs"
// @Success 200 {object} resp.Response "Delete success"
// @Failure 400 {object} resp.Response "Bad request"
// @Failure 500 {object} resp.Response "Internal server error"
// @Router /script/del [post]
func (s *ScriptRouter) Delete(c *gin.Context) {
	var req request.ByIDS
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[delete_script] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[delete_script] request parameter error", c)
		return
	}
	for _, id := range req.IDS {
		script := models.Script{ID: id}
		err := script.FindById()
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("[delete_script] find script by id: %d error: %s", id, err.Error()))
			continue
		}
		err = script.Delete()
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("[delete_script] into db error: %s", err.Error()))
			continue
		}
	}
	resp.OkWithMessage("delete success", c)
}

// @Summary Find a script by ID
// @Description Retrieves the details of a single script by its ID.
// @Tags script
// @Produce  json
// @Param   id query int true "Script ID"
// @Success 200 {object} models.Script "Find success"
// @Failure 400 {object} resp.Response "Bad request"
// @Failure 500 {object} resp.Response "Internal server error"
// @Router /script/find [get]
func (s *ScriptRouter) FindById(c *gin.Context) {
	var req request.ByID
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[find_script] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[find_script] request parameter error", c)
		return
	}
	script := models.Script{ID: req.ID}
	err := script.FindById()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[find_script] find script by id: %d error: %s", req.ID, err.Error()))
		resp.FailWithMessage(resp.ERROR, "[find_script] find script by id error", c)
		return
	}
	resp.OkWithDetailed(script, "find success", c)
}

// @Summary Search for scripts
// @Description Searches for scripts based on specified criteria.
// @Tags script
// @Accept  json
// @Produce  json
// @Param   body  body   request.ReqScriptSearch  true  "Search criteria"
// @Success 200 {object} resp.PageResult "Search results"
// @Failure 400 {object} resp.Response "Bad request"
// @Failure 500 {object} resp.Response "Internal server error"
// @Router /script/search [post]
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
