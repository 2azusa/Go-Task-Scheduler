package handler

import (
	"crony/admin/internal/model/request"
	"crony/admin/internal/model/resp"
	"crony/admin/internal/service"
	"crony/common/models"
	"crony/common/pkg/config"
	"crony/common/pkg/etcdclient"
	"crony/common/pkg/logger"
	"encoding/json"
	"fmt"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/gin-gonic/gin"
)

// JobRouter 结构体用于组织与任务相关的路有处理器
type JobRouter struct{}

// defaultJobRouter 是JobRouter的一个默认实例
var defaultJobRouter = new(JobRouter)

// CreateOrUpdate 用于创建或更新任务的HTTP请求
func (j *JobRouter) CreateOrUpdate(c *gin.Context) {
	var req request.ReqJobUpdate
	// 1. 绑定并解析请求体中的JSON数据到req结构体
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[create_job] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[create_job] request parameter error", c)
		return
	}
	// 2. 验证请求参数的合法性
	if err := req.Valid(); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("create_job check error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorJobFormat, "[create_job] check error", c)
		return
	}

	var err error
	var insertId int
	t := time.Now()

	// 3. 处理节点分配逻辑
	switch req.Allocation {
	case models.AutoAllocation:
		// 自动分配
		if !config.GetConfigModels().System.CmdAutoAllocation && req.Type == models.JobTypeCmd {
			resp.FailWithMessage(resp.ERROR, "[create_job] The shell command is not supported to automtically assign nodes by default", c)
			return
		}
		// 调用服务层的方法来自动选择一个合适的节点
		nodeUUID := service.DefaultJobService.AutoAllocateNode()
		if nodeUUID == "" {
			logger.GetLogger().Error("[create_job] auto allocate node error")
			resp.FailWithMessage(resp.ERROR, "[create_job] auto allocate node error", c)
			return
		}
		req.RunOn = nodeUUID // 设置任务在哪个节点上运行
	case models.ManualAllocation:
		// 手动分配
		if len(req.RunOn) == 0 {
			resp.FailWithMessage(resp.ERROR, "[create_job] manually assigned node can't be null", c)
			return
		}
		// 检查手动指定的节点是否处于活动状态
		node := &models.Node{UUID: req.RunOn}
		_ = node.FindByUUID()
		if node.Status == models.NodeConnFail {
			resp.FailWithMessage(resp.ERROR, "[create_job] manually assigned node inactiveation", c)
			return
		}
	}

	// 4. 根据请求中是否包含ID来判断是创建还是更新操作
	if req.ID > 0 {
		// 更新操作
		job := &models.Job{ID: req.ID}
		_ = job.FindById()
		oldNodeUUID := job.RunOn // 获取任务更新前的运行节点UUID
		if oldNodeUUID != "" {
			// 从etcd中删除旧的Job信息
			_, err = etcdclient.Delete(fmt.Sprintf(etcdclient.KeyEtcdJob, oldNodeUUID, req.ID))
			if err != nil {
				logger.GetLogger().Error(fmt.Sprintf("[update_job] delete etcd node[%s] error: %s", oldNodeUUID, err.Error()))
				resp.FailWithMessage(resp.ERROR, "[update_job] delete etcd node error", c)
				return
			}
		}
		// 设置更新时间
		req.Updated = t.Unix()
		// 关系数据库中的任务记录
		err = req.Update()
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("[update_job] into db error: %s", err.Error()))
			resp.FailWithMessage(resp.ERROR, "[update_job] into db error", c)
			return
		}
	} else {
		// 创建操作
		req.Created = t.Unix()
		// 设置创建时间
		insertId, err = req.Insert()
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("[create_job] json marshal job error: %s", err.Error()))
			resp.FailWithMessage(resp.ERROR, "[create_job] instert job into db error", c)
			return
		}
		// 将生成的ID赋值给请求对象
		req.ID = insertId
	}

	// 5. 将任务信息写入etcd以便目标节点能够监听到并执行
	b, err := json.Marshal(req)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[create_job] json marshal job error: %s", err.Error()))
		resp.FailWithMessage(resp.ERROR, "[create_job] json marshal job error", c)
		return
	}
	// 将任务数据Put到etcd， key格式为 /crony/job/<node_uuid>/<job_id>
	_, err = etcdclient.Put(fmt.Sprintf(etcdclient.KeyEtcdJob, req.RunOn, req.ID), string(b))
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[create_job] etcd put job error:%s", err.Error()))
		resp.FailWithMessage(resp.ERROR, "[create_job] etcd put job error", c)
		return
	}

	// 6. 返回成功响应
	resp.OkWithDetailed(req, "operate success", c)
}

// Delete 用于删除一个或多个任务
func (j *JobRouter) Delete(c *gin.Context) {
	var req request.ByIDS
	// 绑定请求体中的ID列表
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[delete_job] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[delete_job] request parameter error", c)
		return
	}
	// 遍历所有要删除的任务ID
	for _, id := range req.IDS {
		job := models.Job{ID: id}
		// 从数据库中查找任务，获取其运行在哪个节点
		err := job.FindById()
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("[delete_job] find job by id: %d error: %s", id, err.Error()))
			continue
		}
		// 从etcd中删除任务，通知对应的节点停止执行该任务
		_, err = etcdclient.Delete(fmt.Sprintf(etcdclient.KeyEtcdJob, job.RunOn, id))
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("[delete_job] etcd delete job error: %s", err.Error()))
			continue
		}
		// 从数据库中删除任务记录
		err = job.Delete()
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("[delete_job] into db error: %s", err.Error()))
			continue
		}
	}
	resp.OkWithMessage("delete success", c)
}

// FindById 用于根据ID查找单个任务
func (j *JobRouter) FindById(c *gin.Context) {
	var req request.ByID
	// 从URL查询参数中绑定ID
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[find_job] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[find_job] request parameter error", c)
		return
	}
	job := models.Job{ID: req.ID}
	// 从数据库中查找任务
	err := job.FindById()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[find_job] find job by id: %d error: %s", req.ID, err.Error()))
		resp.FailWithMessage(resp.ERROR, "[find_job] find job by id error", c)
		return
	}
	// 如果任务有额外的JSON格式数据，则进行反序列化
	if len(job.NotifyTo) != 0 {
		_ = job.Unmarshal()
	}
	resp.OkWithDetailed(job, "find success", c)
}

// Search 用于搜索任务
func (j *JobRouter) Search(c *gin.Context) {
	var req request.ReqJobSearch
	// 绑定请求体中的搜索条件
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[search_job] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[search_job] request parameter error", c)
		return
	}
	// 检查并设置默认分页参数
	req.Check()
	// 调用服务层的方法执行搜索
	jobs, total, err := service.DefaultJobService.Search(&req)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[serach_job] search job error: %s", err.Error()))
		resp.FailWithMessage(resp.ERROR, "[serach_job] search job error", c)
		return
	}
	var resultJobs []models.Job
	// 遍历搜索条件，反序列化每个任务的附加JSON数据
	for _, job := range jobs {
		_ = job.Unmarshal()
		resultJobs = append(resultJobs, job)
	}
	// 返回分页格式的响应数据
	resp.OkWithDetailed(resp.PageResult{
		List:     resultJobs,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, "search success", c)
}

// SearchLog 用于搜索任务执行日志
func (j *JobRouter) SearchLog(c *gin.Context) {
	var req request.ReqJobLogSearch
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[search_job_log] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[search_job_log] request parameter error", c)
		return
	}
	// 检查并设置默认分页参数
	req.Check()
	// 调用服务层方法搜索日志
	jobs, total, err := service.DefaultJobService.SearchJobLog(&req)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[search_job_log] db error: %s", err.Error()))
		resp.FailWithMessage(resp.ERROR, "[serach_job_log] db error", c)
		return
	}
	// 返回分页格式的响应数据
	resp.OkWithDetailed(resp.PageResult{
		List:     jobs,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, "search success", c)
}

// Once 用于处理立即执行一个任务的请求
func (j *JobRouter) Once(c *gin.Context) {
	var req request.ReqJobOnce
	var err error
	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[job_once] request parameter error:%s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[job_once] request parameter error", c)
		return
	}
	// 查找目标节点是否存在且在线
	node := &models.Node{UUID: req.NodeUUID}
	err = node.FindByUUID()
	if err != nil || node.Status == models.NodeConnFail {
		logger.GetLogger().Error(fmt.Sprintf("[job_once] node[%s] conn fail:%v", req.NodeUUID, err))
		resp.FailWithMessage(resp.ERROR, "[job_once] node conn fail ", c)
		return
	}
	// 确认指定任务在数据库中存在
	job := &models.Job{ID: req.JobId}
	err = job.FindById()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[job_once] job_id[%d] not exist db:%s", req.JobId, err.Error()))
		resp.FailWithMessage(resp.ERROR, "[job_once] job not exist ", c)
		return
	}

	// 调用服务层的方法
	err = service.DefaultJobService.Once(&req)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[job_once] etcd put job_id :%d error:%s", req.JobId, err.Error()))
		resp.FailWithMessage(resp.ERROR, "[job_once] put  error", c)
		return
	}
	resp.OkWithMessage("job once success", c)
}

// Kill 用于处理强行终止一个正在执行的任务的请求
func (j *JobRouter) Kill(c *gin.Context) {
	var req request.ReqJobKill
	var err error
	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[job_once] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[job_once] request parameter error", c)
		return
	}
	// 从etcd中查找该任务正在运行的进程信息
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
	// 遍历所有找到的进程
	for _, p := range resps.Kvs {
		var proc models.JobProcVal
		// 反序列化从etcd中获取进程信息
		if err := json.Unmarshal(p.Value, &proc); err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("job_kill[%s] unmarshal error: %s", string(p.Key), err.Error()))
			continue
		}
		// 进程已经被杀死
		if proc.Killed {
			continue
		}
		proc.Killed = true
		b, err := json.Marshal(&proc)
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("job_kill[%s] marshal error: %s", string(p.Key), err.Error()))
			continue
		}
		// 修改
		_, err = etcdclient.Put(string(p.Key), string(b))
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("job_kill[%s] etcd put error: %s", string(p.Key), err.Error()))
			continue
		}
	}
	// 返回成功信息
	resp.OkWithMessage("job kill success", c)
}
