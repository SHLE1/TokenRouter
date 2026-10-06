package app

import (
	"net/http"
	"testing"

	"github.com/TokenFlux/TokenRouter/internal/gateway/ws"
	"github.com/stretchr/testify/require"
)

func TestSettingsResponsesWSPatchAndNull(t *testing.T) {
	h, repo := newStepUpSwitchTestHandler(t, map[string]string{ws.SettingKey: `{"read_timeout_seconds":45,"max_ingress_connections_per_api_key":5}`})
	response := doUpdateSettings(t, h, map[string]any{"responses_ws": map[string]any{"read_timeout_seconds": nil, "max_ingress_connections_per_api_key": 0}}, nil)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.JSONEq(t, `{"max_ingress_connections_per_api_key":0}`, repo.values[ws.SettingKey])
	require.Contains(t, response.Body.String(), `"responses_ws_effective"`)
	response = doUpdateSettings(t, h, map[string]any{"responses_ws": nil}, nil)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.JSONEq(t, `{}`, repo.values[ws.SettingKey])
}

func TestSettingsResponsesWSInvalidPatchDoesNotWrite(t *testing.T) {
	h, repo := newStepUpSwitchTestHandler(t, map[string]string{ws.SettingKey: `{"read_timeout_seconds":45}`})
	response := doUpdateSettings(t, h, map[string]any{"responses_ws": map[string]any{"max_conns_per_provider": 1}}, nil)
	require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
	require.JSONEq(t, `{"read_timeout_seconds":45}`, repo.values[ws.SettingKey])
}
