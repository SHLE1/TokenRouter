package pricing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// MergeConflict 保留冲突模型和原条目编号，供迁移预检定位需要人工处理的价卡。
type MergeConflict struct {
	Models   []string `json:"models"`
	EntryIDs []int64  `json:"entry_ids"`
	Reason   string   `json:"reason"`
}

// MergePriceCards 合并同一作用域的历史价卡；不能比较的配置返回完整冲突清单。
// 单价逐桶取高，空值保留继承含义；不同作用域的提供商成本规则必须分别调用。
func MergePriceCards(entries []ModelPricingEntry) ([]ModelPricingEntry, []MergeConflict) {
	type mergedEntry struct {
		card ModelPricingEntry
		ids  []int64
	}
	byModel := make(map[string]*mergedEntry)
	var conflicts []MergeConflict
	for _, entry := range entries {
		for _, model := range entry.Models {
			model = NormalizePriceModelName(model)
			if model == "" || strings.Contains(strings.TrimSuffix(model, "*"), "*") {
				conflicts = append(conflicts, MergeConflict{Models: []string{model}, EntryIDs: []int64{entry.ID}, Reason: "invalid model pattern"})
				continue
			}
			card := entry.Clone()
			card.Models = []string{model}
			if card.BillingMode == "" {
				card.BillingMode = BillingModeToken
			}
			previous := byModel[model]
			if previous == nil {
				byModel[model] = &mergedEntry{card: card, ids: []int64{entry.ID}}
				continue
			}
			previous.ids = append(previous.ids, entry.ID)
			merged, reason := mergePriceCard(previous.card, card)
			if reason != "" {
				conflicts = append(conflicts, MergeConflict{Models: []string{model}, EntryIDs: append([]int64(nil), previous.ids...), Reason: reason})
				continue
			}
			previous.card = merged
		}
	}
	models := make([]string, 0, len(byModel))
	for model := range byModel {
		models = append(models, model)
	}
	sort.Strings(models)
	for i, left := range models {
		for _, right := range models[i+1:] {
			if modelPatternsOverlap(left, right) {
				ids := append([]int64(nil), byModel[left].ids...)
				ids = append(ids, byModel[right].ids...)
				conflicts = append(conflicts, MergeConflict{Models: []string{left, right}, EntryIDs: ids, Reason: "overlapping model patterns"})
			}
		}
	}
	if len(conflicts) > 0 {
		return nil, conflicts
	}
	out := make([]ModelPricingEntry, 0, len(models))
	for _, model := range models {
		out = append(out, byModel[model].card)
	}
	return out, nil
}

// modelPatternsOverlap 检查精确名称和末尾通配符的交集，不扩展未知模型目录。
func modelPatternsOverlap(left, right string) bool {
	return (strings.HasSuffix(left, "*") && strings.HasPrefix(strings.TrimSuffix(right, "*"), strings.TrimSuffix(left, "*"))) ||
		(strings.HasSuffix(right, "*") && strings.HasPrefix(strings.TrimSuffix(left, "*"), strings.TrimSuffix(right, "*")))
}

// mergePriceCard 先核对所有非价格规则，再对相同计量单位的显式单价取高。
func mergePriceCard(left, right ModelPricingEntry) (ModelPricingEntry, string) {
	l, err := comparablePriceRules(left)
	if err != nil {
		return ModelPricingEntry{}, err.Error()
	}
	r, err := comparablePriceRules(right)
	if err != nil {
		return ModelPricingEntry{}, err.Error()
	}
	if !bytes.Equal(l, r) {
		return ModelPricingEntry{}, "billing mode, intervals, multipliers or time rules differ"
	}
	merged := left.Clone()
	destination := cardUnitPrices(&merged)
	source := cardUnitPrices(&right)
	for i := range destination {
		a, b := *destination[i], *source[i]
		if (a == nil) != (b == nil) {
			return ModelPricingEntry{}, "inherited and explicit prices cannot be compared"
		}
		if a != nil {
			if math.IsNaN(*a) || math.IsNaN(*b) || math.IsInf(*a, 0) || math.IsInf(*b, 0) || *a < 0 || *b < 0 {
				return ModelPricingEntry{}, "invalid unit price"
			}
			value := math.Max(*a, *b)
			*destination[i] = &value
		}
	}
	return merged, ""
}

// cardUnitPrices 只枚举金额桶；倍率、区间边界及排序不参加逐项取高。
func cardUnitPrices(card *ModelPricingEntry) []**float64 {
	prices := []**float64{&card.InputPrice, &card.OutputPrice, &card.CacheWritePrice, &card.CacheWrite1hPrice, &card.CacheReadPrice, &card.ImageInputPrice, &card.ImageOutputPrice, &card.PerRequestPrice}
	for i := range card.Intervals {
		interval := &card.Intervals[i]
		prices = append(prices, &interval.InputPrice, &interval.OutputPrice, &interval.CacheWritePrice, &interval.CacheWrite1hPrice, &interval.CacheReadPrice, &interval.PerRequestPrice)
	}
	return prices
}

// comparablePriceRules 去掉存储身份及金额，仅比较会改变计费口径的规则。
func comparablePriceRules(card ModelPricingEntry) ([]byte, error) {
	card = card.Clone()
	card.ID, card.PricingConfigID = 0, 0
	card.Models = nil
	card.CreatedAt, card.UpdatedAt = time.Time{}, time.Time{}
	for i := range card.Intervals {
		interval := &card.Intervals[i]
		interval.ID, interval.PricingID = 0, 0
		interval.CreatedAt, interval.UpdatedAt = time.Time{}, time.Time{}
	}
	for _, price := range cardUnitPrices(&card) {
		*price = nil
	}
	data, err := json.Marshal(card)
	if err != nil {
		return nil, fmt.Errorf("invalid pricing rules: %w", err)
	}
	return data, nil
}
