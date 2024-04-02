package etcd

import (
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type WatchCallback struct{
	Type string //事件类型 PUT/DELETE
	Key string //服务key
	Value string //服务value
}

// 注册服务
func (e *EtcdService) RegisterService() error {
	if e.Client == nil {
		return fmt.Errorf("连接Etcd未初始化")
	}
	kv := clientv3.NewKV(e.Client)
	//ctx := context.Background()
	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()
	// 创建租约
	lease := clientv3.NewLease(e.Client)
	leaseResp, err := lease.Grant(ctx, 30) //租约时间，秒
	if err != nil {
		return err
	}
	// 注册自己的服务，并绑定租约
	_, err = kv.Put(ctx, servicePrefix + e.RegContent.Key, e.RegContent.Value, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}
	// 续约，keepRespChan是个只读的Channel
	keepRespChan, err := lease.KeepAlive(context.Background(), leaseResp.ID)
	if err != nil {
		return err
	}
	go PrintEtcdKeepRespChan(keepRespChan)
	return nil
}

// 打印续约信息
func PrintEtcdKeepRespChan(keepRespChan <-chan *clientv3.LeaseKeepAliveResponse) {
	for {
		select {
		case keepResp := <-keepRespChan:
			if keepResp == nil {
				return
			}
			// 续约成功
			fmt.Printf("Etcd服务续约成功：%+v\n", keepResp)
		}
	}
}

// 获取指定服务列表
func (e *EtcdService) GetServiceList() ([]string, error) {
	if e.Client == nil {
		return nil, fmt.Errorf("Etcd连接未初始化")
	}
	kv := clientv3.NewKV(e.Client)
	//ctx := context.Background()
	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()
	resp, err := kv.Get(ctx, servicePrefix + e.RegContent.Key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	var serviceList []string
	for _, kvpair := range resp.Kvs {
		serviceList = append(serviceList, string(kvpair.Value))
	}
	return serviceList, nil
}

// 监听指定服务列表
func (e *EtcdService) WatchService(callback func(WatchCallback)) error{
    if e.Client == nil {
		return fmt.Errorf("Etcd连接未初始化")
	}
	watchRespCh := e.Client.Watch(context.Background(), servicePrefix + e.RegContent.Key, clientv3.WithPrefix())
	go func() {
		fmt.Println("开始监听"+e.RegContent.Key+"服务......")
	    for {
	        select {
	        case watchResp := <-watchRespCh:
	            for _, event := range watchResp.Events {
	                switch event.Type {
	                case clientv3.EventTypePut:
	                    go callback(WatchCallback{
							Type: "PUT",
							Key: string(event.Kv.Key),
							Value: string(event.Kv.Value),
						})
	                case clientv3.EventTypeDelete:
	                    go callback(WatchCallback{
							Type: "DELETE",
							Key: string(event.Kv.Key),
						})
	                }
	            }
	        }
	    }
	}()
	return nil
}