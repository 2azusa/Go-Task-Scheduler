package handler

import (
	"fmt"
	"pulse/common/pkg/etcdclient"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// WatchSystem creates an etcd watch channel for system-level events or switches on a specific node.
func WatchSystem(nodeUUID string) clientv3.WatchChan {
	// The key is specific to the node and is used to trigger system-level actions, like fetching system info.
	key := fmt.Sprintf(etcdclient.KeyEtcdSystemSwitch, nodeUUID)
	return etcdclient.Watch(key, clientv3.WithPrefix())
}
