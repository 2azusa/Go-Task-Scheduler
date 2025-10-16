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

// NodeServer 代表一个工作节点服务实例
type NodeServer struct {
	*etcdclient.ServerReg // 内嵌 Etcd 服务注册器，用于节点的注册和心跳维持
	*models.Node          // 内嵌节点的信息模型
	*cron.Cron            // 内嵌定时任务调度器实例

	jobs handler.Jobs // 当前节点上运行的任务集合
}

// NewNodeServer 创建并初始化一个新的 NodeServer 实例
func NewNodeServer() (*NodeServer, error) {
	// 生成唯一的 uuid
	uuid, err := utils.UUID()
	if err != nil {
		return nil, err
	}
	// 获取本地 IP 地址
	ip, err := utils.LocalIP()
	if err != nil {
		return nil, err
	}
	// 获取主机名
	hostname, err := os.Hostname()
	if err != nil {
		hostname = uuid
		err = nil
	}
	return &NodeServer{
		Node: &models.Node{
			UUID:     uuid,
			PID:      strconv.Itoa(os.Getpid()), // 获取当前进程的 PID
			IP:       ip.String(),
			Hostname: hostname,
			UpTime:   time.Now().Unix(),
			Status:   models.NodeConnSuccess, // 初始状态为连接成功
			Version:  config.GetConfigModels().System.Version,
		},
		Cron:      cron.New(),                                                       // 创建一个新的 cron 调度器
		jobs:      make(handler.Jobs, 8),                                            // 初始化任务 map
		ServerReg: etcdclient.NewServerReg(config.GetConfigModels().System.NodeTtl), // 创建 Etcd 注册器，并设置 TTL
	}, nil

}

// exist 检查具有相同 UUID 的节点是否已经在 Etcd 中注册并且仍在运行
func (srv *NodeServer) exist(nodeUUID string) (pid int, err error) {
	// 从 Etcd 获取指定 UUID 的节点信息
	resp, err := etcdclient.Get(fmt.Sprintf(etcdclient.KeyEtcdNode, nodeUUID))
	if err != nil {
		return
	}

	// 如果 Etcd 中没有改键，说明节点不存在
	if len(resp.Kvs) == 0 {
		return -1, nil
	}

	// 从 Etcd 的 Value 中解析出 PID
	if pid, err = strconv.Atoi(string(resp.Kvs[0].Value)); err != nil {
		// 如果解析失败，说明 Etcd 中的数据有问题，删除这个无效的键
		if _, err = etcdclient.Delete(fmt.Sprintf(etcdclient.KeyEtcdNode, nodeUUID)); err != nil {
			return
		}
		return -1, nil
	}

	// 查找该 PID 对应的进程是否存在
	p, err := os.FindProcess(pid)
	if err != nil {
		return -1, nil
	}

	// 发送一个信号 0 来检查进程是否存活
	if p != nil && p.Signal(syscall.Signal(0)) == nil {
		return
	}
	return -1, nil
}

// Register 将当前节点注册到 Etcd，键为 /pulse/node/<node_uuid>
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
	//creates a new lease
	if err := srv.ServerReg.Register(fmt.Sprintf(etcdclient.KeyEtcdNode, srv.UUID), string(b)); err != nil {
		return err
	}
	return nil
}

// Stop 停止节点服务，执行清理工作
func (srv *NodeServer) Stop(i any) {
	srv.Down() // 更新数据库中节点和任务的状态

	// 从 Etcd 中删除节点的注册信息
	_, err := etcdclient.Delete(fmt.Sprintf(etcdclient.KeyEtcdNode, srv.UUID))
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("failed to delete etcd node[%s] key error:%s", srv.UUID, err.Error()))
	}
	// 从 Etcd 中删除系统信息获取的键
	_, err = etcdclient.Delete(fmt.Sprintf(etcdclient.KeyEtcdSystemGet, srv.UUID))
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("failed to delete system etcd node[%s] key error:%s", srv.UUID, err.Error()))
	}

	_ = srv.Client.Close() // 关闭 Etcd 客户端连接
	srv.Cron.Stop()        // 停止定时任务调度器
}

// Down 在节点实例停机后，更新 Mysql 中的存活信息
func (srv *NodeServer) Down() {
	srv.Status = models.NodeConnFail // 将节点状态更新为连接失败
	srv.DownTime = time.Now().Unix()
	err := srv.Node.Update() // 更新数据库中的节点记录
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("failed to update  node[%s] down  error:%s", srv.UUID, err.Error()))
	}
	// 将所有分配该节点的任务状态重置为 “未分配”，以便 Admin 服务可以重新调度它们
	err = dbclient.GetMysqlDB().Table(models.PulseJobTableName).Select("status").Where("run_on = ? ", srv.UUID).Updates(models.Job{
		Status: models.JobStatusNotAssigned,
	}).Error

	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("failed to update job on node[%s] down  error:%s", srv.UUID, err.Error()))
	}
}

// Run 启动节点服务的主逻辑
func (srv *NodeServer) Run() (err error) {
	// 使用 defer 确保在函数退出时执行 Stop 清理逻辑
	defer func() {
		if err != nil {
			srv.Stop(err)
		}
	}()

	// 从数据库加载分配给此节点的任务
	if err = srv.loadJobs(); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("node[%s] failed to load job error:%s", srv.UUID, err.Error()))
		return
	}
	// 将本节点插入到数据库
	insertId, err := srv.Node.Insert()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("failed to create node[%s] into db error:%s", srv.UUID, err.Error()))
		return
	}
	srv.Node.ID = insertId // 记录数据库自增 ID
	// 启动 cron 调度器
	srv.Cron.Start()
	// 启动多个 goroutine 在后台监听 Etcd 中的变化
	go srv.watchJobs()       // 监听任务的增删改查
	go srv.watchOnce()       // 监听一次性任务的触发
	go srv.watchSystemInfo() // 监听获取系统信息的请求
	return
}

// loadJobs 从数据库加载已分配给此节点的任务
func (srv *NodeServer) loadJobs() (err error) {
	defer func() {
		// 捕获可能发生的 panic，防止程序崩溃
		if r := recover(); r != nil {
			logger.GetLogger().Error(fmt.Sprintf("load jobs panic:%v", r))
		}
	}()
	// 从数据库捕获分配给本节点的任务列表
	jobs, err := handler.GetJobs(srv.UUID)
	if err != nil {
		return
	}

	if len(jobs) == 0 {
		return
	}
	srv.jobs = jobs

	// 遍历任务列表，并将它们添加到 cron 调度器中
	for _, j := range jobs {
		j.InitNodeInfo(models.JobStatusAssigned, srv.UUID, srv.Hostname, srv.IP)
		srv.addJob(j)
	}

	return
}

// addJob 将一个任务添加到 cron 调度器中
func (srv *NodeServer) addJob(j *handler.Job) {
	if err := j.Check(); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("job[%d] check error :%s", j.ID, err.Error()))
		return
	}

	// 如果任务类型是执行预设脚本
	if j.Type == models.JobTypeCmd {
		for _, id := range j.ScriptIDArray {
			script := &models.Script{ID: id}

			// 从数据库查找脚本
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

			// 执行预设脚本
			result, err := handler.RunPresetScript(script)
			if err != nil {
				logger.GetLogger().Warn(fmt.Sprintf("job[%d] run script[%d] error :%s", j.ID, id, err.Error()))
			}
			logger.GetLogger().Info(fmt.Sprintf("job[%d] run script[%d] result :%s", j.ID, id, result))
		}
	}

	// 创建任务的执行函数
	taskFunc := handler.CreateJob(j)
	if taskFunc == nil {
		logger.GetLogger().Error("Failed to create a task to process the Job. The task protocol was not supported")
		return
	}

	// 将任务函数和 cron 表达式添加到调度器
	err := panicToError(func() {
		srv.Cron.AddFunc(j.Spec, taskFunc, srv.jobCronName(j.ID))
	})
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("Failed to add a task to the scheduler#%v", err.Error()))
	}
}

// jobCronName 为 cron 任务生成一个唯一的名称
func (srv *NodeServer) jobCronName(jobId int) string {
	return fmt.Sprintf(srv.UUID+"/%d", jobId)
}

// modifyJob 修改一个已存在的任务
func (srv *NodeServer) modifyJob(j *handler.Job) {
	oldJob, ok := srv.jobs[j.ID]
	if !ok {
		// 如果任务不存在，则直接添加
		srv.addJob(j)
		return
	}
	srv.deleteJob(oldJob.ID)
	srv.addJob(j)
}

// deleteJob 从 cron 调度器和本地任务 map 中删除一个任务
func (srv *NodeServer) deleteJob(jobId int) {
	if _, ok := srv.jobs[jobId]; ok {
		srv.Cron.RemoveJob(srv.jobCronName(jobId))
		delete(srv.jobs, jobId)
		return
	}
}

// panicToError 是一个辅助函数，它执行一个函数并将其中的 panic 转换为 error 返回
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
