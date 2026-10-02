package provider

import (
	"github.com/TokenFlux/TokenRouter/internal/upstream/grok"
	"github.com/google/uuid"
)

// GrokBodyCodec 在构造请求体时生成随机 ID。
func GrokBodyCodec() grok.BodyCodec { return grok.BodyCodec{NewID: uuid.NewString} }
