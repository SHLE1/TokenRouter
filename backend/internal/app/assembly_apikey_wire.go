//go:build wireinject

package app

import (
	apikey "github.com/TokenFlux/TokenRouter/internal/apikey"

	keypostgres "github.com/TokenFlux/TokenRouter/internal/apikey/postgres"

	keyredis "github.com/TokenFlux/TokenRouter/internal/apikey/rediscache"

	"github.com/google/wire"
)

// apikeyAssemblyProviders 汇总 apikey 模块的 Wire provider。
var apikeyAssemblyProviders = wire.NewSet(
	provideKeyAdminHTTP,
	provideKeyHTTP,
	provideKeyAdmin,
	provideKeyStore,
	provideKeyRepository,
	provideKeys,
	provideKeyInvalidator,
	apikey.ProvideAuthCacheInvalidationWorker,
	keyredis.NewAPIKeyCache,
	keypostgres.NewAuthCacheInvalidationOutboxRepository,
	provideAPIKeyAuth,
)
