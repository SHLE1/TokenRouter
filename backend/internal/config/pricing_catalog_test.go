package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPricingCatalogLegacySourceMigration(t *testing.T) {
	for _, source := range []string{
		"https://raw.githubusercontent.com/Wei-Shaw/model-price-repo/main/model_prices_and_context_window.json",
		"https://raw.githubusercontent.com/Wei-Shaw/model-price-repo/refs/heads/main//model_prices_and_context_window.json",
	} {
		cfg := Config{Pricing: PricingConfig{RemoteURL: source, HashURL: "old-hash", OverrideFile: "custom.json"}}
		cfg.normalizePricingCatalogSource()
		cfg.normalizePricingCatalogSource()
		require.Equal(t, "https://models.dev/catalog.json", cfg.Pricing.RemoteURL)
		require.Empty(t, cfg.Pricing.HashURL)
		require.Equal(t, "custom.json", cfg.Pricing.OverrideFile)
		require.Equal(t, []string{"models.dev"}, cfg.Security.URLAllowlist.PricingHosts)
	}
	custom := "https://mirror.example/catalog.json?source=raw.githubusercontent.com/wei-shaw/model-price-repo/main/model_prices_and_context_window.json"
	cfg := Config{Pricing: PricingConfig{RemoteURL: custom}}
	cfg.normalizePricingCatalogSource()
	require.Equal(t, custom, cfg.Pricing.RemoteURL)
}
