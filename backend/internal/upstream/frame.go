package upstream

import "context"

type FrameKind int

const (
	FrameText   FrameKind = 1
	FrameBinary FrameKind = 2
)

type FrameConn interface {
	ReadFrame(context.Context) (FrameKind, []byte, error)
	WriteFrame(context.Context, FrameKind, []byte) error
	Close() error
}
