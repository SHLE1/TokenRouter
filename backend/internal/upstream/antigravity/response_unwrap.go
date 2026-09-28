package antigravity

import (
	"github.com/tidwall/gjson"
) // UnwrapV1InternalResponse 解包 v1internal 响应

// UnwrapV1InternalResponse 使用 gjson 零拷贝提取 response 字段，避免 Unmarshal+Marshal 双重开销
func (s *ResponseAdapter) UnwrapV1InternalResponse(body []byte) ([]byte, error) {
	result := gjson.GetBytes(body, "response")
	if result.Exists() {
		return []byte(result.Raw), nil
	}
	return body, nil
}
