package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"slices"

	"github.com/TokenFlux/TokenRouter/internal/provider"
	providerpostgres "github.com/TokenFlux/TokenRouter/internal/provider/postgres"
	provideradapter "github.com/TokenFlux/TokenRouter/internal/provider/provider"

	"github.com/TokenFlux/TokenRouter/internal/config"
	"github.com/TokenFlux/TokenRouter/internal/infra/postgres"
	routingpostgres "github.com/TokenFlux/TokenRouter/internal/routing/postgres"
)

// PreviewGroupMigration 只连接数据库执行预检，不启动迁移、网关或后台任务。
func PreviewGroupMigration(ctx context.Context, cfg *config.Config, output io.Writer) error {
	db, err := postgres.Open(cfg.Database.DSNWithTimezone(cfg.Timezone), false)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	report, err := routingpostgres.PreviewPricingMigration(ctx, db)
	if err != nil {
		return err
	}
	warnings, err := providerpostgres.PreviewModelScopeMigration(ctx, db, func(value *provider.Record, model string) bool {
		return value.FinalModelWhitelisted(model, provideradapter.ModelDefaults(), provideradapter.ModelRules(value))
	}, configuredModelsOutsideDefaults)
	if err != nil {
		return err
	}
	preview := struct {
		routingpostgres.PricingMigrationPreview
		ProviderModelWarnings []providerpostgres.ModelScopeMigrationWarning `json:"provider_model_warnings"`
	}{report, warnings}
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(preview); err != nil {
		return err
	}
	if report.Blocked {
		return fmt.Errorf("group migration blocked by incomparable pricing rules; resolve the reported conflicts before upgrading")
	}
	return nil
}

// configuredModelsOutsideDefaults 单列显式扩展，避免未曾调用过的自定义模型在预检中遗漏。
func configuredModelsOutsideDefaults(value *provider.Record) []string {
	defaults := provideradapter.ModelDefaults()
	base := &provider.Record{Platform: value.Platform, Type: value.Type, ParentProviderID: value.ParentProviderID, Credentials: maps.Clone(value.Credentials), Extra: value.Extra}
	delete(base.Credentials, "model_mapping")
	delete(base.Credentials, "model_whitelist")
	known := base.GetConfiguredRequestModels(defaults)
	var outside []string
	for _, model := range value.GetConfiguredRequestModels(defaults) {
		if !slices.Contains(known, model) {
			outside = append(outside, model)
		}
	}
	return outside
}
