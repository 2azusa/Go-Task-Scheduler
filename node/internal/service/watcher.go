package service

import (
	"encoding/json"
	"fmt"
	"pulse/common/models"
	"pulse/common/pkg/etcdclient"
	"pulse/common/pkg/logger"
	"pulse/common/pkg/utils"
	"pulse/node/internal/handler"
	"strings"

	"go.etcd.io/etcd/api/v3/mvccpb"
)

// watchJobs 监听分配给当前节点的任务的变化，新增、修改、删除
func (srv *NodeServer) watchJobs() {
	// 获取一个针对本节点任务的 Etcd watch channel
	rch := handler.WatchJobs(srv.UUID)
	// 持续从 channel 中接收事件
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch {
			case ev.IsCreate(): // 如果是创建事件
				var job handler.Job
				// 反序列化任务数据
				if err := json.Unmarshal(ev.Kv.Value, &job); err != nil {
					logger.GetLogger().Error(fmt.Sprintf("watch job[%s] create json unmarshal err: %s", string(ev.Kv.Key), err.Error()))
					continue
				}
				srv.jobs[job.ID] = &job
				job.InitNodeInfo(models.JobStatusAssigned, srv.UUID, srv.Hostname, srv.IP)
				srv.addJob(&job) // 将任务添加到调度器
			case ev.IsModify(): // 如果是修改事件
				var job handler.Job
				if err := json.Unmarshal(ev.Kv.Value, &job); err != nil {
					logger.GetLogger().Error(fmt.Sprintf("watch job[%s] modify json unmarshal err: %s", string(ev.Kv.Key), err.Error()))
					continue
				}
				job.InitNodeInfo(models.JobStatusAssigned, srv.UUID, srv.Hostname, srv.IP)
				srv.modifyJob(&job) // 修改调度器中的任务
			case ev.Type == mvccpb.DELETE: // 如果是删除事件
				srv.deleteJob(handler.GetJobIDFromKey(string(ev.Kv.Key))) //删除任务
			default:
				logger.GetLogger().Warn(fmt.Sprintf("watch job unknown event type[%v] from job[%s]", ev.Type, string(ev.Kv.Key)))
			}
		}
	}
}

//fixme kill executing job
/*
func (srv *NodeServer) watchKilledProc() {
	rch := handler.WatchProc(srv.UUID)
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch {
			case ev.IsModify():
				proc, err := handler.GetProcFromKey(string(ev.Kv.Key))
				if err != nil {
					logger.GetLogger().Error(fmt.Sprintf("watch killed proc error:%s kv:%s", err.Error(), ev.Kv.String()))
					continue
				}
				procVal := &models.JobProcVal{}
				err = json.Unmarshal(ev.Kv.Value, procVal)
				if err != nil {
					logger.GetLogger().Warn(fmt.Sprintf("watch killed proc json warn:%s kv:%s", err.Error(), ev.Kv.String()))
					continue
				}
				proc.JobProcVal = *procVal
				if proc.Killed {
					if err := syscall.Kill(proc.ID, syscall.SIGKILL); err != nil {
						logger.GetLogger().Error(fmt.Sprintf("process:[%d] force kill failed, error:[%s]", proc.ID, err))
						return
					}
				}

			}
		}
	}
}*/

// watchSystemInfo 监听获取系统信息的请求
func (srv *NodeServer) watchSystemInfo() {
	// 获取一个针对本节点系统信息请求的 Etcd watch channel
	rch := handler.WatchSystem(srv.UUID)
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch {
			case ev.IsCreate() || ev.IsModify():
				key := string(ev.Kv.Key)
				// 校验指令是否有效
				if string(ev.Kv.Value) != models.NodeSystemInfoSwitch || srv.Node.UUID != getUUID(key) {
					logger.GetLogger().Error(fmt.Sprintf("get system info from node[%s] ,switch is not alive ", srv.UUID))
					continue
				}
				// 获取本机的服务器信息 (CPU, 内存, 磁盘等)
				s, err := utils.GetServerInfo()
				if err != nil {
					logger.GetLogger().Error(fmt.Sprintf("get system info from node[%s] error: %s", srv.UUID, err.Error()))
					continue
				}
				// 序列化为 JSON
				b, err := json.Marshal(s)
				if err != nil {
					logger.GetLogger().Error(fmt.Sprintf("get system info from node[%s] json marshal error: %s", srv.UUID, err.Error()))
					continue
				}
				// 将获取到的信息写入 Etcd 的另一个 key，并设置一个较短的 TTL，供 Admin 读取
				_, err = etcdclient.PutWithTtl(fmt.Sprintf(etcdclient.KeyEtcdSystemGet, getUUID(key)), string(b), 5*60)
				if err != nil {
					logger.GetLogger().Error(fmt.Sprintf("get system info from node[%s] etcd put error: %s", srv.UUID, err.Error()))
					continue
				}

			}
		}
	}
}

// getUUID 是一个从 Etcd key 路径中提取 UUID 的辅助函数
func getUUID(key string) string {
	// /pulse/node/<node_uuid>
	index := strings.LastIndex(key, "/")
	if index == -1 {
		return ""
	}
	return key[index+1:]
}

// watchOnce 监听一次性任务的执行触发
func (srv *NodeServer) watchOnce() {
	// 获取一个针对一次性任务的 Etcd watch channel
	rch := handler.WatchOnce()
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch {
			case ev.IsModify(), ev.IsCreate():
				// 检查该指令是否是发给本节点的
				if len(ev.Kv.Value) != 0 && string(ev.Kv.Value) != srv.UUID {
					continue
				}
				// 从本地任务 map 中查找对应的任务
				j, ok := srv.jobs[handler.GetJobIDFromKey(string(ev.Kv.Key))]
				if !ok {
					continue
				}
				// 在一个新的 goroutine 中立即执行该任务
				go j.RunWithRecovery()
			}
		}
	}
}
