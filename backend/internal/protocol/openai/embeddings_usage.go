package openai

import "github.com/tidwall/gjson"

func ExtractEmbeddingsUsage(body []byte) ForwardUsage {
	usage := gjson.GetBytes(body, "usage")
	if !usage.Exists() || !usage.IsObject() {
		return ForwardUsage{}
	}
	inputTokens := FirstPositiveGJSONInt(
		usage.Get("prompt_tokens"),
		usage.Get("input_tokens"),
		usage.Get("total_tokens"),
	)
	outputTokens := FirstPositiveGJSONInt(
		usage.Get("completion_tokens"),
		usage.Get("output_tokens"),
	)
	cacheReadTokens := OpenAICacheReadTokensFromUsage(usage)
	cacheCreationTokens := OpenAICacheCreationTokensFromUsage(usage)
	// 多模态 embedding 会返回图文 token 拆分，用于图文不同价计费。
	imageInputTokens := FirstPositiveGJSONInt(
		usage.Get("prompt_tokens_details.image_tokens"),
		usage.Get("input_tokens_details.image_tokens"),
	)
	return ForwardUsage{
		InputTokens:              inputTokens,
		ImageInputTokens:         imageInputTokens,
		OutputTokens:             outputTokens,
		CacheReadInputTokens:     cacheReadTokens,
		CacheCreationInputTokens: cacheCreationTokens,
	}
}
