package app

import (
	forwardcore "github.com/TokenFlux/TokenRouter/internal/gateway/forward"
	gatewayhttp "github.com/TokenFlux/TokenRouter/internal/gateway/httpapi"
	wshttp "github.com/TokenFlux/TokenRouter/internal/gateway/ws/httpapi"

	coderws "github.com/coder/websocket"
	"github.com/gin-gonic/gin"
)

func closeOpenAIWSFailoverExhausted(c *gin.Context, conn *coderws.Conn, err *forwardcore.UpstreamFailoverError) {
	gatewayhttp.CloseResponsesWSFailure(c, conn, wshttp.FailoverPresentation(err), gatewayhttp.MarkOpsStreamFailure)
}
