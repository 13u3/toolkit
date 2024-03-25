package etcd

import (
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type DefaultConfig struct {
	Endpoints   []string
	DialTimeout time.Duration
	ContextTimeout time.Duration
}

// 初始化默认配置
func NewDefaultConfig(endpoints []string, dialTimeout time.Duration, ContextTimeout time.Duration) *DefaultConfig {
	return &DefaultConfig{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
		ContextTimeout: ContextTimeout,
	}
}

// 创建etcd客户端
func (dc *DefaultConfig) NewClient() (*clientv3.Client, error) {
    return clientv3.New(clientv3.Config{
		Endpoints:   dc.Endpoints,
		DialTimeout: time.Duration(dc.DialTimeout) * time.Second,
	})
}

// 注册服务
func (dc *DefaultConfig) RegisterService(client *clientv3.Client, key, value string) error {
    ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dc.ContextTimeout) * time.Second)
	resp, err := client.Put(ctx, key, value)
	defer cancel()
	if err != nil {
	    fmt.Printf("register service failed, err:%v\n", err)
	}
	fmt.Printf("register service success, resp:%v\n", resp)
	return err
}