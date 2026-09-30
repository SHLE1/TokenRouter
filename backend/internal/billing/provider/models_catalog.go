package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/billing/pricing"
	"github.com/TokenFlux/TokenRouter/internal/modelcatalog"
)

// AttributeSnapshot 只供管理与展示使用，与价格在同一个锁内发布。
type AttributeSnapshot struct {
	Items       []modelcatalog.Entry `json:"items"`
	Version     string               `json:"version"`
	LastUpdated time.Time            `json:"last_updated"`
	LastError   string               `json:"last_error,omitempty"`
}

// AttributesSnapshot 返回当前目录的独立属性快照，不触发网络请求。
func (s *PricingService) AttributesSnapshot() AttributeSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := AttributeSnapshot{Items: s.modelCatalog.Rows(), LastUpdated: s.lastUpdated, LastError: s.lastCatalogError}
	if s.modelCatalog != nil {
		result.Version = s.modelCatalog.Version
	}
	return result
}

// ModelAttributes 仅解析明确的模型身份，不借用跨型号的计费回退。
func (s *PricingService) ModelAttributes(model string) modelcatalog.Attributes {
	candidates := []string{model}
	if s.options.ModelLookupCandidates != nil {
		candidates = append(candidates, s.options.ModelLookupCandidates()(model)...)
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, _ := s.modelCatalog.Lookup(s.modelCatalog.IdentityCandidates(model, candidates))
	return entry.Attributes
}

func (s *PricingService) buildModelsCatalog(body []byte) (*modelcatalog.Catalog, map[string]*LiteLLMModelPricing, error) {
	if err := s.ValidateCustomPricingFiles(); err != nil {
		return nil, nil, err
	}
	catalog, err := modelcatalog.Parse(body)
	if err != nil {
		return nil, nil, err
	}
	raw := pricing.ModelsDevPrices(catalog)
	// 本地补充仅填补模型或媒体维度缺口，不覆盖远程已有的普通 token 报价。
	if data, readErr := os.ReadFile(s.options.FallbackFile); readErr == nil {
		var supplement map[string]json.RawMessage
		if err := json.Unmarshal(data, &supplement); err != nil {
			return nil, nil, err
		}
		// 先应用精确键，再为同一原厂记录的其他查价键补齐空缺。
		// 这样两种名称显式配置了不同补充价时，不受 map 遍历顺序影响。
		for model, entry := range supplement {
			if err := mergeMediaSupplement(raw, model, entry); err != nil {
				return nil, nil, err
			}
		}
		for model, entry := range supplement {
			for _, alias := range catalog.FirstPartyAliases(model) {
				if alias == model {
					continue
				}
				if err := mergeMediaSupplement(raw, alias, entry); err != nil {
					return nil, nil, err
				}
			}
		}
	}
	overrides := s.LoadPricingOverrideEntries()
	for model, patch := range overrides {
		base, exists := raw[model]
		if !exists {
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(patch, &fields); err != nil || fields == nil {
				return nil, nil, fmt.Errorf("invalid pricing override: %s", model)
			}
			fields["source"] = json.RawMessage(`"local_override"`)
			raw[model], _ = json.Marshal(fields)
			continue
		}
		merged, valid := pricing.MergePricingOverrideEntry(base, patch)
		if !valid {
			return nil, nil, fmt.Errorf("invalid pricing override: %s", model)
		}
		var fields, patchFields map[string]json.RawMessage
		_ = json.Unmarshal(merged, &fields)
		_ = json.Unmarshal(patch, &patchFields)
		var sources map[string]string
		_ = json.Unmarshal(fields["price_sources"], &sources)
		if sources == nil {
			sources = map[string]string{}
		}
		for key, label := range map[string]string{
			"input_cost_per_token":                      "input",
			"output_cost_per_token":                     "output",
			"cache_read_input_token_cost":               "cache_read",
			"cache_creation_input_token_cost":           "cache_write",
			"cache_creation_input_token_cost_above_1hr": "cache_write_1h",
			"output_cost_per_image":                     "image",
			"input_cost_per_image_token":                "image_input",
			"output_cost_per_image_token":               "image_output",
		} {
			if _, exists := patchFields[key]; exists {
				sources[label] = "local_override"
			}
		}
		fields["price_sources"], _ = json.Marshal(sources)
		raw[model], _ = json.Marshal(fields)
	}
	if len(raw) == 0 {
		return catalog, map[string]*LiteLLMModelPricing{}, nil
	}
	prices, diagnostics, err := pricing.ParsePricingEntries(raw)
	if validationErr := diagnostics.ValidationError(); validationErr != nil {
		return nil, nil, validationErr
	}
	return catalog, prices, err
}

// mergeMediaSupplement 填补媒体单价及生图模型缺失的文本输出价，保留目录已有单价。
func mergeMediaSupplement(raw map[string]json.RawMessage, model string, entry json.RawMessage) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(entry, &fields); err != nil {
		return fmt.Errorf("invalid pricing supplement %s: %w", model, err)
	}
	if fields == nil {
		return fmt.Errorf("invalid pricing supplement %s: expected an object", model)
	}
	fields["source"] = json.RawMessage(`"local_supplement"`)
	base, exists := raw[model]
	if !exists {
		body, err := json.Marshal(fields)
		if err != nil {
			return err
		}
		raw[model] = body
		return nil
	}
	var combined map[string]json.RawMessage
	if err := json.Unmarshal(base, &combined); err != nil {
		return err
	}
	var sources map[string]string
	if value, exists := combined["price_sources"]; exists {
		if err := json.Unmarshal(value, &sources); err != nil {
			return fmt.Errorf("invalid price_sources for %s: %w", model, err)
		}
	}
	if sources == nil {
		sources = map[string]string{}
	}
	for key, label := range map[string]string{
		"output_cost_per_token":       "output",
		"output_cost_per_image":       "image",
		"output_cost_per_image_token": "image_output",
		"input_cost_per_image_token":  "image_input",
	} {
		if _, exists := combined[key]; exists {
			continue
		}
		if value, exists := fields[key]; exists {
			combined[key] = value
			sources[label] = "local_supplement"
		}
	}
	combined["price_sources"], _ = json.Marshal(sources)
	body, err := json.Marshal(combined)
	if err != nil {
		return err
	}
	raw[model] = body
	return nil
}

// modelsCatalogFallbackPrices 只将原厂裸名及明确的本地价格交给既有回退策略。
// 中继和供应商限定名称保留精确查价，不能参与其他模型的日期或系列回退。
// @project-doc docs/interfaces/model_catalog_and_marketplace.md#model_catalog_metadata_lookup
func modelsCatalogFallbackPrices(catalog *modelcatalog.Catalog, prices map[string]*LiteLLMModelPricing) map[string]*LiteLLMModelPricing {
	if catalog == nil {
		return nil
	}
	result := make(map[string]*LiteLLMModelPricing)
	for name, price := range prices {
		if price == nil || strings.Contains(name, "/") {
			continue
		}
		entry := catalog.Entries[strings.ToLower(strings.TrimSpace(name))]
		if entry.FirstParty || price.Source == "local_override" || price.Source == "local_supplement" {
			result[name] = price
		}
	}
	return result
}

// publishModelsCatalog 只有完整构建成功才替换价格和属性，调用期间由更新锁串行化。
func (s *PricingService) publishModelsCatalog(body []byte, updated time.Time, persist bool) error {
	var catalog *modelcatalog.Catalog
	var prices map[string]*LiteLLMModelPricing
	var fingerprint string
	// 文件编辑可能与目录同步重叠，只发布来自同一组本地文件内容的价格投影。
	for attempt := 0; attempt < 3; attempt++ {
		before := s.CustomPricingFilesFingerprint()
		var err error
		catalog, prices, err = s.buildModelsCatalog(body)
		if err != nil {
			return err
		}
		fingerprint = s.CustomPricingFilesFingerprint()
		if before == fingerprint {
			break
		}
		if attempt == 2 {
			return fmt.Errorf("custom pricing files changed during catalog update")
		}
	}
	if persist {
		path := s.GetPricingFilePath()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		temp, err := os.CreateTemp(filepath.Dir(path), "models-catalog-*.json")
		if err != nil {
			return err
		}
		name := temp.Name()
		defer func() { _ = os.Remove(name) }()
		_, writeErr := temp.Write(body)
		closeErr := temp.Close()
		if writeErr != nil {
			return writeErr
		}
		if closeErr != nil {
			return closeErr
		}
		if err := os.Rename(name, path); err != nil {
			return err
		}
	}
	fallbackPrices := modelsCatalogFallbackPrices(catalog, prices)
	s.mu.Lock()
	s.modelCatalog, s.pricingData = catalog, prices
	s.fallbackPricingData = fallbackPrices
	s.catalogBody = append([]byte(nil), body...)
	s.lastUpdated, s.localHash, s.lastCatalogError = updated, catalog.Version, ""
	s.customFilesHash = fingerprint
	s.mu.Unlock()
	return nil
}

func (s *PricingService) loadModelsCatalog() (err error) {
	s.updateMu.Lock()
	defer s.updateMu.Unlock()
	defer func() { s.recordCatalogError(err) }()
	body, readErr := os.ReadFile(s.GetPricingFilePath())
	updated := time.Now()
	s.mu.RLock()
	hasCurrent := s.modelCatalog != nil
	if readErr != nil && len(s.catalogBody) > 0 {
		body = append([]byte(nil), s.catalogBody...)
		readErr = nil
	}
	s.mu.RUnlock()
	if info, err := os.Stat(s.GetPricingFilePath()); err == nil {
		updated = info.ModTime()
	}
	var candidateErr error
	if readErr == nil {
		candidateErr = s.publishModelsCatalog(body, updated, false)
		if candidateErr == nil {
			return nil
		}
		if hasCurrent {
			return candidateErr
		}
	}
	cachedBody := body
	body, err = modelcatalog.Offline()
	if err != nil {
		return err
	}
	candidateErr = s.publishModelsCatalog(body, time.Now(), true)
	if candidateErr == nil {
		return nil
	}
	// 缓存目录不可写时仍发布内存目录，不把落盘能力作为查询前提。
	if err := s.publishModelsCatalog(body, time.Now(), false); err == nil {
		return nil
	}
	if hasCurrent {
		return candidateErr
	}
	// 首次启动时损坏的本地补充不能使整个目录消失；保留错误并等待修复后热重载。
	bootstrap := NewPricingService(Options{ModelsDev: true, DataDir: s.options.DataDir}, nil)
	for _, candidate := range [][]byte{cachedBody, body} {
		if len(candidate) == 0 {
			continue
		}
		if err := bootstrap.publishModelsCatalog(candidate, updated, true); err != nil {
			if err := bootstrap.publishModelsCatalog(candidate, updated, false); err != nil {
				continue
			}
		}
		s.mu.Lock()
		s.modelCatalog, s.pricingData = bootstrap.modelCatalog, bootstrap.pricingData
		s.fallbackPricingData = bootstrap.fallbackPricingData
		s.catalogBody = bootstrap.catalogBody
		s.lastUpdated, s.localHash = bootstrap.lastUpdated, bootstrap.localHash
		s.lastCatalogError = candidateErr.Error()
		s.mu.Unlock()
		return nil
	}
	return candidateErr
}

// recordCatalogError 统一记录远程更新和本地重载失败，供管理页查询。
func (s *PricingService) recordCatalogError(err error) {
	if err == nil {
		return
	}
	s.mu.Lock()
	s.lastCatalogError = err.Error()
	s.mu.Unlock()
}

// updateModelsCatalog 使用条件请求；网络或解析失败保留整个旧版本。
func (s *PricingService) updateModelsCatalog(force bool) (err error) {
	// 本地重载自己持有更新锁，并可回退到内存中的离线目录。
	if strings.TrimSpace(s.options.RemoteURL) == "" {
		return s.loadModelsCatalog()
	}
	s.updateMu.Lock()
	defer s.updateMu.Unlock()
	defer func() { s.recordCatalogError(err) }()
	url, err := s.ValidatePricingURL(s.options.RemoteURL)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var body []byte
	var etag string
	if remote, ok := s.remoteClient.(interface {
		FetchCatalog(context.Context, string, string) ([]byte, string, bool, error)
	}); ok {
		validator := s.catalogETag
		if force {
			validator = ""
		}
		var unchanged bool
		body, etag, unchanged, err = remote.FetchCatalog(ctx, url, validator)
		if err != nil {
			return err
		}
		if unchanged {
			s.mu.Lock()
			s.lastCatalogError = ""
			s.mu.Unlock()
			return nil
		}
	} else {
		body, err = s.remoteClient.FetchPricingJSON(ctx, url)
	}
	if err != nil {
		return err
	}
	hash := sha256.Sum256(body)
	s.mu.RLock()
	same := s.localHash == hex.EncodeToString(hash[:]) && s.customFilesHash == s.CustomPricingFilesFingerprint()
	s.mu.RUnlock()
	if same && !force {
		s.catalogETag = etag
		s.mu.Lock()
		s.lastCatalogError = ""
		s.mu.Unlock()
		return nil
	}
	if err = s.publishModelsCatalog(body, time.Now(), true); err != nil {
		return err
	}
	s.catalogETag = etag
	return nil
}
