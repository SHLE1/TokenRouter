package provider

import (
	"time"

	"github.com/TokenFlux/TokenRouter/internal/upstream/usageview"
)

// OllamaUsageFetchInput 保存出站用量查询参数。
type OllamaUsageFetchInput struct {
	ProviderID  int64
	Concurrency int
	ProxyURL    string `json:"-"`
	Cookie      string `json:"-"`
	ObservedAt  time.Time
}

type OllamaUsageObservation = usageview.OllamaUsageObservation
