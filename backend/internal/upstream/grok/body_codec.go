package grok

type BodyCodec struct{ NewID func() string }

const (
	ComposerImageBridgeVisionModel     = "grok-build-0.1"
	ComposerImageBridgeMaxOutputTokens = 512
)
