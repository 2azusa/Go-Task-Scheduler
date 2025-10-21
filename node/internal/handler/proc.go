package handler

import (
	"encoding/json"
	"fmt"
	"pulse/common/models"
	"pulse/common/pkg/config"
	"pulse/common/pkg/etcdclient"
	"pulse/common/pkg/logger"
	"strconv"
	"strings"
	"sync/atomic"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// JobProc 有关正在运行的任务进程的信息
type JobProc struct {
	*models.JobProc
}

// GetProcFromKey 从 etcd key 中解析 JobProc 信息
// /pulse/proc/{nodeUUID}/{jobId}/{procId}
func GetProcFromKey(key string) (proc *JobProc, err error) {
	ss := strings.Split(key, "/")
	var sslen = len(ss)
	if sslen < 5 {
		err = fmt.Errorf("invalid proc key [%s]", key)
		return
	}
	id, err := strconv.Atoi(ss[sslen-1])
	if err != nil {
		return
	}
	jobId, err := strconv.Atoi(ss[sslen-2])
	if err != nil {
		return
	}
	proc = &JobProc{
		JobProc: &models.JobProc{
			ID:       id,
			JobID:    jobId,
			NodeUUID: ss[sslen-3],
		},
	}
	return
}

// Key 为 JobProc 实例生成唯一的 etcd key
func (p *JobProc) Key() string {
	return fmt.Sprintf(etcdclient.KeyEtcdProc, p.NodeUUID, p.JobID, p.ID)
}

// del 从 etcd 中删除进程的 key
func (p *JobProc) del() error {
	_, err := etcdclient.Delete(p.Key())
	return err
}

// Start 从 etcd 中注册一个正在运行的进程，这是线程安全的
func (p *JobProc) Start() error {
	// 使用原子操作确保Start只执行一次
	if !atomic.CompareAndSwapInt32(&p.Running, 0, 1) {
		return nil
	}

	p.Wg.Add(1)
	defer p.Wg.Done()

	b, err := json.Marshal(p.JobProcVal)
	if err != nil {
		return err
	}

	// 使用 TTL 将进程信息写入 etcd，这充当了心跳机制
	// 如果节点崩溃，key 将过期并被 etcd 自动删除
	_, err = etcdclient.PutWithTtl(p.Key(), string(b), config.GetConfigModels().System.JobProcTtl)
	if err != nil {
		return err
	}
	return nil
}

// Stop 停止跟踪进程并从 etcd 中清理其记录，这是线程安全的
func (p *JobProc) Stop() {
	if p == nil {
		return
	}
	// 使用原子操作确保 Stop 只执行一次
	if !atomic.CompareAndSwapInt32(&p.Running, 1, 0) {
		return
	}

	// 等待待处理的 etcd 操作完成
	p.Wg.Wait()

	if err := p.del(); err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("proc del[%s] err: %s", p.Key(), err.Error()))
	}
}

// WatchProc 为指定节点上的所有进程变化创建一个 etcd 监视通道
func WatchProc(nodeUUID string) clientv3.WatchChan {
	keyPrefix := fmt.Sprintf(etcdclient.KeyEtcdNodeProcProfile, nodeUUID)
	return etcdclient.Watch(keyPrefix, clientv3.WithPrefix())
}
