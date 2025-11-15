package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"pulse/admin/internal/model/request"
	"pulse/admin/internal/model/resp"
	"pulse/admin/internal/service"
	"pulse/common/models"
	"pulse/common/pkg/config"
	"pulse/common/pkg/etcdclient"
	"pulse/common/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gorm.io/gorm"
)

// JobRouter 结构体用于组织与任务相关的路由处理器
type JobRouter struct{}

var defaultJobRouter = new(JobRouter)

// @Summary 创建或更新任务
// @Description 创建新任务或更新现有任务
// @Tags job
// @Accept  json
// @Produce  json
// @Param   body  body   request.ReqJobUpdate  true  "Job details"
// @Success 200 {object} request.ReqJobUpdate "Operation success"
// @Failure 400 {object} resp.Response "Bad request"
// @Failure 500 {object} resp.Response "Internal server error"
// @Router /job/add [post]
func (j *JobRouter) CreateOrUpdate(c *gin.Context) {
	var req request.ReqJobUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[create_job] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[create_job] request parameter error", c)
		return
	}
	if err := req.Valid(); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("create_job check error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorJobFormat, "[create_job] check error", c)
		return
	}

	var err error
	var insertId int
	t := time.Now()

	switch req.Allocation {
	case models.AutoAllocation:
		if !config.GetConfigModels().System.CmdAutoAllocation && req.Type == models.JobTypeCmd {
			resp.FailWithMessage(resp.ERROR, "[create_job] The shell command is not supported to automtically assign nodes by default", c)
			return
		}
		nodeUUID := service.DefaultJobService.AutoAllocateNode()
		if nodeUUID == "" {
			logger.GetLogger().Error("[create_job] auto allocate node error")
			resp.FailWithMessage(resp.ERROR, "[create_job] auto allocate node error", c)
			return
		}
		req.RunOn = nodeUUID
	case models.ManualAllocation:
		if len(req.RunOn) == 0 {
			resp.FailWithMessage(resp.ERROR, "[create_job] manually assigned node can't be null", c)
			return
		}
		node := &models.Node{UUID: req.RunOn}
		_ = node.FindByUUID()
		if node.Status == models.NodeConnFail {
			resp.FailWithMessage(resp.ERROR, "[create_job] manually assigned node inactiveation", c)
			return
		}
	}

	if req.ID > 0 {
		job := &models.Job{ID: req.ID}
		_ = job.FindById()
		oldNodeUUID := job.RunOn
		if oldNodeUUID != "" {
			_, err = etcdclient.Delete(fmt.Sprintf(etcdclient.KeyEtcdJob, oldNodeUUID, req.ID))
			if err != nil {
				logger.GetLogger().Error(fmt.Sprintf("[update_job] delete etcd node[%s] error: %s", oldNodeUUID, err.Error()))
				resp.FailWithMessage(resp.ERROR, "[update_job] delete etcd node error", c)
				return
			}
		}
		req.Updated = t.Unix()
		err = req.Update()
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("[update_job] into db error: %s", err.Error()))
			resp.FailWithMessage(resp.ERROR, "[update_job] into db error", c)
			return
		}
	} else {
		req.Created = t.Unix()
		insertId, err = req.Insert()
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("[create_job] json marshal job error: %s", err.Error()))
			resp.FailWithMessage(resp.ERROR, "[create_job] instert job into db error", c)
			return
		}
		req.ID = insertId
	}

	b, err := json.Marshal(req)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[create_job] json marshal job error: %s", err.Error()))
		resp.FailWithMessage(resp.ERROR, "[create_job] json marshal job error", c)
		return
	}
	// 将任务数据Put到etcd， key格式为 /pulse/job/<node_uuid>/<job_id>
	_, err = etcdclient.Put(fmt.Sprintf(etcdclient.KeyEtcdJob, req.RunOn, req.ID), string(b))
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[create_job] etcd put job error:%s", err.Error()))
		resp.FailWithMessage(resp.ERROR, "[create_job] etcd put job error", c)
		return
	}

	resp.OkWithDetailed(req, "operate success", c)
}

// @Summary 删除任务
// @Description 根据ID删除一个或多个任务。
// @Tags job
// @Accept  json
// @Produce  json
// @Param   body  body   request.ByIDS  true  "List of job IDs"
// @Success 200 {object} resp.Response "Delete success"
// @Failure 400 {object} resp.Response "Bad request"
// @Failure 500 {object} resp.Response "Internal server error"
// @Router /job/del [post]
func (j *JobRouter) Delete(c *gin.Context) {
	var req request.ByIDS
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[delete_job] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[delete_job] request parameter error", c)
		return
	}
	for _, id := range req.IDS {
		job := models.Job{ID: id}
		err := job.FindById()
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("[delete_job] find job by id: %d error: %s", id, err.Error()))
			continue
		}
		_, err = etcdclient.Delete(fmt.Sprintf(etcdclient.KeyEtcdJob, job.RunOn, id))
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("[delete_job] etcd delete job error: %s", err.Error()))
			continue
		}
		err = job.Delete()
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("[delete_job] into db error: %s", err.Error()))
			continue
		}
	}
	resp.OkWithMessage("delete success", c)
}

// @Summary 根据ID查找任务
// @Description 根据任务ID检索单个任务的详细信息
// @Tags job
// @Produce  json
// @Param   id query int true "Job ID"
// @Success 200 {object} models.Job "Find success"
// @Failure 400 {object} resp.Response "Bad request"
// @Failure 500 {object} resp.Response "Internal server error"
// @Router /job/find [get]
func (j *JobRouter) FindById(c *gin.Context) {
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

	var req request.ByID
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[find_job] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[find_job] request parameter error", c)
		return
	}
	job := models.Job{ID: req.ID}
	err := job.FindById()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.GetLogger().Info(fmt.Sprintf("[find_user] 未找到 ID 为 %d 的任务", req.ID))
			// 以 404 Not Found 状态码响应
			c.JSON(http.StatusNotFound, gin.H{
				"code": 404,
				"msg":  fmt.Sprintf("未找到 ID 为 %d 的任务", req.ID),
			})
			return
		}
		logger.GetLogger().Error(fmt.Sprintf("[find_job] find job by id: %d error: %s", req.ID, err.Error()))
		resp.FailWithMessage(resp.ERROR, "[find_job] find job by id error", c)
		return
	}
	if len(job.NotifyTo) != 0 {
		_ = job.Unmarshal()
	}
	resp.OkWithDetailed(job, "find success", c)
}

// @Description 根据指定条件搜索任务。
// @Tags job
// @Accept  json
// @Produce  json
// @Param   body  body   request.ReqJobSearch  true  "Search criteria"
// @Success 200 {object} resp.PageResult "Search results"
// @Failure 400 {object} resp.Response "Bad request"
// @Failure 500 {object} resp.Response "Internal server error"
// @Router /job/search [post]
func (j *JobRouter) Search(c *gin.Context) {
	// bodyBytes, err := io.ReadAll(c.Request.Body)
	// if err != nil {
	// 	logger.GetLogger().Error(fmt.Sprintf("[search_job] failed to read request body: %s", err.Error()))
	// 	resp.FailWithMessage(resp.ERROR, "[search_job] failed to read request body", c)
	// 	return
	// }
	// c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// logger.GetLogger().Info("================ DEBUGGING SEARCH REQUEST ================")
	// logger.GetLogger().Info(fmt.Sprintf("Request Method: %s", c.Request.Method))
	// logger.GetLogger().Info(fmt.Sprintf("Content-Type Header: %s", c.GetHeader("Content-Type")))
	// logger.GetLogger().Info(fmt.Sprintf("Raw Request Body: %s", string(bodyBytes)))
	// logger.GetLogger().Info("==========================================================")

	var req request.ReqJobSearch
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[search_job] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[search_job] request parameter error", c)
		return
	}
	req.Check()
	jobs, total, err := service.DefaultJobService.Search(&req)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[serach_job] search job error: %s", err.Error()))
		resp.FailWithMessage(resp.ERROR, "[serach_job] search job error", c)
		return
	}
	var resultJobs []models.Job
	for _, job := range jobs {
		_ = job.Unmarshal()
		resultJobs = append(resultJobs, job)
	}
	resp.OkWithDetailed(resp.PageResult{
		List:     resultJobs,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, "search success", c)
}

// @Summary 搜索任务日志
// @Description 根据指定条件搜索任务执行日志。
// @Tags job
// @Accept  json
// @Produce  json
// @Param   body  body   request.ReqJobLogSearch  true  "Search criteria for logs"
// @Success 200 {object} resp.PageResult "Search results for logs"
// @Failure 400 {object} resp.Response "Bad request"
// @Failure 500 {object} resp.Response "Internal server error"
// @Router /job/log [post]
func (j *JobRouter) SearchLog(c *gin.Context) {
	var req request.ReqJobLogSearch
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[search_job_log] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[search_job_log] request parameter error", c)
		return
	}
	req.Check()
	jobs, total, err := service.DefaultJobService.SearchJobLog(&req)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[search_job_log] db error: %s", err.Error()))
		resp.FailWithMessage(resp.ERROR, "[serach_job_log] db error", c)
		return
	}
	resp.OkWithDetailed(resp.PageResult{
		List:     jobs,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, "search success", c)
}

// @Summary 单次执行任务
// @Description 在指定节点上触发指定任务的单次执行
// @Tags job
// @Accept  json
// @Produce  json
// @Param   body  body   request.ReqJobOnce  true  "Job execution details"
// @Success 200 {object} resp.Response "Job once success"
// @Failure 400 {object} resp.Response "Bad request"
// @Failure 500 {object} resp.Response "Internal server error"
// @Router /job/once [post]
func (j *JobRouter) Once(c *gin.Context) {
	var req request.ReqJobOnce
	var err error
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[job_once] request parameter error:%s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[job_once] request parameter error", c)
		return
	}
	node := &models.Node{UUID: req.NodeUUID}
	err = node.FindByUUID()
	if err != nil || node.Status == models.NodeConnFail {
		logger.GetLogger().Error(fmt.Sprintf("[job_once] node[%s] conn fail:%v", req.NodeUUID, err))
		resp.FailWithMessage(resp.ERROR, "[job_once] node conn fail ", c)
		return
	}
	job := &models.Job{ID: req.JobId}
	err = job.FindById()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[job_once] job_id[%d] not exist db:%s", req.JobId, err.Error()))
		resp.FailWithMessage(resp.ERROR, "[job_once] job not exist ", c)
		return
	}

	err = service.DefaultJobService.Once(&req)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[job_once] etcd put job_id :%d error:%s", req.JobId, err.Error()))
		resp.FailWithMessage(resp.ERROR, "[job_once] put  error", c)
		return
	}
	resp.OkWithMessage("job once success", c)
}

// @Summary 终止运行中的任务
// @Description 根据任务ID和节点UUID终止运行中的任务进程
// @Tags job
// @Accept  json
// @Produce  json
// @Param   body  body   request.ReqJobKill  true  "Job kill details"
// @Success 200 {object} resp.Response "Job kill success"
// @Failure 400 {object} resp.Response "Bad request"
// @Failure 500 {object} resp.Response "Internal server error"
// @Router /job/kill [post]
func (j *JobRouter) Kill(c *gin.Context) {
	var req request.ReqJobKill
	var err error
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[job_once] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[job_once] request parameter error", c)
		return
	}
	resps, err := etcdclient.Get(fmt.Sprintf(etcdclient.KeyEtcdJobProcProfile, req.NodeUUID, req.JobId), clientv3.WithPrefix())
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[job_kill] etcd get error: %s", err.Error()))
		resp.FailWithMessage(resp.ERROR, "[job_kill] etcd get error", c)
		return
	}
	if len(resps.Kvs) == 0 {
		resp.FailWithMessage(resp.ERROR, "[job_kill] don't have such process", c)
		return
	}
	for _, p := range resps.Kvs {
		var proc models.JobProcVal
		if err := json.Unmarshal(p.Value, &proc); err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("job_kill[%s] unmarshal error: %s", string(p.Key), err.Error()))
			continue
		}
		if proc.Killed {
			continue
		}
		proc.Killed = true
		b, err := json.Marshal(&proc)
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("job_kill[%s] marshal error: %s", string(p.Key), err.Error()))
			continue
		}
		_, err = etcdclient.Put(string(p.Key), string(b))
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("job_kill[%s] etcd put error: %s", string(p.Key), err.Error()))
			continue
		}
	}
	resp.OkWithMessage("job kill success", c)
}
