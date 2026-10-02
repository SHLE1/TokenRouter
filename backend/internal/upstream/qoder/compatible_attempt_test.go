package qoder

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestQoderGatewayShouldFailoverRetryableUpstreamErrors 检查兼容请求可重试的上游错误，普通错误和 401 返回 false。
func TestQoderGatewayShouldFailoverRetryableUpstreamErrors(t *testing.T) {
	require.True(t, MaySwitchCompatibleAttempt(&APIError{StatusCode: http.StatusTooManyRequests}))
	require.True(t, MaySwitchCompatibleAttempt(&APIError{StatusCode: http.StatusBadGateway, Code: "115"}))
	require.True(t, MaySwitchCompatibleAttempt(&APIError{StatusCode: http.StatusForbidden, Code: "115"}))
	require.True(t, MaySwitchCompatibleAttempt(&APIError{StatusCode: http.StatusForbidden, Code: "112"}))
	require.True(t, MaySwitchCompatibleAttempt(&APIError{StatusCode: http.StatusInternalServerError}))
	require.False(t, MaySwitchCompatibleAttempt(&APIError{StatusCode: http.StatusUnauthorized}))
	require.False(t, MaySwitchCompatibleAttempt(fmt.Errorf("plain error")))
}
