package handler

import (
	"fmt"
	"pulse/common/pkg/etcdclient"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// WatchSystem 为指定节点上的系统级事件或开关创建 etcd 监视通道
func WatchSystem(nodeUUID string) clientv3.WatchChan {
	// 用于出发系统级操作，如获取系统信息
	key := fmt.Sprintf(etcdclient.KeyEtcdSystemSwitch, nodeUUID)
	return etcdclient.Watch(key, clientv3.WithPrefix())
}
