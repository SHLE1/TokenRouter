//go:build wireinject

package app

import (
	redisinfra "github.com/TokenFlux/TokenRouter/internal/infra/redis"
	"github.com/TokenFlux/TokenRouter/internal/provider"
	"github.com/TokenFlux/TokenRouter/internal/provider/rediscache"
	"github.com/google/wire"
)

// cacheProviders 绑定各缓存实现，使用方共用周期任务锁。
var cacheProviders = wire.NewSet(
	rediscache.NewInternal500CounterCache,
	redisinfra.NewLeaderLockCache,
	wire.Bind(new(provider.CNMonitorLeader), new(*redisinfra.LeaderLock)),
)
