package etcd

import "time"

type DefaultConfig struct {
	Endpoints   []string
	DialTimeout time.Duration
}

// 初始化配置
func NewDefaultConfig(endpoints []string, dialTimeout time.Duration) *DefaultConfig {
	return &DefaultConfig{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	}
}

// 连接Etcd