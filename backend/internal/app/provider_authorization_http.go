package app

import (
	providerauth "github.com/TokenFlux/TokenRouter/internal/provider"
	provideradapter "github.com/TokenFlux/TokenRouter/internal/provider/provider"

	providerhttp "github.com/TokenFlux/TokenRouter/internal/provider/httpapi"
)

// provideClaudeAuthorizationHTTP 为 Claude 授权 HTTP 入口绑定授权用例。
func provideClaudeAuthorizationHTTP(source *providerauth.ClaudeAuthorization) *providerhttp.ClaudeOAuthHandler {
	return providerhttp.NewClaudeOAuthHandler(source)
}

func provideQoderAuthorizationHTTP(source *provideradapter.QoderAuthorization) *providerhttp.QoderOAuthHandler {
	return providerhttp.NewQoderOAuthHandler(source)
}
