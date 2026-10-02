package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TokenFlux/TokenRouter/internal/identity/httpapi/authctx"

	"github.com/TokenFlux/TokenRouter/internal/audit"
	"github.com/TokenFlux/TokenRouter/internal/config"
	"github.com/TokenFlux/TokenRouter/internal/identity"
	"github.com/TokenFlux/TokenRouter/internal/server/runtimeconfig"
	settingshttp "github.com/TokenFlux/TokenRouter/internal/settings/httpapi"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// step-up 开关转换的门控测试。
// 测试环境不注入认证上下文/userService，因此一旦触发校验会以 401/403/500 中止；
// 借此区分“触发了转换校验”与“直接放行到常规保存（200）”。

func newStepUpSwitchTestHandler(t *testing.T, stored map[string]string) (*settingshttp.Handler, *settingHandlerRepoStub) {
	t.Helper()

	repo := &settingHandlerRepoStub{values: stored}
	options, _ := newCompositeSettingsHTTPFixture(repo, &config.Config{Default: config.DefaultConfig{UserConcurrency: 5}})
	return settingshttp.NewHandler(options), repo
}

func doUpdateSettings(t *testing.T, h *settingshttp.Handler, body map[string]any, prepare func(c *gin.Context)) *httptest.ResponseRecorder {
	t.Helper()
	rawBody, err := json.Marshal(body)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", bytes.NewReader(rawBody))
	c.Request.Header.Set("Content-Type", "application/json")
	if prepare != nil {
		prepare(c)
	}

	h.UpdateSettings(c)
	return rec
}

// TestUpdateSettingsEnableStepUpRejectsWithoutSession 检查 step-up 从关闭切为开启时，无认证上下文的请求被拒绝并返回专用错误标记。
func TestUpdateSettingsEnableStepUpRejectsWithoutSession(t *testing.T) {
	h, repo := newStepUpSwitchTestHandler(t, map[string]string{})

	rec := doUpdateSettings(t, h, map[string]any{"step_up_enabled": true}, nil)

	require.Equal(t, http.StatusForbidden, rec.Code)
	require.Contains(t, rec.Body.String(), "STEP_UP_ENABLE_REQUIRES_TOTP")
	require.NotEqual(t, "true", repo.values[identity.SettingKeyStepUpEnabled])
}

// TestUpdateSettingsEnableStepUpRejectsAdminAPIKey 检查 admin API key 开启 step-up 时被拒绝，reason 供前端区分错误。
func TestUpdateSettingsEnableStepUpRejectsAdminAPIKey(t *testing.T) {
	h, _ := newStepUpSwitchTestHandler(t, map[string]string{})

	rec := doUpdateSettings(t, h, map[string]any{"step_up_enabled": true}, func(c *gin.Context) {
		c.Set("auth_method", audit.AuditAuthMethodAdminAPIKey)
	})

	require.Equal(t, http.StatusForbidden, rec.Code)
	require.Contains(t, rec.Body.String(), "STEP_UP_ADMIN_API_KEY_FORBIDDEN")
}

// TestUpdateSettingsEnableStepUpFailsClosedWithoutUserService 检查开启 step-up 时，有会话但缺少 userService 的请求返回 500。
func TestUpdateSettingsEnableStepUpFailsClosedWithoutUserService(t *testing.T) {
	h, repo := newStepUpSwitchTestHandler(t, map[string]string{})

	rec := doUpdateSettings(t, h, map[string]any{"step_up_enabled": true}, func(c *gin.Context) {
		c.Set(string(authctx.ContextKeyUser), authctx.AuthSubject{UserID: 1})
	})

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.NotEqual(t, "true", repo.values[identity.SettingKeyStepUpEnabled])
}

// TestUpdateSettingsDisableStepUpRequiresStepUp 检查关闭 step-up 时，无认证上下文的请求返回 401。
func TestUpdateSettingsDisableStepUpRequiresStepUp(t *testing.T) {
	h, repo := newStepUpSwitchTestHandler(t, map[string]string{
		identity.SettingKeyStepUpEnabled: "true",
	})

	rec := doUpdateSettings(t, h, map[string]any{"step_up_enabled": false}, nil)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.Equal(t, "true", repo.values[identity.SettingKeyStepUpEnabled])
}

// TestUpdateSettingsDisableStepUpRejectsAdminAPIKey 检查 admin API key 关闭 step-up 时返回 403。
func TestUpdateSettingsDisableStepUpRejectsAdminAPIKey(t *testing.T) {
	h, repo := newStepUpSwitchTestHandler(t, map[string]string{
		identity.SettingKeyStepUpEnabled: "true",
	})

	rec := doUpdateSettings(t, h, map[string]any{"step_up_enabled": false}, func(c *gin.Context) {
		c.Set("auth_method", audit.AuditAuthMethodAdminAPIKey)
	})

	require.Equal(t, http.StatusForbidden, rec.Code)
	require.Contains(t, rec.Body.String(), "STEP_UP_ADMIN_API_KEY_FORBIDDEN")
	require.Equal(t, "true", repo.values[identity.SettingKeyStepUpEnabled])
}

// TestUpdateSettingsStepUpNoTransitionSkipsGate 检查 step-up 保持关闭时常规保存成功，并持久化 false。
func TestUpdateSettingsStepUpNoTransitionSkipsGate(t *testing.T) {
	h, repo := newStepUpSwitchTestHandler(t, map[string]string{})

	rec := doUpdateSettings(t, h, map[string]any{"step_up_enabled": false}, nil)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "false", repo.values[identity.SettingKeyStepUpEnabled])
	// 会话 IP/UA 绑定默认关闭，未提交该字段时持久化 false。
	require.Equal(t, "false", repo.values[identity.SettingKeySessionBindingEnabled])
}

// TestUpdateSettingsStepUpKeepEnabledSkipsGate 检查 step-up 保持开启时直接执行常规保存。
func TestUpdateSettingsStepUpKeepEnabledSkipsGate(t *testing.T) {
	h, repo := newStepUpSwitchTestHandler(t, map[string]string{
		identity.SettingKeyStepUpEnabled: "true",
	})

	rec := doUpdateSettings(t, h, map[string]any{"step_up_enabled": true}, nil)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "true", repo.values[identity.SettingKeyStepUpEnabled])
}

// TestUpdateSettingsOmittedSecuritySwitchesKeepStoredValues 检查请求省略 step_up_enabled 和 session_binding_enabled 时保持存储值。
// 开关保持开启时按常规流程保存。
func TestUpdateSettingsOmittedSecuritySwitchesKeepStoredValues(t *testing.T) {
	h, repo := newStepUpSwitchTestHandler(t, map[string]string{
		identity.SettingKeyStepUpEnabled:                       "true",
		identity.SettingKeySessionBindingEnabled:               "true",
		identity.SettingKeyRegistrationEmailDomainQuotaEnabled: "true",
		identity.SettingKeyUserEmailChangeEnabled:              "true",
	})

	rec := doUpdateSettings(t, h, map[string]any{"registration_enabled": true}, nil)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "true", repo.values[identity.SettingKeyStepUpEnabled])
	require.Equal(t, "true", repo.values[identity.SettingKeySessionBindingEnabled])
	require.Equal(t, "true", repo.values[identity.SettingKeyRegistrationEmailDomainQuotaEnabled])
	require.Equal(t, "true", repo.values[identity.SettingKeyUserEmailChangeEnabled])
}

// TestUpdateSettingsOmittedSecuritySwitchesKeepDisabled 验证省略字段在开关本就关闭时同样保持关闭（默认值路径）。
func TestUpdateSettingsOmittedSecuritySwitchesKeepDisabled(t *testing.T) {
	h, repo := newStepUpSwitchTestHandler(t, map[string]string{})

	rec := doUpdateSettings(t, h, map[string]any{"registration_enabled": true}, nil)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "false", repo.values[identity.SettingKeyStepUpEnabled])
	require.Equal(t, "false", repo.values[identity.SettingKeySessionBindingEnabled])
	require.Equal(t, "false", repo.values[identity.SettingKeyRegistrationEmailDomainQuotaEnabled])
	require.Equal(t, "false", repo.values[identity.SettingKeyUserEmailChangeEnabled])
}

// TestUpdateSettingsEnablesUserEmailChange 检查邮箱换绑开关在合并部分更新请求后持久化为开启。
func TestUpdateSettingsEnablesUserEmailChange(t *testing.T) {
	h, repo := newStepUpSwitchTestHandler(t, map[string]string{})

	rec := doUpdateSettings(t, h, map[string]any{"user_email_change_enabled": true}, nil)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "true", repo.values[identity.SettingKeyUserEmailChangeEnabled])
}

// TestUpdateSettingsOmittedDefaultUserAPIKeyLimitKeepsStoredValue 检查请求省略默认 Key 上限时保持存储值。
func TestUpdateSettingsOmittedDefaultUserAPIKeyLimitKeepsStoredValue(t *testing.T) {
	h, repo := newStepUpSwitchTestHandler(t, map[string]string{
		identity.SettingKeyDefaultUserAPIKeyLimit: "33",
	})

	rec := doUpdateSettings(t, h, map[string]any{"registration_enabled": true}, nil)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "33", repo.values[identity.SettingKeyDefaultUserAPIKeyLimit])
}

// TestUpdateSettingsExplicitZeroDefaultUserAPIKeyLimit 检查请求发送默认 Key 上限 0 时设置为无限制。
func TestUpdateSettingsExplicitZeroDefaultUserAPIKeyLimit(t *testing.T) {
	h, repo := newStepUpSwitchTestHandler(t, map[string]string{
		identity.SettingKeyDefaultUserAPIKeyLimit: "33",
	})

	rec := doUpdateSettings(t, h, map[string]any{"default_user_api_key_limit": 0}, nil)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "0", repo.values[identity.SettingKeyDefaultUserAPIKeyLimit])
}

// TestUpdateSettingsRejectsNegativeDefaultUserAPIKeyLimit 验证负数设置在进入持久化前被拒绝，原值保持不变。
func TestUpdateSettingsRejectsNegativeDefaultUserAPIKeyLimit(t *testing.T) {
	h, repo := newStepUpSwitchTestHandler(t, map[string]string{
		identity.SettingKeyDefaultUserAPIKeyLimit: "33",
	})

	rec := doUpdateSettings(t, h, map[string]any{"default_user_api_key_limit": -1}, nil)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "INVALID_API_KEY_LIMIT")
	require.Equal(t, "33", repo.values[identity.SettingKeyDefaultUserAPIKeyLimit])
}

// TestUpdateSettingsRejectsDefaultUserAPIKeyLimitAboveDatabaseRange 检查超出数据库 INTEGER 范围的默认值在保存前被拒绝，这类值会导致注册失败。
func TestUpdateSettingsRejectsDefaultUserAPIKeyLimitAboveDatabaseRange(t *testing.T) {
	h, repo := newStepUpSwitchTestHandler(t, map[string]string{
		identity.SettingKeyDefaultUserAPIKeyLimit: "33",
	})

	rec := doUpdateSettings(t, h, map[string]any{"default_user_api_key_limit": identity.MaxUserAPIKeyLimit + 1}, nil)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "INVALID_API_KEY_LIMIT")
	require.Equal(t, "33", repo.values[identity.SettingKeyDefaultUserAPIKeyLimit])
}

func TestUpdateSettingsForwardedClientIPHeadersOmittedPreservesAndEmptyClears(t *testing.T) {
	h, repo := newStepUpSwitchTestHandler(t, map[string]string{
		runtimeconfig.SettingKeyForwardedClientIPHeaders: `["X-Cdn-Ip","True-Client-Ip"]`,
	})

	preserved := doUpdateSettings(t, h, map[string]any{"registration_enabled": true}, nil)
	require.Equal(t, http.StatusOK, preserved.Code)
	require.JSONEq(t, `["X-Cdn-Ip","True-Client-Ip"]`, repo.values[runtimeconfig.SettingKeyForwardedClientIPHeaders])
	require.Contains(t, preserved.Body.String(), `"forwarded_client_ip_headers":["X-Cdn-Ip","True-Client-Ip"]`)

	cleared := doUpdateSettings(t, h, map[string]any{"forwarded_client_ip_headers": []string{}}, nil)
	require.Equal(t, http.StatusOK, cleared.Code)
	require.JSONEq(t, `[]`, repo.values[runtimeconfig.SettingKeyForwardedClientIPHeaders])
	require.Contains(t, cleared.Body.String(), `"forwarded_client_ip_headers":[]`)
}

func TestUpdateSettingsMalformedForwardedClientIPHeadersRemainFailClosedWhenOmitted(t *testing.T) {
	cfg := &config.Config{Default: config.DefaultConfig{UserConcurrency: 5}}
	repo := &settingHandlerRepoStub{values: map[string]string{
		runtimeconfig.SettingKeyAPIKeyACLTrustForwardedIP: "true",
		runtimeconfig.SettingKeyForwardedClientIPHeaders:  `{"not":"an array"}`,
	}}
	options, store := newCompositeSettingsHTTPFixture(repo, cfg)
	require.ErrorContains(t, provideForwardedSettings(store, cfg).LoadForwardedClientIPSettings(context.Background()), "load forwarded client ip headers")
	require.False(t, cfg.ForwardedClientIPSettings().TrustForwardedIP)
	h := settingshttp.NewHandler(options)

	rec := doUpdateSettings(t, h, map[string]any{"registration_enabled": true}, nil)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "false", repo.values[runtimeconfig.SettingKeyAPIKeyACLTrustForwardedIP])
	require.JSONEq(t, `[]`, repo.values[runtimeconfig.SettingKeyForwardedClientIPHeaders])
	runtimeSettings := cfg.ForwardedClientIPSettings()
	require.False(t, runtimeSettings.TrustForwardedIP)
	require.Empty(t, runtimeSettings.Headers)
	require.Contains(t, rec.Body.String(), `"api_key_acl_trust_forwarded_ip":false`)
	require.Contains(t, rec.Body.String(), `"forwarded_client_ip_headers":[]`)
}

func TestUpdateSettingsRejectsInvalidForwardedClientIPHeader(t *testing.T) {
	h, repo := newStepUpSwitchTestHandler(t, map[string]string{
		runtimeconfig.SettingKeyForwardedClientIPHeaders: `["X-Existing-IP"]`,
	})

	rec := doUpdateSettings(t, h, map[string]any{
		"forwarded_client_ip_headers": []string{"X Invalid"},
	}, nil)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.JSONEq(t, `["X-Existing-IP"]`, repo.values[runtimeconfig.SettingKeyForwardedClientIPHeaders])
}

// TestUpdateSettingsRejectsDeprecatedAdvancedSchedulerFields 检查已停用的 OpenAI 实验调度字段被拒绝，面板据此提示保存失败。
func TestUpdateSettingsRejectsDeprecatedAdvancedSchedulerFields(t *testing.T) {
	for _, field := range []string{
		"advanced_scheduler_enabled",
		"openai_advanced_scheduler_enabled",
		"openai_advanced_scheduler_weight_priority",
	} {
		t.Run(field, func(t *testing.T) {
			h, _ := newStepUpSwitchTestHandler(t, map[string]string{})

			rec := doUpdateSettings(t, h, map[string]any{field: true}, nil)

			require.Equal(t, http.StatusBadRequest, rec.Code)
			require.Contains(t, rec.Body.String(), "DEPRECATED_ADVANCED_SCHEDULER_SETTING")
			require.Contains(t, rec.Body.String(), field)
		})
	}
}
