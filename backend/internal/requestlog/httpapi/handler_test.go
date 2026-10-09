package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/TokenFlux/TokenRouter/internal/identity/httpapi/authctx"
	"github.com/TokenFlux/TokenRouter/internal/infra/telemetry"
	"github.com/TokenFlux/TokenRouter/internal/requestlog"
)

type requestRepositoryStub struct {
	userID int64
	admin  bool
}

func (r *requestRepositoryStub) Save(context.Context, []telemetry.RequestRecord) error { return nil }

func (r *requestRepositoryStub) Cleanup(context.Context, time.Time) error { return nil }

func (r *requestRepositoryStub) Find(_ context.Context, _ string, userID int64, admin bool) ([]requestlog.Detail, error) {
	r.userID, r.admin = userID, admin
	return []requestlog.Detail{{RequestRecord: telemetry.RequestRecord{RequestID: "request", UserID: 42, ProviderID: 99, Aliases: []telemetry.RequestAlias{{Kind: "upstream", Value: "supplier"}}, Attempts: []telemetry.RequestAttempt{{Number: 1, ProviderID: 99, RequestID: "supplier"}}}, Errors: []requestlog.Failure{{ID: 7, Status: 500}}, AuditIDs: []int64{8}}}, nil
}

// TestFindUsesAuthenticatedScopeAndRedactsDetails 检查查询参数无法切换身份或管理员视图。
func TestFindUsesAuthenticatedScopeAndRedactsDetails(t *testing.T) {
	repo := &requestRepositoryStub{}
	handler := NewHandler(requestlog.NewService(repo, nil, 30, nil), func(context.Context) bool { return false })
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set(authctx.ContextKeyUser, authctx.AuthSubject{UserID: 42}); c.Next() })
	router.GET("/requests", handler.Find)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/requests?request_id=request&user_id=99&admin=true", nil))
	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, int64(42), repo.userID)
	require.False(t, repo.admin)
	var body struct {
		Data struct {
			Items []requestlog.Detail `json:"items"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	require.Len(t, body.Data.Items, 1)
	item := body.Data.Items[0]
	require.Zero(t, item.ProviderID)
	require.Empty(t, item.Aliases)
	require.Empty(t, item.Errors)
	require.Empty(t, item.AuditIDs)
	require.Zero(t, item.Attempts[0].ProviderID)
	require.Empty(t, item.Attempts[0].RequestID)
}
