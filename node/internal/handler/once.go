package handler

import (
	"pulse/common/pkg/etcdclient"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// WatchOnce 为一次性任务创建一个etcd监视通道
func WatchOnce() clientv3.WatchChan {
	// 监视一次性任务的预定义键前缀
	return etcdclient.Watch(etcdclient.KeyEtcdOnceProfile, clientv3.WithPrefix())
}
