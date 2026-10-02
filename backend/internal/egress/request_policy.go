package egress

// RequestPolicyInput 包含调用方已解析的出站参数。
// Header 在构造请求时应用，TLS 和重定向参数在获取客户端前确定。
type RequestPolicyInput struct {
	ProxyURL                                              string
	TLSProfile                                            *TLSFingerprintProfile
	Headers                                               map[string]string
	ValidateResolvedIP, PublicHostsOnly, DisableRedirects bool
}

// RequestPolicy 返回独立策略副本；允许的 Header 始终使用同一安全规则。
func RequestPolicy(input RequestPolicyInput) EgressPolicy {
	return EgressPolicy{ProxyURL: input.ProxyURL, TLSProfile: CloneTLSFingerprintProfile(input.TLSProfile), Headers: ResolveHeaderOverrides(input.Headers), ValidateResolvedIP: input.ValidateResolvedIP, PublicHostsOnly: input.PublicHostsOnly, DisableRedirects: input.DisableRedirects}
}

// RequiresHostValidation 在全局 IP 校验或请求级公网限制启用时返回 true。
func (p EgressPolicy) RequiresHostValidation() bool { return p.ValidateResolvedIP || p.PublicHostsOnly }

// String 与 GoString 不把代理认证或 Header 值写入普通诊断日志。
func (p EgressPolicy) String() string   { return "egress policy (" + p.TransportMode + ")" }
func (p EgressPolicy) GoString() string { return p.String() }
