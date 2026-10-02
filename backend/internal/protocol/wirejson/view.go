package wirejson

import (
	"unsafe"

	"github.com/tidwall/gjson"
)

func ParseView(raw []byte) gjson.Result {
	if len(raw) == 0 {
		return gjson.Result{}
	}
	// 同步解析直接引用 raw，适用于较大的 messages 和 contents。
	return gjson.Parse(*(*string)(unsafe.Pointer(&raw)))
}
