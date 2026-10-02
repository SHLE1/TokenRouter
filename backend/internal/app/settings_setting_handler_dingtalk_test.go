package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TokenFlux/TokenRouter/internal/config"
	"github.com/TokenFlux/TokenRouter/internal/identity"

	response "github.com/TokenFlux/TokenRouter/internal/server/httpx"
	settingshttp "github.com/TokenFlux/TokenRouter/internal/settings/httpapi"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// dingtalkSettingsRepoStub 复用 settingHandlerRepoStub（已在 setting_handler_auth_source_defaults_test.go 定义）

func newDingTalkSettingsHandler() (*settingshttp.Handler, *settingHandlerRepoStub) {
	repo := &settingHandlerRepoStub{values: map[string]string{}}
	options, _ := newCompositeSettingsHTTPFixture(repo, &config.Config{Default: config.DefaultConfig{UserConcurrency: 5}})
	handler := settingshttp.NewHandler(options)
	return handler, repo
}

// baseValidDingTalkBody 返回一个可以通过所有校验的最小合法 body。
func baseValidDingTalkBody() map[string]any {
	return map[string]any{
		"dingtalk_connect_enabled":                 true,
		"dingtalk_connect_client_id":               "test-client-id",
		"dingtalk_connect_client_secret":           "test-client-secret",
		"dingtalk_connect_redirect_url":            "https://example.com/auth/dingtalk/callback",
		"dingtalk_connect_corp_restriction_policy": "none",
	}
}

// TestSettingsPUT_DingTalk_V3_InternalOnlyAllowsEmptyCorpID 检查 internal_only 策略下
// internal_corp_id 为空时保存成功并返回 200。
func TestSettingsPUT_DingTalk_V3_InternalOnlyAllowsEmptyCorpID(t *testing.T) {
	handler, _ := newDingTalkSettingsHandler()

	body := baseValidDingTalkBody()
	body["dingtalk_connect_corp_restriction_policy"] = "internal_only"
	body["dingtalk_connect_internal_corp_id"] = "" // 空值现在合法

	rawBody, err := json.Marshal(body)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", bytes.NewReader(rawBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateSettings(c)

	require.Equal(t, http.StatusOK, rec.Code)
}

// TestSettingsPUT_DingTalk_HappyPath_None 检查 none 策略下保存返回 200。
func TestSettingsPUT_DingTalk_HappyPath_None(t *testing.T) {
	handler, _ := newDingTalkSettingsHandler()

	body := baseValidDingTalkBody()
	body["dingtalk_connect_corp_restriction_policy"] = "none"

	rawBody, err := json.Marshal(body)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", bytes.NewReader(rawBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateSettings(c)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp response.Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	data, ok := resp.Data.(map[string]any)
	require.True(t, ok)
	require.Equal(t, true, data["dingtalk_connect_enabled"])
}

// TestSettingsPUT_DingTalk_HappyPath_InternalOnly_WithCorpID 检查 internal_only 策略提供 corp_id 时返回 200。
func TestSettingsPUT_DingTalk_HappyPath_InternalOnly_WithCorpID(t *testing.T) {
	handler, _ := newDingTalkSettingsHandler()

	body := baseValidDingTalkBody()
	body["dingtalk_connect_corp_restriction_policy"] = "internal_only"
	body["dingtalk_connect_internal_corp_id"] = "ding-corp-123"

	rawBody, err := json.Marshal(body)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", bytes.NewReader(rawBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateSettings(c)

	require.Equal(t, http.StatusOK, rec.Code)
}

// TestSettingsPUT_DingTalk_BypassRegistration_RoundTrip 验证 bypass_registration 字段 save+load。
// bypass_registration 在 internal_only 策略下生效，其他策略写入 false。
func TestSettingsPUT_DingTalk_BypassRegistration_RoundTrip(t *testing.T) {
	handler, _ := newDingTalkSettingsHandler()

	body := baseValidDingTalkBody()
	body["dingtalk_connect_corp_restriction_policy"] = "internal_only"
	body["dingtalk_connect_bypass_registration"] = true

	rawBody, err := json.Marshal(body)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", bytes.NewReader(rawBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateSettings(c)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp response.Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	data, ok := resp.Data.(map[string]any)
	require.True(t, ok)
	require.Equal(t, true, data["dingtalk_connect_bypass_registration"])
}

// TestSettingsPUT_DingTalk_Disabled_SkipsValidation 检查禁用钉钉时跳过企业信息校验并返回 200。
// 测试将 client_id 留空，启用时该输入会返回 Client ID is required when enabled，
// 禁用时保存成功。
func TestSettingsPUT_DingTalk_Disabled_SkipsValidation(t *testing.T) {
	handler, _ := newDingTalkSettingsHandler()

	body := map[string]any{
		"dingtalk_connect_enabled":                 false,
		"dingtalk_connect_client_id":               "", // 这种空值在 enabled=true 时会被 400 拒绝
		"dingtalk_connect_corp_restriction_policy": "internal_only",
	}

	rawBody, err := json.Marshal(body)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", bytes.NewReader(rawBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateSettings(c)

	require.Equal(t, http.StatusOK, rec.Code)
}

// TestSettingsPUT_DingTalk_SyncFlags_InternalOnly_RoundTrip 验证三个 sync 开关在 internal_only 下可正常 save+load。
func TestSettingsPUT_DingTalk_SyncFlags_InternalOnly_RoundTrip(t *testing.T) {
	handler, _ := newDingTalkSettingsHandler()

	body := baseValidDingTalkBody()
	body["dingtalk_connect_corp_restriction_policy"] = "internal_only"
	body["dingtalk_connect_sync_corp_email"] = true
	body["dingtalk_connect_sync_display_name"] = true
	body["dingtalk_connect_sync_dept"] = true

	rawBody, err := json.Marshal(body)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", bytes.NewReader(rawBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateSettings(c)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp response.Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	data, ok := resp.Data.(map[string]any)
	require.True(t, ok)
	require.Equal(t, true, data["dingtalk_connect_sync_corp_email"], "sync_corp_email should be true for internal_only")
	require.Equal(t, true, data["dingtalk_connect_sync_display_name"], "sync_display_name should be true for internal_only")
	require.Equal(t, true, data["dingtalk_connect_sync_dept"], "sync_dept should be true for internal_only")
}

// TestSettingsPUT_DingTalk_SyncFlags_PolicyNone_CoercedToFalse 验证 policy=none 时三个 sync 开关被 coerce 为 false。
func TestSettingsPUT_DingTalk_SyncFlags_PolicyNone_CoercedToFalse(t *testing.T) {
	handler, _ := newDingTalkSettingsHandler()

	body := baseValidDingTalkBody()
	body["dingtalk_connect_corp_restriction_policy"] = "none"
	body["dingtalk_connect_sync_corp_email"] = true
	body["dingtalk_connect_sync_display_name"] = true
	body["dingtalk_connect_sync_dept"] = true

	rawBody, err := json.Marshal(body)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", bytes.NewReader(rawBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateSettings(c)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp response.Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	data, ok := resp.Data.(map[string]any)
	require.True(t, ok)
	require.Equal(t, false, data["dingtalk_connect_sync_corp_email"], "sync_corp_email must be coerced to false when policy=none")
	require.Equal(t, false, data["dingtalk_connect_sync_display_name"], "sync_display_name must be coerced to false when policy=none")
	require.Equal(t, false, data["dingtalk_connect_sync_dept"], "sync_dept must be coerced to false when policy=none")
}

// TestSettingsPUT_DingTalk_StaleWhitelist_CoercedToNone 检查兼容的 whitelist 输入：
// 管理员通过 API 提交 corp_restriction_policy=whitelist 时，
// 服务端将其转换为 none 后保存。
func TestSettingsPUT_DingTalk_StaleWhitelist_CoercedToNone(t *testing.T) {
	handler, repo := newDingTalkSettingsHandler()

	body := baseValidDingTalkBody()
	body["dingtalk_connect_corp_restriction_policy"] = "whitelist"

	rawBody, err := json.Marshal(body)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", bytes.NewReader(rawBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateSettings(c)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "none", repo.values[identity.SettingKeyDingTalkConnectCorpRestrictionPolicy],
		"stale whitelist 应在写入路径被 coerce 为 none")
}

// TestSettingsPUT_DingTalk_SyncAttrKey_RoundTrip 验证 3 个 attr key 字段 save+load + 空值 fallback 到默认值。
func TestSettingsPUT_DingTalk_SyncAttrKey_RoundTrip(t *testing.T) {
	t.Run("custom_attr_keys_saved", func(t *testing.T) {
		handler, repo := newDingTalkSettingsHandler()

		body := baseValidDingTalkBody()
		body["dingtalk_connect_corp_restriction_policy"] = "internal_only"
		body["dingtalk_connect_sync_corp_email"] = true
		body["dingtalk_connect_sync_display_name"] = true
		body["dingtalk_connect_sync_dept"] = true
		body["dingtalk_connect_sync_corp_email_attr_key"] = "my_email_attr"
		body["dingtalk_connect_sync_display_name_attr_key"] = "my_name_attr"
		body["dingtalk_connect_sync_dept_attr_key"] = "my_dept_attr"

		rawBody, err := json.Marshal(body)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", bytes.NewReader(rawBody))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.UpdateSettings(c)

		require.Equal(t, http.StatusOK, rec.Code)

		// 验证写入 DB 的 key
		require.Equal(t, "my_email_attr", repo.values[identity.SettingKeyDingTalkConnectSyncCorpEmailAttrKey])
		require.Equal(t, "my_name_attr", repo.values[identity.SettingKeyDingTalkConnectSyncDisplayNameAttrKey])
		require.Equal(t, "my_dept_attr", repo.values[identity.SettingKeyDingTalkConnectSyncDeptAttrKey])

		// 验证响应中的 attr key
		var resp response.Response
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		data, ok := resp.Data.(map[string]any)
		require.True(t, ok)
		require.Equal(t, "my_email_attr", data["dingtalk_connect_sync_corp_email_attr_key"])
		require.Equal(t, "my_name_attr", data["dingtalk_connect_sync_display_name_attr_key"])
		require.Equal(t, "my_dept_attr", data["dingtalk_connect_sync_dept_attr_key"])
	})

	t.Run("empty_attr_keys_fallback_to_defaults", func(t *testing.T) {
		handler, repo := newDingTalkSettingsHandler()

		body := baseValidDingTalkBody()
		body["dingtalk_connect_corp_restriction_policy"] = "internal_only"
		// 省略 attr key 时，写入层使用默认值。
		body["dingtalk_connect_sync_corp_email_attr_key"] = ""
		body["dingtalk_connect_sync_display_name_attr_key"] = ""
		body["dingtalk_connect_sync_dept_attr_key"] = ""

		rawBody, err := json.Marshal(body)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", bytes.NewReader(rawBody))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.UpdateSettings(c)

		require.Equal(t, http.StatusOK, rec.Code)

		// 空值应 fallback 到默认值并持久化
		require.Equal(t, "dingtalk_email", repo.values[identity.SettingKeyDingTalkConnectSyncCorpEmailAttrKey])
		require.Equal(t, "dingtalk_name", repo.values[identity.SettingKeyDingTalkConnectSyncDisplayNameAttrKey])
		require.Equal(t, "dingtalk_department", repo.values[identity.SettingKeyDingTalkConnectSyncDeptAttrKey])
	})
}
