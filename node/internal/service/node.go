package service

import (
	"encoding/json"
	"fmt"
	"os"
	"pulse/common/models"
	"pulse/common/pkg/config"
	"pulse/common/pkg/dbclient"
	"pulse/common/pkg/etcdclient"
	"pulse/common/pkg/logger"
	"pulse/common/pkg/utils"
	"pulse/node/internal/handler"
	"strconv"
	"syscall"
	"time"

	"github.com/jakecoffman/cron"
)

// NodeServer 表示一个工作节点实例
type NodeServer struct {
	*etcdclient.ServerReg // 嵌入式 Etcd 服务器注册表，用于节点注册和心跳
	*models.Node          // 嵌入式节点信息模型
	*cron.Cron            // 嵌入式 cron 调度程序实例

	jobs handler.Jobs // 在当前节点上运行的任务集合
}

// NewNodeServer 创建并初始化一个新的 NodeServer 实例
func NewNodeServer() (*NodeServer, error) {
	uuid, err := utils.UUID()
	if err != nil {
		return nil, err
	}
	ip, err := utils.LocalIP()
	if err != nil {
		return nil, err
	}
	hostname, err := os.Hostname()
	if err != nil {
		hostname = uuid
		err = nil
	}
	return &NodeServer{
		Node: &models.Node{
			UUID:     uuid,
			PID:      strconv.Itoa(os.Getpid()),
			IP:       ip.String(),
			Hostname: hostname,
			UpTime:   time.Now().Unix(),
			Status:   models.NodeConnSuccess,
			Version:  config.GetConfigModels().System.Version,
		},
		Cron:      cron.New(),
		jobs:      make(handler.Jobs, 8),
		ServerReg: etcdclient.NewServerReg(config.GetConfigModels().System.NodeTtl),
	}, nil

}

// exist 检查具有相同 UUID 的节点是否已在 Etcd 中注册并正在运行
func (srv *NodeServer) exist(nodeUUID string) (pid int, err error) {
	resp, err := etcdclient.Get(fmt.Sprintf(etcdclient.KeyEtcdNode, nodeUUID))
	if err != nil {
		return
	}

	if len(resp.Kvs) == 0 {
		return -1, nil
	}

	if pid, err = strconv.Atoi(string(resp.Kvs[0].Value)); err != nil {
		if _, err = etcdclient.Delete(fmt.Sprintf(etcdclient.KeyEtcdNode, nodeUUID)); err != nil {
			return
		}
		return -1, nil
	}

	p, err := os.FindProcess(pid)
	if err != nil {
		return -1, nil
	}

	// 向进程发送信号 0 是检查其是否处于活动状态的标准方法
	if p != nil && p.Signal(syscall.Signal(0)) == nil {
		return
	}
	return -1, nil
}

// Register 将当前节点注册到 Etcd
func (srv *NodeServer) Register() error {
	pid, err := srv.exist(srv.UUID)
	if err != nil {
		return err
	}
	if pid != -1 {
		return fmt.Errorf("node[%s] with pid[%d] exist", srv.UUID, pid)
	}
	b, err := json.Marshal(&srv.Node)
	if err != nil {
		return fmt.Errorf("node[%s] with pid[%d] json error%s", srv.UUID, pid, err.Error())
	}
	// 使用租约（TTL）将服务器注册到 etcd
	if err := srv.ServerReg.Register(fmt.Sprintf(etcdclient.KeyEtcdNode, srv.UUID), string(b)); err != nil {
		return err
	}
	return nil
}

// Stop 停止节点服务器并执行清理
func (srv *NodeServer) Stop(i any) {
	srv.Down()

	_, err := etcdclient.Delete(fmt.Sprintf(etcdclient.KeyEtcdNode, srv.UUID))
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("failed to delete etcd node[%s] key error:%s", srv.UUID, err.Error()))
	}
	_, err = etcdclient.Delete(fmt.Sprintf(etcdclient.KeyEtcdSystemGet, srv.UUID))
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("failed to delete system etcd node[%s] key error:%s", srv.UUID, err.Error()))
	}

	_ = srv.Client.Close()
	srv.Cron.Stop()
}

// Down 在节点关闭时更新数据库中的节点和任务状态
func (srv *NodeServer) Down() {
	srv.Status = models.NodeConnFail
	srv.DownTime = time.Now().Unix()
	err := srv.Node.Update()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("failed to update  node[%s] down  error:%s", srv.UUID, err.Error()))
	}
	// 将分配给此节点的所有任务的状态重置为“NotAssigned”
	// 以便管理服务可以重新调度它们
	err = dbclient.GetMysqlDB().Table(models.PulseJobTableName).Select("status").Where("run_on = ? ", srv.UUID).Updates(models.Job{
		Status: models.JobStatusNotAssigned,
	}).Error

	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("failed to update job on node[%s] down  error:%s", srv.UUID, err.Error()))
	}
}

// Run 启动节点服务器的主要逻辑
func (srv *NodeServer) Run() (err error) {
	defer func() {
		if err != nil {
			srv.Stop(err)
		}
	}()

	if err = srv.loadJobs(); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("node[%s] failed to load job error:%s", srv.UUID, err.Error()))
		return
	}

	insertId, err := srv.Node.Insert()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("failed to create node[%s] into db error:%s", srv.UUID, err.Error()))
		return
	}
	srv.Node.ID = insertId

	srv.Cron.Start()

	go srv.watchJobs()
	go srv.watchOnce()
	go srv.watchSystemInfo()
	return
}

// loadJobs 从数据库加载分配给此节点的任务
func (srv *NodeServer) loadJobs() (err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.GetLogger().Error(fmt.Sprintf("load jobs panic:%v", r))
		}
	}()

	jobs, err := handler.GetJobs(srv.UUID)
	if err != nil {
		return
	}

	if len(jobs) == 0 {
		return
	}
	srv.jobs = jobs

	for _, j := range jobs {
		j.InitNodeInfo(models.JobStatusAssigned, srv.UUID, srv.Hostname, srv.IP)
		srv.addJob(j)
	}

	return
}

// addJob 将任务添加到 cron 调度程序
func (srv *NodeServer) addJob(j *handler.Job) {
	if err := j.Check(); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("job[%d] check error :%s", j.ID, err.Error()))
		return
	}

	if j.Type == models.JobTypeCmd {
		for _, id := range j.ScriptIDArray {
			script := &models.Script{ID: id}

			err := script.FindById()
			if err != nil {
				logger.GetLogger().Warn(fmt.Sprintf("job[%d] find script[%d] error :%s", j.ID, id, err.Error()))
				continue
			}

			err = script.Check()
			if err != nil {
				logger.GetLogger().Warn(fmt.Sprintf("script[%d] check error :%s", id, err.Error()))
				continue
			}

			result, err := handler.RunPresetScript(script)
			if err != nil {
				logger.GetLogger().Warn(fmt.Sprintf("job[%d] run script[%d] error :%s", j.ID, id, err.Error()))
			}
			logger.GetLogger().Info(fmt.Sprintf("job[%d] run script[%d] result :%s", j.ID, id, result))
		}
	}

	taskFunc := handler.CreateJob(j)
	if taskFunc == nil {
		logger.GetLogger().Error("Failed to create a task to process the Job. The task protocol was not supported")
		return
	}

	err := panicToError(func() {
		srv.Cron.AddFunc(j.Spec, taskFunc, srv.jobCronName(j.ID))
	})
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("Failed to add a task to the scheduler#%v", err.Error()))
	}
}

// jobCronName 生成 cron 任务的唯一名称
func (srv *NodeServer) jobCronName(jobId int) string {
	return fmt.Sprintf(srv.UUID+"/%d", jobId)
}

// modifyJob 修改调度程序中的现有任务
func (srv *NodeServer) modifyJob(j *handler.Job) {
	oldJob, ok := srv.jobs[j.ID]
	if !ok {
		srv.addJob(j)
		return
	}
	srv.deleteJob(oldJob.ID)
	srv.addJob(j)
}

// deleteJob removes a job from the cron scheduler and the local job map.
func (srv *NodeServer) deleteJob(jobId int) {
	if _, ok := srv.jobs[jobId]; ok {
		srv.Cron.RemoveJob(srv.jobCronName(jobId))
		delete(srv.jobs, jobId)
		return
	}
}

// panicToError is a helper function that executes a function and converts any panic to an error.
func panicToError(fn func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				err = v
			case string:
				err = fmt.Errorf("panic: %s", v)
			default:
				err = fmt.Errorf("panic: %v", v)
			}
		}
	}()

	fn()
	return nil
}
