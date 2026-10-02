package app

import (
	"github.com/TokenFlux/TokenRouter/internal/billing"
	billingredis "github.com/TokenFlux/TokenRouter/internal/billing/rediscache"
	"github.com/redis/go-redis/v9"
)

// provideWindowCostCache 使用共享 Redis 客户端和资金窗口命名空间构造缓存。
func provideWindowCostCache(rdb *redis.Client) billing.WindowCostCache {
	return billingredis.NewWindowCostCache(rdb)
}
