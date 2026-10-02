//go:build wireinject

package app

import (
	anthropicredis "github.com/TokenFlux/TokenRouter/internal/upstream/anthropic/rediscache"

	gatewaytransport "github.com/TokenFlux/TokenRouter/internal/gateway/provider/transport"
	httpclient "github.com/TokenFlux/TokenRouter/internal/infra/httpclient"
	"github.com/google/wire"
)

// upstreamAssemblyProviders 汇总上游客户端的 Wire provider。
var upstreamAssemblyProviders = wire.NewSet(
	wire.Bind(new(httpclient.UpstreamTransport), new(*gatewaytransport.Client)),
	upstreamClientProviders,
	providePrivacyClientFactory,
	anthropicredis.NewFingerprintStore,
)
