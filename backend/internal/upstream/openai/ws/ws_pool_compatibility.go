package ws

// wsConnCompatibility 保存建立连接时的目标、代理和握手参数。
type wsConnCompatibility struct {
	wsURL         string
	proxyURL      string
	tlsProfileKey string
	handshake     openAIWSHandshakeCompatibilityKey
}

// wsCompatibilityForRequest 为拨号、复用和预热生成相同的连接标识。
func wsCompatibilityForRequest(req WSAcquireRequest) wsConnCompatibility {
	return wsConnCompatibility{
		wsURL:         stringsTrim(req.WSURL),
		proxyURL:      stringsTrim(req.ProxyURL),
		tlsProfileKey: openAIWSTLSProfileKey(req.TLSProfile, req.TLSProfileKey),
		handshake:     normalizeOpenAIWSHandshakeCompatibility(req.Provider, req.Headers),
	}
}

// matchesCompatibility 判断连接是否使用本次请求的出站配置。
func (c *WSConn) matchesCompatibility(compatibility wsConnCompatibility) bool {
	return c != nil && c.compatibility == compatibility
}
