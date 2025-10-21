package handler

import (
	"fmt"
	"pulse/admin/internal/model/request"
	"pulse/admin/internal/model/resp"
	"pulse/admin/internal/service"
	"pulse/common/models"
	"pulse/common/pkg/etcdclient"
	"pulse/common/pkg/logger"
	"pulse/common/pkg/utils"
	"time"

	"github.com/gin-gonic/gin"
)

// StatRouter 定义了统计相关的路由
type StatRouter struct{}

var defaultStatRouter = new(StatRouter)

// @Summary 获取当日的统计数据
// @Description 获取当天的统计数据，包括节点数和任务执行数
// @Tags statis
// @Produce  json
// @Success 200 {object} resp.RspSystemStatistics "Today's statistics"
// @Failure 500 {object} resp.Response "Internal server error"
// @Router /statis/today [get]
func (s *StatRouter) GetTodayStatistics(c *gin.Context) {
	jobExcSuccess, err := service.DefaultJobService.GetTodayJobExcCount(models.JobExcSuccess)
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("[get_statisitcs] GetTodayJObExcCount(success) error: %s", err.Error()))
	}
	jobExcFail, err := service.DefaultJobService.GetTodayJobExcCount(models.JobExcFail)
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("[get_statisitcs] GetTodayJobExcCount(fail) error:%s", err.Error()))
	}
	jobRunningCount, err := service.DefaultJobService.GetRunningJobCount()
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("[get_statisitcs] GetNodeCount(success) error: %s", err.Error()))
	}
	normalNodeCount, err := service.DefaultNodeWatcher.GetNodeCount(models.NodeConnSuccess)
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("[get_statisitcs] GetNodeCount(fail) error: %s", err.Error()))
	}
	failNodeCount, err := service.DefaultNodeWatcher.GetNodeCount(models.NodeConnFail)
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("[get_statisitcs] GetNodeCount(fail) error: %s", err.Error()))
	}
	resp.OkWithDetailed(resp.RspSystemStatistics{
		NormalNodeCount:    normalNodeCount,
		FailNodeCount:      failNodeCount,
		JobExcSuccessCount: jobExcSuccess,
		JobRunningCount:    jobRunningCount,
		JobExcFailCount:    jobExcFail,
	}, "ok", c)
}

// @Summary 获取每周统计数据
// @Description 检索过去7天的任务执行统计数据
// @Tags statis
// @Produce  json
// @Success 200 {object} resp.RspDateCount "Weekly statistics"
// @Failure 500 {object} resp.Response "Internal server error"
// @Router /statis/week [get]
func (s *StatRouter) GetWeekStatistics(c *gin.Context) {
	t := time.Now()
	jobExcSuccess, err := service.DefaultJobService.GetJobExcCount(t.Unix()-60*60*24*7, t.Unix(), models.JobExcSuccess)
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("[get_week_statisitcs] GetTodayJobExcCount(success) error: %s", err.Error()))
	}
	jobExcFail, err := service.DefaultJobService.GetJobExcCount(t.Unix()-60*60*24*7, t.Unix(), models.JobExcFail)
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("[get_week_statisitcs] GetTodayJobExcCount(fail) error: %s", err.Error()))
	}
	resp.OkWithDetailed(resp.RspDateCount{
		SuccessDateCount: jobExcSuccess,
		FailDateCount:    jobExcFail,
	}, "ok", c)
}

// @Summary 获取系统信息
// @Description 检索管理服务器或指定工作节点的系统信息
// @Tags statis
// @Produce  json
// @Param   uuid query string false "Node UUID"
// @Success 200 {object} utils.Server "System information"
// @Failure 400 {object} resp.Response "Bad request"
// @Failure 500 {object} resp.Response "Internal server error"
// @Router /statis/system [get]
func (s *StatRouter) GetSystemInfo(c *gin.Context) {
	var req request.ByUUID
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[get_system_info] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[get_system_info] request parameter error", c)
		return
	}
	var server *utils.Server
	var err error
	if req.UUID == "" {
		server, err = utils.GetServerInfo()
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("[get_system_info] error: %s", err.Error()))
			resp.FailWithMessage(resp.ERROR, "[get_system_info] error: %s", c)
			return
		}
	} else {
		_, err := etcdclient.PutWithTtl(fmt.Sprintf(etcdclient.KeyEtcdSystemSwitch, req.UUID), models.NodeSystemInfoSwitch, 30)
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("get system info from node[%s] etcd put error: %s", req.UUID, err.Error()))
			resp.FailWithMessage(resp.ERROR, "[get_system_info] error", c)
			return
		}
		// worker节点上报信息会有延迟，默认等待2秒
		time.Sleep(2 * time.Second)
		server, err = service.GetNodeSystemInfo(req.UUID)
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("get system info from node[%s] watch key error: %s", req.UUID, err.Error()))
			resp.FailWithMessage(resp.ERROR, "[get_system_info] error: %s", c)
		}
	}
	resp.OkWithDetailed(server, "ok", c)
}
