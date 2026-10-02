package provider

import (
	"context"

	"github.com/TokenFlux/TokenRouter/internal/egress"
	egressadapter "github.com/TokenFlux/TokenRouter/internal/egress/provider"
)

// tlsProfileTestStore 通过相同读取入口提供固定测试策略。
type tlsProfileTestStore struct {
	egress.TLSFingerprintProfileRepository
	profiles []*egress.TLSFingerprintProfile
}

func (s *tlsProfileTestStore) List(context.Context) ([]*egress.TLSFingerprintProfile, error) {
	return s.profiles, nil
}

// newTLSProfileServiceWithCacheForTest 通过构造函数和预热填充 TLS 测试缓存。
func newTLSProfileServiceWithCacheForTest(profiles map[int64]*egress.TLSFingerprintProfile) *egressadapter.TLSProfiles {
	values := make([]*egress.TLSFingerprintProfile, 0, len(profiles))
	for _, profile := range profiles {
		values = append(values, profile)
	}
	service := egressadapter.NewTLSProfiles(egress.NewTLSFingerprintProfileService(&tlsProfileTestStore{profiles: values}, nil))
	service.Start()
	return service
}
