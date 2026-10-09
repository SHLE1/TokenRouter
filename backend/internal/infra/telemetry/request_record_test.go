package telemetry

import (
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
)

// TestNormalizeRequestRecordRemovesUnstorableText 覆盖路径空字符、无效 UTF-8 和多字节截断。
func TestNormalizeRequestRecordRemovesUnstorableText(t *testing.T) {
	input := RequestRecord{
		RequestID: "local-id", Path: "/invalid\x00path", Model: "模型\x00" + strings.Repeat("模", 100),
		Method: "GE\x00T", Platform: "open\xffai\x00", ErrorCode: "http\x00_404",
		Aliases:  []RequestAlias{{Kind: "call\x00er", Value: "valid"}, {Kind: "caller", Value: "bad\x00id"}},
		Attempts: []RequestAttempt{{Outcome: "fail\x00ed", RequestID: "bad\x00id"}},
	}
	record := NormalizeRequestRecord(input)
	require.Equal(t, "/invalidpath", record.Path)
	require.Equal(t, "GET", record.Method)
	require.Equal(t, "openai", record.Platform)
	require.Equal(t, "http_404", record.ErrorCode)
	require.LessOrEqual(t, len(record.Model), 200)
	require.True(t, utf8.ValidString(record.Model))
	require.Equal(t, []RequestAlias{{Kind: "caller", Value: "valid"}}, record.Aliases)
	require.Equal(t, []RequestAttempt{{Outcome: "failed"}}, record.Attempts)
	body, err := json.Marshal(record)
	require.NoError(t, err)
	require.NotContains(t, string(body), `\u0000`)
	require.Equal(t, "/invalid\x00path", input.Path)
	require.Equal(t, "fail\x00ed", input.Attempts[0].Outcome)
}
