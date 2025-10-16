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

// defaultStatRouter 是StatRouter的默认实例
var defaultStatRouter = new(StatRouter)

// GetTodayStatistics 获取当日的统计数据
// @Summary Get today's statistics
// @Description Retrieves statistics for the current day, including node counts and job execution counts.
// @Tags statis
// @Produce  json
// @Success 200 {object} resp.RspSystemStatistics "Today's statistics"
// @Failure 500 {object} resp.Response "Internal server error"
// @Router /statis/today [get]
func (s *StatRouter) GetTodayStatistics(c *gin.Context) {
	// 获取今日任务成功执行的次数
	jobExcSuccess, err := service.DefaultJobService.GetTodayJobExcCount(models.JobExcSuccess)
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("[get_statisitcs] GetTodayJObExcCount(success) error: %s", err.Error()))
	}
	// 获取今日任务失败执行次数
	jobExcFail, err := service.DefaultJobService.GetTodayJobExcCount(models.JobExcFail)
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("[get_statisitcs] GetTodayJobExcCount(fail) error:%s", err.Error()))
	}
	// 获取正在运行的任务数量
	jobRunningCount, err := service.DefaultJobService.GetRunningJobCount()
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("[get_statisitcs] GetNodeCount(success) error: %s", err.Error()))
	}
	// 获取正常节点的数量
	normalNodeCount, err := service.DefaultNodeWatcher.GetNodeCount(models.NodeConnSuccess)
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("[get_statisitcs] GetNodeCount(fail) error: %s", err.Error()))
	}
	// 获取异常节点的数量
	failNodeCount, err := service.DefaultNodeWatcher.GetNodeCount(models.NodeConnFail)
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("[get_statisitcs] GetNodeCount(fail) error: %s", err.Error()))
	}
	// 返回包含所有统计数据的成功响应
	resp.OkWithDetailed(resp.RspSystemStatistics{
		NormalNodeCount:    normalNodeCount,
		FailNodeCount:      failNodeCount,
		JobExcSuccessCount: jobExcSuccess,
		JobRunningCount:    jobRunningCount,
		JobExcFailCount:    jobExcFail,
	}, "ok", c)
}

// @Summary Get weekly statistics
// @Description Retrieves job execution statistics for the last 7 days.
// @Tags statis
// @Produce  json
// @Success 200 {object} resp.RspDateCount "Weekly statistics"
// @Failure 500 {object} resp.Response "Internal server error"
// @Router /statis/week [get]
func (s *StatRouter) GetWeekStatistics(c *gin.Context) {
	// 获取当前时间
	t := time.Now()
	// 获取过去7天内任务参观执行的次数
	jobExcSuccess, err := service.DefaultJobService.GetJobExcCount(t.Unix()-60*60*24*7, t.Unix(), models.JobExcSuccess)
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("[get_week_statisitcs] GetTodayJobExcCount(success) error: %s", err.Error()))
	}
	// 获取过去7天内任务失败执行的次数
	jobExcFail, err := service.DefaultJobService.GetJobExcCount(t.Unix()-60*60*24*7, t.Unix(), models.JobExcFail)
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("[get_week_statisitcs] GetTodayJobExcCount(fail) error: %s", err.Error()))
	}
	// 返回包含周统计数据的成功响应
	resp.OkWithDetailed(resp.RspDateCount{
		SuccessDateCount: jobExcSuccess,
		FailDateCount:    jobExcFail,
	}, "ok", c)
}

// @Summary Get system information
// @Description Retrieves system information for the admin server or a specific worker node.
// @Tags statis
// @Produce  json
// @Param   uuid query string false "Node UUID"
// @Success 200 {object} utils.Server "System information"
// @Failure 400 {object} resp.Response "Bad request"
// @Failure 500 {object} resp.Response "Internal server error"
// @Router /statis/system [get]
func (s *StatRouter) GetSystemInfo(c *gin.Context) {
	var req request.ByUUID
	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[get_system_info] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[get_system_info] request parameter error", c)
		return
	}
	var server *utils.Server
	var err error
	// 如果请求中没有UUID，则获取admin自身的系统信息
	if req.UUID == "" {
		// 获取admin的本地服务器信息
		server, err = utils.GetServerInfo()
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("[get_system_info] error: %s", err.Error()))
			resp.FailWithMessage(resp.ERROR, "[get_system_info] error: %s", c)
			return
		}
	} else {
		// 获取指定woeker节点的系统信息
		_, err := etcdclient.PutWithTtl(fmt.Sprintf(etcdclient.KeyEtcdSystemSwitch, req.UUID), models.NodeSystemInfoSwitch, 30)
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("get system info from node[%s] etcd put error: %s", req.UUID, err.Error()))
			resp.FailWithMessage(resp.ERROR, "[get_system_info] error", c)
			return
		}
		// worker节点上报信息会有延迟，默认等待2秒
		time.Sleep(2 * time.Second)
		// 从service中获取节点上报的系统信息
		server, err = service.GetNodeSystemInfo(req.UUID)
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("get system info from node[%s] watch key error: %s", req.UUID, err.Error()))
			resp.FailWithMessage(resp.ERROR, "[get_system_info] error: %s", c)
		}
	}
	// 返回获取到的系统信息
	resp.OkWithDetailed(server, "ok", c)
}
