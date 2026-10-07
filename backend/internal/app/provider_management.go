package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	gatewayprovider "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	catalogprovider "github.com/TokenFlux/TokenRouter/internal/modelcatalog/provider"

	routingprovider "github.com/TokenFlux/TokenRouter/internal/routing/provider"

	usagetypes "github.com/TokenFlux/TokenRouter/internal/usage"
	usagepostgres "github.com/TokenFlux/TokenRouter/internal/usage/postgres"

	"github.com/TokenFlux/TokenRouter/internal/app/lifecycle"
	"github.com/TokenFlux/TokenRouter/internal/pkg/timezone"
	"github.com/TokenFlux/TokenRouter/internal/provider"
	providerhttp "github.com/TokenFlux/TokenRouter/internal/provider/httpapi"
	provideradapter "github.com/TokenFlux/TokenRouter/internal/provider/provider"
	"github.com/TokenFlux/TokenRouter/internal/routing"
)

// provideProviderManagement 为提供商管理 HTTP 接口绑定管理用例和展示参数。
func provideProviderManagement(admin *provider.Admin, presenter *providerhttp.RuntimePresenter, ollama *provider.OllamaCloudUsageService, privacy *provider.PrivacyService, probes *provider.GrokImportProbeScheduler, quota *provider.GrokQuotaService, managed *provider.ManagedRefreshService, tasks *lifecycle.Tasks, recovery *provider.RecoveryService, listing *provider.ManagementList, catalog *routing.AdminCatalog, tier *provider.TierManagement, models *provider.ModelSyncService, usage *usagepostgres.Store, calendar timezone.Calendar) *providerhttp.ManagementHandler {
	query := func(ctx context.Context, id int64, start, end time.Time) (*usagetypes.ProviderUsageStatsResponse, error) {
		value, err := usage.GetProviderUsageStats(ctx, id, start, end)
		if err != nil {
			return nil, fmt.Errorf("get provider usage stats failed: %w", err)
		}
		return value, nil
	}
	return providerhttp.NewManagementHandler(admin, providerhttp.ManagementOptions{Models: models, Reports: providerhttp.ProviderReportOptions{Now: calendar.Now, StartOfDay: calendar.StartOfDay, Query: query}, Tier: tier, Catalog: catalog, ModelDefaults: provideradapter.ModelDefaults(), ModelSupports: func(ctx context.Context, v *provider.Record, id string) bool {
		return (gatewayprovider.ModelPolicy{Record: v}).Supports(ctx, id)
	}, List: listing, RuntimePresenter: presenter, Recovery: recovery, Batch: provider.NewManagementBatch(admin, managed, provider.ManagementCreationOptions{Privacy: privacy, Background: tasks.Go, AfterCreate: func(v *provider.Record) {
		if v == nil {
			return
		}
		snapshot := v.RoutingSnapshot()
		probes.Schedule(grokImportQuotaProbe{source: quota}, &snapshot)
	}, Error: slog.Error}), Managed: managed, Presenter: presenter, Ollama: ollama, Privacy: privacy, AfterCreate: func(v *provider.Record) {
		if v == nil {
			return
		}
		snapshot := v.RoutingSnapshot()
		probes.Schedule(grokImportQuotaProbe{source: quota}, &snapshot)
	}})
}

// provideAdminModelCatalog 在每次查询时读取平台快照。
func provideAdminModelCatalog(catalog *catalogprovider.Service) *routing.AdminCatalog {
	return routing.NewAdminCatalog(routingprovider.AdminCatalogOptions(catalog))
}
