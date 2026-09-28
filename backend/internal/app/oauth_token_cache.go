package app

import (
	"github.com/TokenFlux/TokenRouter/internal/provider"
	"github.com/TokenFlux/TokenRouter/internal/provider/rediscache"
	"github.com/redis/go-redis/v9"
)

func provideOAuthTokenCache(rdb *redis.Client) provider.AccessTokenCache {
	return rediscache.NewOAuthTokenCache(rdb)
}
