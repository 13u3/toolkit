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
}

type Service struct {
	Client *clientv3.Client
}

const (
	ServicePrefix = "/services/"
)

// 初始化默认配置
func NewDefaultConfig(endpoints []string, dialTimeout time.Duration) *DefaultConfig {
	return &DefaultConfig{
		Endpoints:   endpoints,
		DialTimeout: 10 * time.Second,
	}
}

// 创建etcd客户端
func (c *DefaultConfig) NewClient() (*Service, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   c.Endpoints,
		DialTimeout: time.Duration(c.DialTimeout) * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return &Service{
		Client: cli,
	}, nil
}

// 注册服务
func (s *Service) RegisterService(uuid, nodeId, Address string) error {
	kv := clientv3.NewKV(s.Client)
	ctx := context.Background()
	// 创建租约
	lease := clientv3.NewLease(s.Client)
	leaseResp, err := lease.Grant(ctx, 30) //60秒
	if err != nil {
		return err
	}
	// 注册
	_, err = kv.Put(ctx, ServicePrefix+uuid+"/"+nodeId, Address, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}

	// 续约，keepRespChan是个只读的Channel
	keepRespChan, err := lease.KeepAlive(ctx, leaseResp.ID)
	if err != nil {
		return err
	}
	go PrintKeepRespChan(keepRespChan)
	return nil
}

// 打印续约信息
func PrintKeepRespChan(keepRespChan <-chan *clientv3.LeaseKeepAliveResponse) {
	for {
		select {
		case keepResp := <-keepRespChan:
			if keepResp == nil {
				return
			}
			// 续约成功
			fmt.Println("续约成功")
		}
	}
}

// 注销服务
func (s *Service) UnRegisterService(uuid, nodeId string) error {
	kv := clientv3.NewKV(s.Client)
	ctx := context.Background()
	_, err := kv.Delete(ctx, ServicePrefix+uuid+"/"+nodeId)
	if err != nil {
		return err
	}
	return nil
}

// 获取服务列表
func (s *Service) GetServiceList(uuid string) ([]string, error) {
	kv := clientv3.NewKV(s.Client)
	ctx := context.Background()
	resp, err := kv.Get(ctx, ServicePrefix+uuid, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	var serviceList []string
	for _, kvpair := range resp.Kvs {
		serviceList = append(serviceList, string(kvpair.Value))
	}
	return serviceList, nil
}
