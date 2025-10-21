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

// watchJobs 监视当前节点的任务更改（创建、更新、删除）
func (srv *NodeServer) watchJobs() {
	rch := handler.WatchJobs(srv.UUID)
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch {
			case ev.IsCreate():
				var job handler.Job
				if err := json.Unmarshal(ev.Kv.Value, &job); err != nil {
					logger.GetLogger().Error(fmt.Sprintf("watch job[%s] create json unmarshal err: %s", string(ev.Kv.Key), err.Error()))
					continue
				}
				srv.jobs[job.ID] = &job
				job.InitNodeInfo(models.JobStatusAssigned, srv.UUID, srv.Hostname, srv.IP)
				srv.addJob(&job)
			case ev.IsModify():
				var job handler.Job
				if err := json.Unmarshal(ev.Kv.Value, &job); err != nil {
					logger.GetLogger().Error(fmt.Sprintf("watch job[%s] modify json unmarshal err: %s", string(ev.Kv.Key), err.Error()))
					continue
				}
				job.InitNodeInfo(models.JobStatusAssigned, srv.UUID, srv.Hostname, srv.IP)
				srv.modifyJob(&job)
			case ev.Type == mvccpb.DELETE:
				srv.deleteJob(handler.GetJobIDFromKey(string(ev.Kv.Key)))
			default:
				logger.GetLogger().Warn(fmt.Sprintf("watch job unknown event type[%v] from job[%s]", ev.Type, string(ev.Kv.Key)))
			}
		}
	}
}

//fixme 杀死正在执行的任务
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

// watchSystemInfo 监视获取系统信息的请求
func (srv *NodeServer) watchSystemInfo() {
	rch := handler.WatchSystem(srv.UUID)
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch {
			case ev.IsCreate() || ev.IsModify():
				key := string(ev.Kv.Key)
				if string(ev.Kv.Value) != models.NodeSystemInfoSwitch || srv.Node.UUID != getUUID(key) {
					logger.GetLogger().Error(fmt.Sprintf("get system info from node[%s] ,switch is not alive ", srv.UUID))
					continue
				}
				s, err := utils.GetServerInfo()
				if err != nil {
					logger.GetLogger().Error(fmt.Sprintf("get system info from node[%s] error: %s", srv.UUID, err.Error()))
					continue
				}
				b, err := json.Marshal(s)
				if err != nil {
					logger.GetLogger().Error(fmt.Sprintf("get system info from node[%s] json marshal error: %s", srv.UUID, err.Error()))
					continue
				}
				// 将系统信息写入另一个具有较短 TTL 的 etcd 键，以供管理员读取
				_, err = etcdclient.PutWithTtl(fmt.Sprintf(etcdclient.KeyEtcdSystemGet, getUUID(key)), string(b), 5*60)
				if err != nil {
					logger.GetLogger().Error(fmt.Sprintf("get system info from node[%s] etcd put error: %s", srv.UUID, err.Error()))
					continue
				}

			}
		}
	}
}

// getUUID 是一个辅助函数，用于从 etcd 键路径中提取 UUID
func getUUID(key string) string {
	index := strings.LastIndex(key, "/")
	if index == -1 {
		return ""
	}
	return key[index+1:]
}

// watchOnce 监视一次性任务执行触发器
func (srv *NodeServer) watchOnce() {
	rch := handler.WatchOnce()
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch {
			case ev.IsModify(), ev.IsCreate():
				if len(ev.Kv.Value) != 0 && string(ev.Kv.Value) != srv.UUID {
					continue
				}
				j, ok := srv.jobs[handler.GetJobIDFromKey(string(ev.Kv.Key))]
				if !ok {
					continue
				}
				go j.RunWithRecovery()
			}
		}
	}
}
