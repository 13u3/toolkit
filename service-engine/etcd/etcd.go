package service

import (
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)
type EtcdService struct {
	Client *clientv3.Client
}

type WatchCallback struct{
	Type string
	Key string
	Value string
}
const(
	servicePrefix = "/service/"
)

// 创建连接
func (e *EtcdService) ConnectEtcd(host []string, dialTimeout time.Duration) error{
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   host,
		DialTimeout: time.Duration(dialTimeout) * time.Second,
	})
	if err != nil {
		return err
	}
	e.Client = cli
	return nil
}