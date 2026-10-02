package httpapi

import (
	gatewayprovider "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	"github.com/TokenFlux/TokenRouter/internal/upstream"

	forwardcore "github.com/TokenFlux/TokenRouter/internal/gateway/forward"
)

// ProjectOpenAIFailoverError 提取平台确认的错误展示信息，HTTP 层应用展示规则。
func ProjectOpenAIFailoverError(err *forwardcore.UpstreamFailoverError) *OpenAIFailoverError {
	if err == nil {
		return nil
	}
	credentialStatus, credentialMessage := CredentialFailoverClientResponse(err)
	return &OpenAIFailoverError{
		Status:                  err.StatusCode,
		ClientStatus:            err.ClientStatusCode,
		ClientMessage:           err.ClientMessage,
		CredentialStatus:        credentialStatus,
		CredentialMessage:       credentialMessage,
		Headers:                 err.ResponseHeaders,
		Body:                    err.ResponseBody,
		TooLarge:                gatewayprovider.IsOpenAIRequestBodyTooLarge(err),
		TooLargeMessage:         forwardcore.OpenAIRequestBodyTooLargeClientMessage,
		ContinuationUnsupported: err.Reason == forwardcore.OpenAIHTTPContinuationUnsupportedReason,
		Credential:              err.IsCredentialFailure(),
		CapacityShed:            gatewayprovider.IsOpenAICapacityShed(err),
		SilentRefusal:           forwardcore.IsOpenAISilentRefusalErrorBody(err.ResponseBody),
		SilentMessage:           forwardcore.OpenAISilentRefusalClientMessage(),
		CyberWarning:            gatewayprovider.IsOpenAICyberWarningPayload(err.ResponseBody, ""),
		CyberMessage:            gatewayprovider.ExtractOpenAICyberWarningMessage(err.ResponseBody, ""),
		UpstreamMessage:         upstream.ExtractErrorMessage(err.ResponseBody),
	}
}
