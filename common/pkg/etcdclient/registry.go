package etcdclient

import (
	"context"
	"pulse/common/pkg/logger"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// ServerReg 结构体用于管理一个服务在 etcd 上的注册信息和生命周期
type ServerReg struct {
	Client        *Client                                 // 封装了 etcd 客户端连接的自定义结构体
	stop          chan error                              // 用于goroutine停止工作的通道
	leaseId       clientv3.LeaseID                        // etcd 生成的租约ID
	cancelFunc    func()                                  // 与 context 关联的取消函数
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse // 接收 etcd 服务端对租约续约请求响应的只读通道
	// time-to-live
	Ttl int64 // 租约的有效期
}

// NewServerReg 是 ServerReg 的构造函数
func NewServerReg(ttl int64) *ServerReg {
	return &ServerReg{
		// 单例模式, 避免重复创建 etcd 连接
		Client: _defaultEtcd,
		Ttl:    ttl,              // 设置租约的 TTL
		stop:   make(chan error), // 初始化用于停止信号的通道
	}
}

func (s *ServerReg) Register(key string, value string) error {
	if err := s.setLease(s.Ttl); err != nil {
		return err
	}
	go s.keepAlive()
	if err := s.putService(key, value); err != nil {
		return err
	}
	return nil
}

func (s *ServerReg) setLease(ttl int64) error {
	leaseResp, err := Grant(ttl)
	if err != nil {
		return err
	}

	ctx, cancelFunc := context.WithCancel(context.TODO())
	leaseRespChan, err := s.Client.KeepAlive(ctx, leaseResp.ID)

	if err != nil {
		cancelFunc()
		return err
	}
	s.leaseId = leaseResp.ID
	s.cancelFunc = cancelFunc
	s.keepAliveChan = leaseRespChan
	return nil
}
func (s *ServerReg) Stop() {
	s.stop <- nil
}

// Monitor the lease renewal
func (s *ServerReg) keepAlive() {
	for {
		select {
		case <-s.stop:
			return
		case leaseKeepResp := <-s.keepAliveChan:
			if leaseKeepResp == nil {
				logger.GetLogger().Info("the lease renewal function has been turned off\n")
				return
			}
		}
	}
}

func (s *ServerReg) putService(key, val string) error {
	kv := clientv3.NewKV(s.Client.Client)
	_, err := kv.Put(context.TODO(), key, val, clientv3.WithLease(s.leaseId))
	return err
}

func (s *ServerReg) RevokeLease() error {
	s.cancelFunc()
	time.Sleep(2 * time.Second)
	_, err := Revoke(s.leaseId)
	return err
}
