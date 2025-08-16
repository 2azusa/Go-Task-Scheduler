package service

import (
	"context"
	"encoding/json"
	"fmt"
	"pulse/admin/internal/model/request"
	"pulse/common/models"
	"pulse/common/pkg/config"
	"pulse/common/pkg/dbclient"
	"pulse/common/pkg/etcdclient"
	"pulse/common/pkg/logger"
	"pulse/common/pkg/notify"
	"pulse/common/pkg/utils"
	"strings"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/api/v3/mvccpb"
)

// NodeWatcherService 结构体定义了监视器服务
type NodeWatcherService struct {
	client   *etcdclient.Client     // etcd 客户端，用于etcd交互
	nodeList map[string]models.Node // 当前存活的节点列表，key为节点UUID，value为节点信息
	lock     sync.Mutex             // 互斥锁，保证nodeList的并发访问安全
}

// DefauleNodeWatcher 一个全局的节点监视器服务实例
var DefaultNodeWatcher *NodeWatcherService

// NewNodeWatcherService 用于创建一个节点监视器实例
func NewNodeWtacherService() *NodeWatcherService {
	return &NodeWatcherService{
		client:   etcdclient.GetEtcdClient(),   // 获取etcd客户端单例
		nodeList: make(map[string]models.Node), // 初始化节点列表map
	}
}

// Watch 函数用于启动节点监视服务
func (n *NodeWatcherService) Watch() error {
	// 1. 启动时首先获取一次 etcd 中所有已注册的节点信息
	resp, err := n.client.Get(context.Background(), etcdclient.KeyEtcdNodeProfile, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	// 提取并填充初始的节点列表
	_ = n.extractNodes(resp)

	// 2. 启动后台 goroutine 来持续监视etcd中节点的变化
	go n.watcher()
	return nil
}

// watcher 是一个后台运行的 goroutine，持续监控etcd中的节点变化
func (n *NodeWatcherService) watcher() {
	// 监听etcd中 `KeyEtcdNodeProfile` 前缀的键值对变化
	rch := n.client.Watch(context.Background(), etcdclient.KeyEtcdNodeProfile, clientv3.WithPrefix())
	// 循环处理从 watch 通道接收到的事件
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			// PUT：表示有新节点加入或已有节点信息更新
			case mvccpb.PUT:
				n.setNodeList(n.GetUUID(string(ev.Kv.Key)), string(ev.Kv.Value))
				// DELETE：表示有节点下线
			case mvccpb.DELETE:
				uuid := n.GetUUID(string(ev.Kv.Key))
				// 从内存中的节点到列表里删除该节点
				n.delNodeList(uuid)
				logger.GetLogger().Warn(fmt.Sprintf("pulse node[%s] DELETE event detected", uuid))
				// 从数据库中查找该节点的信息，用于后续通知
				node := &models.Node{UUID: uuid}
				err := node.FindByUUID()
				if err != nil {
					logger.GetLogger().Error(fmt.Sprintf("pulse node[%s] find by uuid error: %s", uuid, err.Error()))
					return
				}

				// 核心故障转移逻辑
				success, fail, err := n.FailOver(uuid)
				if err != nil {
					logger.GetLogger().Error(fmt.Sprintf("pulse node[%s] fail over error: %s", uuid, err.Error()))
					return
				}
				// 如果所有任务都成功转移，则从数据库中删除这个已下线的节点记录
				if fail.Count() == 0 {
					err = node.Delete()
					if err != nil {
						logger.GetLogger().Error(fmt.Sprintf("pulse node[%s] delete by uuid error: %s", uuid, err.Error()))
					}
				}
				// 默认通过邮件发送节点失活警告
				msg := &notify.Message{
					Type:      notify.NotifyTypeMail,
					IP:        fmt.Sprintf("%s:%s", node.IP, node.PID),
					Subject:   "节点失活警告",
					Body:      fmt.Sprintf("[Pulse Warning] pulse node[%s] in the cluster has filed, fail over success count: %d jobID are: %s, fail count: %d jobID are: %s", uuid, success.Count(), success.String(), fail.Count(), fail.String()),
					To:        config.GetConfigModels().Email.To, // 从配置中读取收件人
					OccurTime: time.Now().Format(utils.TimeFormatSecond),
				}

				// 异步发送通知
				go notify.Send(msg)
			}
		}
	}
}

// extractNodes 用于从etcd的响应中提取节点信息并填充到 nodeList
func (n *NodeWatcherService) extractNodes(resp *clientv3.GetResponse) []string {
	nodes := make([]string, 0)
	if resp == nil || resp.Kvs == nil {
		return nodes
	}
	for i := range resp.Kvs {
		if v := resp.Kvs[i].Value; v != nil {
			// 设置节点信息到内存 map 中
			n.setNodeList(n.GetUUID(string(resp.Kvs[i].Key)), string(resp.Kvs[i].Value))
			nodes = append(nodes, string(v))
		}
	}
	return nodes
}

// setNodeList 用于将一个节点信息添加到 nodeList 中，并触发未分配任务的重新分配
func (n *NodeWatcherService) setNodeList(key, val string) {
	var node models.Node
	// 解析从etcd获取的JSON字符串到Node结构体
	err := json.Unmarshal([]byte(val), &node)
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("discover node[%s] json error: %s", key, err.Error()))
		return
	}

	// 加锁以保证并发安全
	n.lock.Lock()
	n.nodeList[key] = node
	n.lock.Unlock()
	logger.GetLogger().Debug(fmt.Sprintf("discover node node[%s] with pid[%s]", key, val))

	// 等待一段时间，确保新节点完全启动并准备好接收任务
	time.Sleep(5 * time.Second)

	// 查找所有尚未分配给任何节点的任务
	jobs, err := DefaultJobService.GetNotAssignedJob()
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("discover node[%s], pid[%s] and get not asigned job err: %s", key, val, err.Error()))
		return
	}

	// 遍历这些未分配的任务，并尝试为它们分配节点
	for _, job := range jobs {
		err = job.Unmarshal() // 反序列化job的额外数据
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("assign unassigned job[%d] json nmarshal error: %s", job.ID, err.Error()))
			continue
		}
		oldUUID := job.RunOn // 记录旧的运行节点UUID
		// 自动选择一个最合适的节点来运行次任务
		nodeUUID := DefaultJobService.AutoAllocateNode()
		if nodeUUID == "" {
			// 如果自动分配失败，则字节分配给这个新上线的节点
			nodeUUID = key
		}

		// 执行分配操作
		err = n.assignJob(nodeUUID, &job)
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("assign unassigned job[%d] error: %s", job.ID, err.Error()))
			continue
		}

		// 如果任务转移成功，删除etcd中表示该任务在旧节点的键
		_, err = etcdclient.Delete(fmt.Sprintf(etcdclient.KeyEtcdJob, oldUUID, job.ID))
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("node[%s] job[%d] fail over etcd delete job error: %s", nodeUUID, job.ID, err.Error()))
			continue
		}
	}
}

// delNodeList 用于从 nodeList 中安全地删除一个节点
func (n *NodeWatcherService) delNodeList(key string) {
	n.lock.Lock()
	defer n.lock.Unlock()
	delete(n.nodeList, key)
	logger.GetLogger().Debug(fmt.Sprintf("delete node[%s]", key))
}

// List2Array 用于将内存中的节点列表（map的key）转换为字符串切片
func (n *NodeWatcherService) List2Array() []string {
	n.lock.Lock()
	defer n.lock.Unlock()
	nodes := make([]string, 0)

	for k := range n.nodeList {
		nodes = append(nodes, k)
	}
	return nodes
}

// Close 用于关闭服务
func (n *NodeWatcherService) Close() error {
	return nil
}

// GetUUID 从etcd的key中提取节点的UUID
func (n *NodeWatcherService) GetUUID(key string) string {
	// /pulse/node/<node_uuid>
	index := strings.LastIndex(key, "/")
	if index == -1 {
		return ""
	}
	return key[index+1:]
}

// Search 根据查询条件从数据库中搜索节点信息
func (n *NodeWatcherService) Search(s *request.ReqNodeSearch) ([]models.Node, int64, error) {
	db := dbclient.GetMysqlDB().Table(models.PulseNodeTableName)
	if len(s.UUID) > 0 {
		db = db.Where("uuid = ?", s.UUID)
	}
	if len(s.IP) > 0 {
		db.Where("ip = ?", s.IP)
	}
	if s.Status > 0 {
		db.Where("status = ?", s.Status)
	}
	if s.UpTime > 0 {
		db.Where("up > ?", s.UpTime)
	}
	nodes := make([]models.Node, 2) // 初始化为空切片，避免nil
	var total int64
	// 先计算总数，用于分页
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	// 再根据分页参数查询数据
	err = db.Limit(s.PageSize).Offset((s.Page - 1) * s.PageSize).Order("up desc").Find(&nodes).Error
	if err != nil {
		return nil, 0, err
	}

	return nodes, total, nil
}

// GetJobCount 用于获取指定节点上运行的任务数量
func (n *NodeWatcherService) GetJobCount(nodeUUID string) (int, error) {
	resps, err := etcdclient.Get(fmt.Sprintf(etcdclient.KeyEtcdJobProfile, nodeUUID), clientv3.WithPrefix(), clientv3.WithCountOnly())
	if err != nil {
		return 0, err
	}
	return int(resps.Count), nil
}

// Result 用于存储任务ID的结果集
type Result []int

// Count 用于计算结果集中非零元素的数量
func (r Result) Count() (count int) {
	return len(r)
}

// String 将结果格式化为字符串
func (r Result) String() (str string) {
	str = "["
	for _, v := range r {
		if v != 0 {
			str += fmt.Sprintf("%d", v)
		}
	}
	str += "]"
	return
}

// assignJob 用于将一个任务分配给指定的节点
func (n *NodeWatcherService) assignJob(nodeUUID string, job *models.Job) (err error) {
	if nodeUUID == "" {
		return fmt.Errorf("node uuid can't be null")
	}
	// 检查目标节点是否在当前存活的节点列表中
	node, ok := n.nodeList[nodeUUID]
	if !ok {
		return fmt.Errorf("assign unassigned job[%d] but node[%s] not exist", job.ID, nodeUUID)
	}
	// 更新任务的状态和运行节点信息
	job.InitNodeInfo(models.JobStatusAssigned, node.UUID, node.Hostname, node.IP)

	// 将更新后的任务信息序列化为JSON
	b, err := json.Marshal(job)
	if err != nil {
		return
	}
	// 将任务信息写入etcd，路径为 /pulse/job/<node_uuid>/<job_id>
	_, err = etcdclient.Put(fmt.Sprintf(etcdclient.KeyEtcdJob, nodeUUID, job.ID), string(b))
	if err != nil {
		return
	}
	// 更新数据库中的任务记录
	err = job.Update()
	if err != nil {
		return nil
	}
	return
}

// FailOver 用于执行故障转移逻辑：将一个失效节点上的所有任务重新分配到其他健康节点上
func (n *NodeWatcherService) FailOver(nodeUUID string) (success Result, fail Result, err error) {
	// 1. 获取该失效节点上的所有任务
	jobs, err := n.GetJobs(nodeUUID)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("node[%s] fail over get jobs error: %s", nodeUUID, err.Error()))
		return
	}
	if len(jobs) == 0 {
		return
	}
	// 2. 遍历所有任务，并为每个任务重新分配节点
	for _, job := range jobs {
		oldUUID := job.RunOn
		// 自动选择一个新的、合适的节点
		autoUUID := DefaultJobService.AutoAllocateNode()
		if autoUUID == "" {
			logger.GetLogger().Warn(fmt.Sprintf("node[%s] job[%d] fail over auto allocate node error", nodeUUID, job.ID))
			fail = append(fail, job.ID)
			continue
		}
		// 将任务分配给新节点
		err = n.assignJob(autoUUID, &job)
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("node[%s] job[%d] fail over assign job error", nodeUUID, job.ID))
			fail = append(fail, job.ID)
			continue
		}
		// 3. 转移成功后，删除etcd中旧的键值对(/pulse/job/<old_uuid>/<job_id>)
		_, err := etcdclient.Delete(fmt.Sprintf(etcdclient.KeyEtcdJob, oldUUID, job.ID))
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("node[%s] job[%d] fail over etcd delete job error", nodeUUID, job.ID))
			fail = append(fail, job.ID)
			continue
		}
		// 记录为成功转移
		success = append(success, job.ID)
	}
	return
}

// GetJobs 用于获取一个指定节点(nodeUUID)下的所有任务
func (n *NodeWatcherService) GetJobs(nodeUUID string) (jobs []models.Job, err error) {
	// 从etcd中获取指定前缀(/pulse/job/profile/<node_uuid>/)下的所有键值对
	resps, err := etcdclient.Get(fmt.Sprintf(etcdclient.KeyEtcdJobProfile, nodeUUID), clientv3.WithPrefix())
	if err != nil {
		return
	}
	count := len(resps.Kvs)
	if count == 0 {
		return
	}
	// 遍历返回的键值对
	for _, j := range resps.Kvs {
		var job models.Job
		// 将value(json字符串)反序列化为Job结构体
		if err := json.Unmarshal(j.Value, &job); err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("job[%s] umarshal err: %s", string(j.Key), err.Error()))
			continue
		}
		jobs = append(jobs, job)
	}
	return
}

// GetNodeCount 用于获取数据库中指定状态的节点数量
func (n *NodeWatcherService) GetNodeCount(status int) (int64, error) {
	db := dbclient.GetMysqlDB().Table(models.PulseNodeTableName)
	if status > 0 {
		// 如果status > 0，则添加status的查询条件
		db = db.Where("status = ?", status)
	}
	var total int64
	// 执行COUNT查询
	err := db.Count(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

// GetNodeSystemInfo 用于从etcd中获取指定节点的实时系统信息
func GetNodeSystemInfo(uuid string) (s *utils.Server, err error) {
	// defer语句确保在函数返回前，删除用于获取信息的触发key，防止重复获取
	defer func() {
		// 删除 /pulse/system/switch/<uuid> 这个key
		_, err = etcdclient.Delete(fmt.Sprintf(etcdclient.KeyEtcdSystemSwitch, uuid))
	}()
	s = new(utils.Server)
	// 从etcd中读取节点写入的系统信息，key格式为/pulse/system/get/<uuid>
	res, err := etcdclient.Get(fmt.Sprintf(etcdclient.KeyEtcdSystemGet, uuid), clientv3.WithPrefix())
	if err != nil || len(res.Kvs) == 0 {
		return
	}
	// 将返回的JSON数据反序列化到Server结构体
	err = json.Unmarshal(res.Kvs[0].Value, s)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("json error: %v", err))
	}
	return
}
