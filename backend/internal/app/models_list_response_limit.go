package app

import "github.com/TokenFlux/TokenRouter/internal/config"

// resolveModelsListReadLimit 返回模型目录读取上限，缺省或非正值使用默认值。
func resolveModelsListReadLimit(cfg *config.Config) int64 {
	if cfg != nil && cfg.Gateway.ModelsListReadMaxBytes > 0 {
		return cfg.Gateway.ModelsListReadMaxBytes
	}
	return config.DefaultModelsListReadMaxBytes
}
