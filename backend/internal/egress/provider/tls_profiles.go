package provider

import (
	"github.com/TokenFlux/TokenRouter/internal/egress"
	"github.com/TokenFlux/TokenRouter/internal/infra/httpclient/tlsfingerprint"
)

// TLSProfiles 将 TLS 策略服务返回的配置转换为传输指纹。
type TLSProfiles struct {
	*egress.TLSFingerprintProfileService
}

func NewTLSProfiles(core *egress.TLSFingerprintProfileService) *TLSProfiles {
	return &TLSProfiles{TLSFingerprintProfileService: core}
}

func (p *TLSProfiles) core() *egress.TLSFingerprintProfileService {
	if p == nil {
		return nil
	}
	return p.TLSFingerprintProfileService
}

// ResolveRequestTLS 根据提供商资格和路由匹配结果解析 TLS 策略，再转换为传输指纹。
func (p *TLSProfiles) ResolveRequestTLS(selection egress.TLSSelection) *tlsfingerprint.Profile {
	return ToTLSProfile(p.core().ResolveRequestPolicy(selection).TLSProfile)
}

func (p *TLSProfiles) ResolveTokenTLSProfileByID(id int64) (*tlsfingerprint.Profile, bool) {
	profile, ok := p.core().ResolveTokenTLSProfileByID(id)
	return ToTLSProfile(profile), ok
}
