package httpapi

import "github.com/TokenFlux/TokenRouter/internal/server/httpx"

func HTTPStatusToGoogleStatus(status int) string { return httpx.HTTPStatusToGoogleStatus(status) }
