package redis

import "github.com/redis/go-redis/v9"

// NewClient 按给定选项创建客户端，并按启动配置添加唯一的 timing hook。
func NewClient(options *redis.Options, enableTiming bool) *redis.Client {
	client := redis.NewClient(options)
	if enableTiming {
		client.AddHook(TimingHook{})
	}
	return client
}
