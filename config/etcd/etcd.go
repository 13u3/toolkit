package etcd

import "time"

type DefaultConfig struct {
	Endpoints   []string
	DialTimeout time.Duration
}

// 初始化配置
func NewConfig(endpoints []string, dialTimeout time.Duration) *DefaultConfig {
	return &DefaultConfig{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	}
}
