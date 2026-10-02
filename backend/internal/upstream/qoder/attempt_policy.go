package qoder

import "errors"

// MayRefreshAttempt 允许刷新 401 和 403 认证错误，额度和权益拒绝返回 false。
func MayRefreshAttempt(err error) bool {
	var failure *APIError
	if !errors.As(err, &failure) || failure.IsAgentLimit() || failure.IsEntitlementDenied() {
		return false
	}
	return failure.StatusCode == 401 || failure.StatusCode == 403
}

// MaySwitchAttempt 判断 Qoder 错误是否支持切换提供商，重试窗口和次数由调用方决定。
func MaySwitchAttempt(err error) bool {
	return maySwitchAttempt(err, true)
}

// MaySwitchCompatibleAttempt 判断 Messages/Responses 是否可切换提供商，非 APIError 返回 false。
func MaySwitchCompatibleAttempt(err error) bool {
	return maySwitchAttempt(err, false)
}

func maySwitchAttempt(err error, unknown bool) bool {
	var failure *APIError
	if !errors.As(err, &failure) {
		return unknown
	}
	return failure.IsAgentLimit() || failure.IsEntitlementDenied() || failure.StatusCode == 429 || failure.StatusCode >= 500
}
