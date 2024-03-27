package etcd

import (
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

/* type DefaultConfig struct {
	Endpoints   []string
	DialTimeout time.Duration
} */

type Service struct {
	Client *clientv3.Client
}

const (
	ServicePrefix = "/service/"
)

// 初始化默认配置
/* func NewDefaultConfig(endpoints []string, dialTimeout time.Duration) *DefaultConfig {
	return &DefaultConfig{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	}
} */

// 创建etcd客户端
/* func (c *DefaultConfig) NewClient() (*Service, error) {
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
} */
func NewClient(endpoints []string, dialTimeout time.Duration) (*Service, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: time.Duration(dialTimeout) * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return &Service{
		Client: cli,
	}, nil
}

// 注册服务
func (s *Service) RegisterService(clusterId, nodeId, Address string) error {
	kv := clientv3.NewKV(s.Client)
	//ctx := context.Background()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	// 创建租约
	lease := clientv3.NewLease(s.Client)
	leaseResp, err := lease.Grant(ctx, 30) //60秒
	if err != nil {
		return err
	}
	// 注册
	_, err = kv.Put(ctx, ServicePrefix+clusterId+"/"+nodeId, Address, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}
	// 续约，keepRespChan是个只读的Channel
	keepRespChan, err := lease.KeepAlive(context.Background(), leaseResp.ID)
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
func (s *Service) UnRegisterService(clusterId, nodeId string) error {
	kv := clientv3.NewKV(s.Client)
	//ctx := context.Background()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 10)
	defer cancel()
	_, err := kv.Delete(ctx, ServicePrefix+clusterId+"/"+nodeId)
	if err != nil {
		return err
	}
	defer s.Client.Close()
	return nil
}

// 获取服务列表
func (s *Service) GetServiceList(clusterId string) ([]string, error) {
	kv := clientv3.NewKV(s.Client)
	//ctx := context.Background()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 10)
	defer cancel()
	resp, err := kv.Get(ctx, ServicePrefix+clusterId, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	var serviceList []string
	for _, kvpair := range resp.Kvs {
		serviceList = append(serviceList, string(kvpair.Value))
	}
	defer s.Client.Close()
	return serviceList, nil
}

// 获取服务
func (s *Service) GetService(clusterId, nodeId string) (string, error) {
    kv := clientv3.NewKV(s.Client)
	//ctx := context.Background()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 10)
	defer cancel()
	resp, err := kv.Get(ctx, ServicePrefix+clusterId+"/"+nodeId)
	if err != nil {
		return "", err
	}
	defer s.Client.Close()
	return string(resp.Kvs[0].Value), nil
}

// 监听服务列表
func (s *Service) WatchService(key string, callback func(string), withPrefix ...bool) {
	//ctx := context.Background()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 10)
	defer cancel()
	var rch clientv3.WatchChan
	if len(withPrefix) > 0 && withPrefix[0] {
		rch = s.Client.Watch(ctx, ServicePrefix+key, clientv3.WithPrefix())
	} else {
		rch = s.Client.Watch(ctx, ServicePrefix+key)
	}
	go func() {
	    for {
	        select {
	        case wresp := <-rch:
	            for _, ev := range wresp.Events {
	                switch ev.Type {
	                case clientv3.EventTypePut:
	                    callback(string(ev.Kv.Key) + "-" + string(ev.Kv.Value))
	                case clientv3.EventTypeDelete:
	                    callback(string(ev.Kv.Key) + "-" + string(ev.Kv.Value))
	                }
	            }
	        }
	    }
	}()
}