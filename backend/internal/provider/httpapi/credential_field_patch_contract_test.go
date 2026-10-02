package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TokenFlux/TokenRouter/internal/routing/capability"

	"github.com/TokenFlux/TokenRouter/internal/provider"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// 在预读后轮换 token 并修改普通配置，检查批量字段更新保持这些并发修改。
type credentialFieldRaceAdmin struct {
	ProviderManagement
	current provider.Record
}

func (s *credentialFieldRaceAdmin) GetProvider(context.Context, int64) (*provider.Record, error) {
	v := s.current
	v.Credentials = make(map[string]any, len(s.current.Credentials))
	for k, x := range s.current.Credentials {
		v.Credentials[k] = x
	}
	return &v, nil
}

func (s *credentialFieldRaceAdmin) UpdateProvider(_ context.Context, _ int64, input *provider.UpdateProviderInput) (*provider.Record, error) {
	s.current.Credentials = map[string]any{"refresh_token": "rotated", "base_url": "https://new.invalid", "org_uuid": "old-org"}
	for key, value := range input.Credentials {
		s.current.Credentials[key] = value
	}
	return &s.current, nil
}

func TestBatchCredentialFieldDoesNotRestoreOldSnapshot(t *testing.T) {
	admin := &credentialFieldRaceAdmin{current: provider.Record{ID: 984, Platform: capability.PlatformAnthropic, Type: capability.ProviderTypeOAuth, Credentials: map[string]any{"refresh_token": "old-token", "base_url": "https://old.invalid", "org_uuid": "old-org"}}}
	h := NewManagementHandler(admin, ManagementOptions{Batch: provider.NewManagementBatch(admin, nil)})
	router := gin.New()
	router.POST("/batch", h.BatchUpdateCredentials)
	request := httptest.NewRequest(http.MethodPost, "/batch", bytes.NewBufferString(`{"provider_ids":[984],"field":"org_uuid","value":"new-org"}`))
	request.Header.Set("Content-Type", "application/json")
	result := httptest.NewRecorder()
	router.ServeHTTP(result, request)
	require.Equal(t, http.StatusOK, result.Code)
	require.Equal(t, "rotated", admin.current.Credentials["refresh_token"])
	require.Equal(t, "https://new.invalid", admin.current.Credentials["base_url"])
	require.Equal(t, "new-org", admin.current.Credentials["org_uuid"])
}
