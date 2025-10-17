package service

import (
	"fmt"
	"pulse/admin/internal/model/request"
	"pulse/admin/internal/model/resp"
	"pulse/common/models"
	"pulse/common/pkg/dbclient"
	"pulse/common/pkg/etcdclient"
	"pulse/common/pkg/logger"
	"pulse/common/pkg/utils"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// JobService 定义了任务相关的服务
type JobService struct{}

// DefaultJobService 是 JobService 的默认实例
var DefaultJobService = new(JobService)

// Serach 用于根据查询条件搜索任务
func (j *JobService) Search(s *request.ReqJobSearch) ([]models.Job, int64, error) {
	// 获取数据可连接并指定表名
	db := dbclient.GetMysqlDB().Table(models.PulseJobTableName)
	// 如果请求中包含ID， 则添加ID的查询条件
	if s.ID > 0 {
		db = db.Where("id = ?", s.ID)
	}
	// 如果请求中包含名称，则添加名称的模糊查询条件
	if len(s.Name) > 0 {
		db = db.Where("name like ?", s.Name+"%")
	}
	// 如果请求中包含运行节点，则添加运行节点的查询条件
	if len(s.RunOn) > 0 {
		db.Where("run_on = ?", s.RunOn)
	}
	// 如果请求中包含类型，则添加类型的查询条件
	if s.Type > 0 {
		db.Where("type = ?", s.Type)
	}
	// 如果请求中包含状态，则添加状态的查询条件
	if s.Status > 0 {
		db.Where("status = ?", s.Status)
	}
	// 创建一个用于存储的任务切片
	jobs := make([]models.Job, 2)
	var total int64
	// 计算符合条件的总任务
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	// 根据分页参数查询任务列表
	err = db.Limit(s.PageSize).Offset((s.Page - 1) * s.PageSize).Find(&jobs).Error
	if err != nil {
		return nil, 0, err
	}
	return jobs, total, nil
}

// SearchJobLog 用于根据条件搜索任务日志
func (j *JobService) SearchJobLog(s *request.ReqJobLogSearch) ([]models.JobLog, int64, error) {
	// 获取数据库并连接指定表名
	db := dbclient.GetMysqlDB().Table(models.PulseJobLogTableName)
	// 如果请求中包含名称，则添加名称的模糊查询条件
	if len(s.Name) > 0 {
		db = db.Where("name like ?", s.Name+"%")
	}
	// 如果请求中包含任务ID，则添加任务ID的查询条件
	if s.JobId > 0 {
		db.Where("job_id = ?", s.JobId)
	}
	// 如果请求中包含节点的UUID，则添加节点UUID的查询条件
	if len(s.NodeUUID) > 0 {
		db.Where("node_uuid = ?", s.NodeUUID)
	}
	// 如果请求中包含是否成功的标识，则添加相应的查询条件
	if s.Success != nil {
		db.Where("success = ?", *s.Success)
	}
	// 创建一个用于存储任务日志的切片
	jobLogs := make([]models.JobLog, 2)
	var total int64
	// 计算符合条件的总日志数
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	// 根据分页参数并按开始时间降序查询日志列表
	err = db.Limit(s.PageSize).Offset((s.Page - 1) * s.PageSize).Order("start_time desc").Find(&jobLogs).Error
	if err != nil {
		return nil, 0, err
	}

	return jobLogs, total, nil
}

// GetToDayJobExcCount 获取今天执行的任务总数
func (j *JobService) GetTodayJobExcCount(success int) (int64, error) {
	// 查询今天开始时间之后、已结束且执行结果符合条件的任务日志
	db := dbclient.GetMysqlDB().Table(models.PulseJobLogTableName).Where("start_time > ? and end_time != 0 and success =?", utils.GetTodayUnix(), success)
	var total int64
	// 计算总数
	err := db.Count(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

// GetJobExcCount 获取在特定时间段内的每天任务执行数量
func (j *JobService) GetJobExcCount(start, end int64, success int) ([]resp.DateCount, error) {
	var dateCount []resp.DateCount
	// 查询在指定时间段内、已结束且执行结果符合条件的任务，并按天分组计数
	db := dbclient.GetMysqlDB().Table(models.PulseJobLogTableName).Select("FROM_UNIXTIME(start_time, '%Y-%m-%d') AS date", "COUNT(*) AS count").Group("date").Order("date ASC").Where("start_time > ? and start_time < ? and end_time != 0 and success = ?", start, end, success)
	err := db.Find(&dateCount).Error
	if err != nil {
		return nil, err
	}
	return dateCount, nil
}

// GetNotAssignedJob 获取所有未分配的任务
func (j *JobService) GetNotAssignedJob() (jobs []models.Job, err error) {
	// 查询状态为“未分配”的任务
	err = dbclient.GetMysqlDB().Table(models.PulseJobTableName).Where("status = ?", models.JobStatusNotAssigned).Find(&jobs).Error
	return
}

// GetRunningJobCount 获取当前正在执行的任务数量
func (j *JobService) GetRunningJobCount() (int64, error) {
	// 从 etcd 中获取正在执行的任务进程数量
	wresp, err := etcdclient.Get(fmt.Sprintf(etcdclient.KeyEtcdProcProfile), clientv3.WithPrefix(), clientv3.WithCountOnly())
	if err != nil {
		return 0, err
	}

	return wresp.Count, nil
}

// MaxJobCount 定义一个节点上最大的任务数量，用于节点分配算法
const MaxJobCount = 10000

// Give priority to the node with the least number of tasks
// AutoAllocateNode 自动分配节点，优先选择任务最小的节点
func (j *JobService) AutoAllocateNode() string {
	// Get all the living nodes
	// 获取所有存活的节点列表
	nodeList := DefaultNodeWatcher.List2Array()
	resultCount, resultNodeUUID := MaxJobCount, ""
	// 遍历所有节点
	for _, nodeUUID := range nodeList {
		// Check the datebase to see if it is alive
		// 从数据库中检查节点是否真的存在
		node := &models.Node{UUID: nodeUUID}
		err := node.FindByUUID()
		if err != nil {
			continue
		}
		// 如果节点连接失败
		if node.Status == models.NodeConnFail {
			// The node has failed
			// 将该节点从监视器列表中删除
			delete(DefaultNodeWatcher.nodeList, nodeUUID)
			continue
		}
		// 获取该节点上的任务数量
		count, err := DefaultNodeWatcher.GetJobCount(nodeUUID)
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("node[%s] get job count error: %s", nodeUUID, err.Error()))
			continue
		}
		// 如果当前节点的任务数量比记录的要少，则更新结束
		if resultCount > count {
			resultCount, resultNodeUUID = count, nodeUUID
		}
	}

	return resultNodeUUID
}

// Once 用于标记一个任务在特定节点上执行一次
func (j *JobService) Once(once *request.ReqJobOnce) (err error) {
	//The default existence time is 60 seconds
	_, err = etcdclient.PutWithTtl(fmt.Sprintf(etcdclient.KeyEtcdOnce, once.JobId), once.NodeUUID, 60)
	return
}

// RunLogCleaner 启动一个 goroutine 来定期清理就的日志
func RunLogCleaner(cleanPeriod time.Duration, expiration int64) (close chan struct{}) {
	// 创建一个定时器
	t := time.NewTicker(cleanPeriod)
	close = make(chan struct{})
	// 启动一个 goroutine
	go func() {
		for {
			select {
			// 定时器触发时
			case <-t.C:
				// 清理过期的日志
				err := cleanupLogs(expiration)
				if err != nil {
					logger.GetLogger().Error(fmt.Sprintf("clean up logs at time: %v error: %s", time.Now(), err.Error()))
				}
			// 接受关闭信号时
			case <-close:
				// 停止定时器并返回
				t.Stop()
				return
			}
		}
	}()
	return
}

// cleanupLogs 删除指定时间戳之前的日志
func cleanupLogs(expirationTime int64) error {
	// 构造删除语句
	sql := fmt.Sprintf("delete from %s where start_time < ?", models.PulseJobLogTableName)
	// 执行删除操作
	return dbclient.GetMysqlDB().Exec(sql, time.Now().Unix()-expirationTime).Error
}
