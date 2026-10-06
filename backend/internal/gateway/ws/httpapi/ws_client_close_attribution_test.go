package httpapi

import (
	"context"
	"errors"
	"fmt"
	"testing"

	gatewayhttp "github.com/TokenFlux/TokenRouter/internal/gateway/httpapi"

	gatewayws "github.com/TokenFlux/TokenRouter/internal/gateway/ws"
	coderws "github.com/coder/websocket"
	"github.com/stretchr/testify/require"
)

// 这些测试检查 Responses WebSocket 的客户端结束归因（#6105）。
// 客户端正常关闭可返回 coderws.CloseError{1000}，中途取消可返回 context.Canceled 并以 1001 关闭。
// 两者都按客户端结束处理，提供商故障由 shouldReportOpenAIWSProxyProviderFailure 判断，供健康状态和 scheduler.ReportResult 使用。

// TestOpenAIWSIngressEndedByClient_BareNormalClosureIsNotProviderFailure 验证缺陷主复现之一：客户端干净关闭。底层 conn.Read 的错误被 ReadOpenAIWSClientMessage
// 原样返回，没有任何地方把它包成 *OpenAIWSClientCloseError，所以旧断言看不见它。
func TestOpenAIWSIngressEndedByClient_BareNormalClosureIsNotProviderFailure(t *testing.T) {
	err := coderws.CloseError{Code: coderws.StatusNormalClosure, Reason: "client done"}

	// 裸 CloseError 与 OpenAIWSClientCloseError 的类型不同，关闭码都为 1000。
	var closeErr *gatewayhttp.OpenAIWSClientCloseError
	require.False(t, errors.As(err, &closeErr),
		"裸 coderws.CloseError 不是 *OpenAIWSClientCloseError，旧的 errors.As 必然为假")
	// 而按关闭码读，它确实是 1000。
	require.Equal(t, coderws.StatusNormalClosure, coderws.CloseStatus(err))

	require.True(t, gatewayhttp.ResponsesWSEndedByClient(err, gatewayhttp.ResponsesWSCloseInfo(err)))
}

// TestOpenAIWSIngressEndedByClient_WrappedBareNormalClosureIsNotProviderFailure 验证带包装的正常关闭错误也归为客户端结束。
func TestOpenAIWSIngressEndedByClient_WrappedBareNormalClosureIsNotProviderFailure(t *testing.T) {
	err := fmt.Errorf("ingress turn 3: %w",
		coderws.CloseError{Code: coderws.StatusNormalClosure, Reason: "client done"})

	require.True(t, gatewayhttp.ResponsesWSEndedByClient(err, gatewayhttp.ResponsesWSCloseInfo(err)))
}

// TestOpenAIWSIngressEndedByClient_ClientCancelDuringTurnIsNotProviderFailure 验证缺陷主复现之二：客户端中途断开。ReadOpenAIWSClientMessage 在 controlCtx.Done()
// 分支用 StatusGoingAway 收尾并把 context.Canceled 作为 cause，所以「只认 1000」
// 这一条判据根本匹配不到它。
func TestOpenAIWSIngressEndedByClient_ClientCancelDuringTurnIsNotProviderFailure(t *testing.T) {
	err := gatewayhttp.NewOpenAIWSClientCloseError(
		coderws.StatusGoingAway, "websocket request canceled", context.Canceled)

	// 前提：状态码是 1001 不是 1000，旧判据必然放行到提供商归因。
	var closeErr *gatewayhttp.OpenAIWSClientCloseError
	require.ErrorAs(t, err, &closeErr)
	require.Equal(t, coderws.StatusGoingAway, closeErr.StatusCode())
	require.NotEqual(t, coderws.StatusNormalClosure, closeErr.StatusCode())

	require.True(t, gatewayhttp.ResponsesWSEndedByClient(err, gatewayhttp.ResponsesWSCloseInfo(err)))
}

// TestOpenAIWSIngressEndedByClient_GatewayNormalClosureStillRecognised 验证既有行为不得回退：网关自己用 1000 收尾（inter-turn idle timeout 就是这条）
// 原本就被认作正常关闭。
func TestOpenAIWSIngressEndedByClient_GatewayNormalClosureStillRecognised(t *testing.T) {
	err := gatewayhttp.NewOpenAIWSClientCloseError(
		coderws.StatusNormalClosure, "websocket idle timeout", context.DeadlineExceeded)

	require.True(t, gatewayhttp.ResponsesWSEndedByClient(err, gatewayhttp.ResponsesWSCloseInfo(err)))
}

// TestOpenAIWSIngressEndedByClient_GoingAwayWithoutCancellationStillReported 验证收窄证明：1001 本身不足以豁免。网关也会因自身原因用 GoingAway 收场，
// 客户端取消那一支已由 context.Canceled 覆盖，无需整类放行。
func TestOpenAIWSIngressEndedByClient_GoingAwayWithoutCancellationStillReported(t *testing.T) {
	err := gatewayhttp.NewOpenAIWSClientCloseError(
		coderws.StatusGoingAway, "upstream going away", errors.New("upstream closed session"))

	require.False(t, gatewayhttp.ResponsesWSEndedByClient(err, gatewayhttp.ResponsesWSCloseInfo(err)))
	require.True(t, gatewayws.EntryShouldReportFailure(err), "真实上游故障仍须归因提供商")
}

// TestOpenAIWSIngressEndedByClient_AbnormalClosuresStillReportProviderFailure 验证异常关闭进入提供商故障判断。
// 调用顺序为先检查 openAIWSIngressEndedByClient，再调用 shouldReportOpenAIWSProxyProviderFailure。
func TestOpenAIWSIngressEndedByClient_AbnormalClosuresStillReportProviderFailure(t *testing.T) {
	cases := []struct {
		name string
		err  error
	}{
		{
			name: "upstream_policy_violation",
			err: gatewayhttp.NewOpenAIWSClientCloseError(
				coderws.StatusPolicyViolation, "upstream websocket authentication failed",
				errors.New("upstream rejected credentials")),
		},
		{
			name: "upstream_internal_error",
			err: gatewayhttp.NewOpenAIWSClientCloseError(
				coderws.StatusInternalError, "upstream websocket proxy failed", nil),
		},
		{
			name: "bare_abnormal_closure",
			err:  coderws.CloseError{Code: coderws.StatusAbnormalClosure, Reason: "connection reset"},
		},
		{
			name: "generic_read_failure",
			err:  errors.New("upstream websocket read failed"),
		},
		{
			// 空闲超时之外的 deadline 按停滞故障处理。
			name: "deadline_without_normal_close",
			err:  fmt.Errorf("upstream stalled: %w", context.DeadlineExceeded),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.False(t, gatewayhttp.ResponsesWSEndedByClient(tc.err, gatewayhttp.ResponsesWSCloseInfo(tc.err)))
			require.True(t, gatewayws.EntryShouldReportFailure(tc.err))
		})
	}
}

// TestOpenAIWSIngressEndedByClient_MatchesCloseCodeReportedInLog 验证日志关闭码与客户端结束归因一致。
// summarizeWSCloseErrorForLog 一直用 coderws.CloseStatus 读关闭码，这正是缺陷时期
// WARN 打印 close_status=1000(StatusNormalClosure) 却同时把提供商记为故障的原因。
// 以后任何一侧改了读法，这条会红。
func TestOpenAIWSIngressEndedByClient_MatchesCloseCodeReportedInLog(t *testing.T) {
	errs := []error{
		coderws.CloseError{Code: coderws.StatusNormalClosure, Reason: "client done"},
		fmt.Errorf("ingress turn 3: %w", coderws.CloseError{Code: coderws.StatusNormalClosure}), gatewayhttp.NewOpenAIWSClientCloseError(coderws.StatusNormalClosure, "websocket idle timeout", context.DeadlineExceeded), gatewayhttp.NewOpenAIWSClientCloseError(coderws.StatusGoingAway, "websocket request canceled", context.Canceled),
		coderws.CloseError{Code: coderws.StatusAbnormalClosure, Reason: "connection reset"},
		errors.New("upstream websocket read failed"),
	}

	for _, err := range errs {
		t.Run(err.Error(), func(t *testing.T) {
			closeStatus, _ := gatewayhttp.SummarizeWSCloseErrorForLog(err)
			if closeStatus == "1000(StatusNormalClosure)" {
				require.True(t, gatewayhttp.ResponsesWSEndedByClient(err, gatewayhttp.ResponsesWSCloseInfo(err)),
					"日志按 1000 归类为正常关闭，归因侧不得同时判为提供商故障")
			}
		})
	}
}
