package httpapi

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/TokenFlux/TokenRouter/internal/identity/httpapi/authctx"
	"github.com/TokenFlux/TokenRouter/internal/requestlog"
	"github.com/TokenFlux/TokenRouter/internal/server/httpx"
)

// Handler 返回当前调用者可见的请求摘要。
type Handler struct {
	service    *requestlog.Service
	userErrors func(context.Context) bool
}

func NewHandler(service *requestlog.Service, userErrors func(context.Context) bool) *Handler {
	return &Handler{service: service, userErrors: userErrors}
}

func (h *Handler) Find(c *gin.Context)      { h.find(c, false) }
func (h *Handler) FindAdmin(c *gin.Context) { h.find(c, true) }

func (h *Handler) find(c *gin.Context, admin bool) {
	subject, ok := authctx.GetAuthSubjectFromContext(c)
	if !ok {
		httpx.Unauthorized(c, "User not authenticated")
		return
	}
	id := strings.TrimSpace(c.Query("request_id"))
	if id == "" {
		id = strings.TrimSpace(c.Param("request_id"))
	}
	if id == "" || len(id) > 255 || strings.ContainsAny(id, "\r\n\x00") {
		httpx.BadRequest(c, "Invalid request_id")
		return
	}
	items, err := h.service.Find(c.Request.Context(), id, subject.UserID, admin)
	if err != nil {
		httpx.ErrorFrom(c, err)
		return
	}
	more := len(items) > 100
	if more {
		items = items[:100]
	}
	if !admin {
		allowErrors := h.userErrors != nil && h.userErrors(c.Request.Context())
		for i := range items {
			if !allowErrors {
				items[i].Errors = nil
			}
			items[i].ProviderID = 0
			items[i].Aliases = nil
			items[i].AuditIDs = nil
			for j := range items[i].Attempts {
				items[i].Attempts[j].ProviderID = 0
				items[i].Attempts[j].RequestID = ""
			}
		}
	}
	httpx.Success(c, gin.H{"items": items, "has_more": more})
}

func (h *Handler) Health(c *gin.Context) {
	health, err := h.service.Health()
	if err != nil {
		httpx.ErrorFrom(c, err)
		return
	}
	httpx.Success(c, health)
}
