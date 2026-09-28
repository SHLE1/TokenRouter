package antigravity

import (
	"net/http"

	"github.com/tidwall/gjson" // IsAntigravityInternalServerError 检测特定的 INTERNAL 500 错误
	// 必须同时匹配 error.code==500, error.message=="Internal error encountered.", error.status=="INTERNAL"
)

func IsAntigravityInternalServerError(statusCode int, body []byte) bool {
	if statusCode != http.StatusInternalServerError {
		return false
	}
	return gjson.GetBytes(body, "error.code").Int() == 500 &&
		gjson.GetBytes(body, "error.message").String() == "Internal error encountered." &&
		gjson.GetBytes(body, "error.status").String() == "INTERNAL"
}
