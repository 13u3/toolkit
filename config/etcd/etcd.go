package etcd

import "time"
type Config struct {
    Endpoints []string
	DialTimeout time.Duration
}

func NewConfig(endpoints []string, dialTimeout time.Duration) *Config {
    return &Config{
        Endpoints: endpoints,
        DialTimeout: dialTimeout,
    }
}

func (c *Config) WithDefaults() *Config {
    if c.DialTimeout == 0 {
        c.DialTimeout = 5 * time.Second // 默认5秒
    }
    return c
}

