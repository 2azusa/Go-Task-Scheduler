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

// JobProc represents information about a running job process.
type JobProc struct {
	*models.JobProc
}

// GetProcFromKey parses JobProc information from an etcd key.
// The expected key format is: /pulse/proc/{nodeUUID}/{jobId}/{procId}
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

// Key generates a unique etcd key for the JobProc instance.
func (p *JobProc) Key() string {
	return fmt.Sprintf(etcdclient.KeyEtcdProc, p.NodeUUID, p.JobID, p.ID)
}

// del deletes the process key from etcd.
func (p *JobProc) del() error {
	_, err := etcdclient.Delete(p.Key())
	return err
}

// Start registers a running process in etcd. It is thread-safe.
func (p *JobProc) Start() error {
	// Use atomic operation to ensure Start is executed only once.
	if !atomic.CompareAndSwapInt32(&p.Running, 0, 1) {
		return nil
	}

	p.Wg.Add(1)
	defer p.Wg.Done()

	b, err := json.Marshal(p.JobProcVal)
	if err != nil {
		return err
	}

	// Write process info to etcd with a TTL. This acts as a heartbeat mechanism.
	// If the node crashes, the key will expire and be automatically deleted by etcd.
	_, err = etcdclient.PutWithTtl(p.Key(), string(b), config.GetConfigModels().System.JobProcTtl)
	if err != nil {
		return err
	}
	return nil
}

// Stop stops tracking the process and cleans up its record from etcd. It is thread-safe.
func (p *JobProc) Stop() {
	if p == nil {
		return
	}
	// Use atomic operation to ensure Stop is executed only once.
	if !atomic.CompareAndSwapInt32(&p.Running, 1, 0) {
		return
	}

	// Wait for any pending etcd operations to complete.
	p.Wg.Wait()

	if err := p.del(); err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("proc del[%s] err: %s", p.Key(), err.Error()))
	}
}

// WatchProc creates an etcd watch channel for all process changes on a specific node.
func WatchProc(nodeUUID string) clientv3.WatchChan {
	keyPrefix := fmt.Sprintf(etcdclient.KeyEtcdNodeProcProfile, nodeUUID)
	return etcdclient.Watch(keyPrefix, clientv3.WithPrefix())
}
