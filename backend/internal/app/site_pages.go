package app

import (
	"github.com/TokenFlux/TokenRouter/internal/config"
	"github.com/TokenFlux/TokenRouter/internal/site"
	"github.com/TokenFlux/TokenRouter/internal/site/filesystem"
	"github.com/TokenFlux/TokenRouter/internal/site/httpapi"
)

func provideSitePages(cfg *config.Config, settings *site.DisplaySettings) *httpapi.PageHandler {
	return httpapi.NewPageHandler(site.NewPages(filesystem.New(cfg.Pricing.DataDir), settings))
}
