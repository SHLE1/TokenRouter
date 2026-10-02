package completion

import (
	"context"
	"strings"
)

// RecordCyber 对符合条件的错误补记用量，调用 RecordOpenAI 应用计费或零费用规则。
// 调用方在提交前复制 Input。
func (s *Recorder) RecordCyber(ctx context.Context, in *Input) {
	if s == nil || in == nil || in.APIKey == nil || in.User == nil || in.Provider == nil || in.Result == nil || strings.TrimSpace(in.Result.Model) == "" {
		return
	}
	snapshot := Snapshot(in)
	snapshot.Result.Model = strings.TrimSpace(snapshot.Result.Model)
	snapshot.CyberBlocked = true
	if err := s.Record(ctx, snapshot, true); err != nil {
		s.printf("service.openai_gateway", "cyber usage record failed: request_id=%s err=%v", snapshot.Result.RequestID, err)
	}
}
