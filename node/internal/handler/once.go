package handler

import (
	"pulse/common/pkg/etcdclient"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// WatchOnce creates an etcd watch channel specifically for one-time jobs.
func WatchOnce() clientv3.WatchChan {
	// It watches the predefined key prefix for one-time jobs.
	return etcdclient.Watch(etcdclient.KeyEtcdOnceProfile, clientv3.WithPrefix())
}
