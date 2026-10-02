package clientmeta_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/TokenFlux/TokenRouter/internal/gateway/clientmeta"
	"github.com/TokenFlux/TokenRouter/internal/gateway/requeststate"
	"github.com/stretchr/testify/require"
)

const claudeCodeMetadataUserIDJSON = `{"device_id":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef","account_uuid":"","session_id":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"}`

func TestClaudeCodeValidator_ProbeBypass(t *testing.T) {
	validator := clientmeta.NewClaudeCodeValidator()
	req := httptest.NewRequest(http.MethodPost, "http://example.com/v1/messages", nil)
	req.Header.Set("User-Agent", "claude-cli/1.2.3 (darwin; arm64)")
	req = req.WithContext(requeststate.WithIsMaxTokensOneHaikuRequest(req.Context(), true))

	ok := validator.Validate(claudeCodeInputFixture(req), map[string]any{
		"model":      "claude-haiku-4-5",
		"max_tokens": 1,
	})
	require.True(t, ok)
}

func TestClaudeCodeValidator_ProbeBypassRequiresUA(t *testing.T) {
	validator := clientmeta.NewClaudeCodeValidator()
	req := httptest.NewRequest(http.MethodPost, "http://example.com/v1/messages", nil)
	req.Header.Set("User-Agent", "curl/8.0.0")
	req = req.WithContext(requeststate.WithIsMaxTokensOneHaikuRequest(req.Context(), true))

	ok := validator.Validate(claudeCodeInputFixture(req), map[string]any{
		"model":      "claude-haiku-4-5",
		"max_tokens": 1,
	})
	require.False(t, ok)
}

func TestClaudeCodeValidator_MessagesWithoutProbeStillNeedStrictValidation(t *testing.T) {
	validator := clientmeta.NewClaudeCodeValidator()
	req := httptest.NewRequest(http.MethodPost, "http://example.com/v1/messages", nil)
	req.Header.Set("User-Agent", "claude-cli/1.2.3 (darwin; arm64)")

	ok := validator.Validate(claudeCodeInputFixture(req), map[string]any{
		"model":      "claude-haiku-4-5",
		"max_tokens": 1,
	})
	require.False(t, ok)
}

func TestClaudeCodeValidator_CountTokensPathUAOnly(t *testing.T) {
	validator := clientmeta.NewClaudeCodeValidator()
	req := httptest.NewRequest(http.MethodPost, "http://example.com/v1/messages/count_tokens", nil)
	req.Header.Set("User-Agent", "claude-cli/2.1.156 (Claude Code)")

	ok := validator.Validate(claudeCodeInputFixture(req), map[string]any{
		"model": "claude-opus-4-8",
	})
	require.True(t, ok)
}

func TestClaudeCodeValidator_CountTokensPathRequiresUA(t *testing.T) {
	validator := clientmeta.NewClaudeCodeValidator()
	req := httptest.NewRequest(http.MethodPost, "http://example.com/v1/messages/count_tokens", nil)
	req.Header.Set("User-Agent", "curl/8.0.0")

	ok := validator.Validate(claudeCodeInputFixture(req), map[string]any{
		"model": "claude-opus-4-8",
	})
	require.False(t, ok)
}

func TestClaudeCodeValidator_MessagesPathFullValid(t *testing.T) {
	validator := clientmeta.NewClaudeCodeValidator()
	req := httptest.NewRequest(http.MethodPost, "http://example.com/v1/messages", nil)
	req.Header.Set("User-Agent", "claude-cli/2.1.156 (Claude Code)")
	req.Header.Set("X-App", "claude-code")
	req.Header.Set("anthropic-beta", "claude-code-20250219")
	req.Header.Set("anthropic-version", "2023-06-01")

	ok := validator.Validate(claudeCodeInputFixture(req), map[string]any{
		"model":  "claude-opus-4-8",
		"stream": true,
		"system": []any{
			map[string]any{
				"type": "text",
				"text": "You are Claude Code, Anthropic's official CLI for Claude.",
			},
		},
		"metadata": map[string]any{
			"user_id": "user_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa_account__session_aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		},
	})
	require.True(t, ok)
}

func TestClaudeCodeValidator_BillingBlockRecognizedWithoutIdentityPrompt(t *testing.T) {
	// 使用抓取的安全监视器系统提示词，内容没有身份说明文本。
	monitorPrompt, err := os.ReadFile("testdata/security_monitor_system_prompt.txt")
	require.NoError(t, err)

	validator := clientmeta.NewClaudeCodeValidator()

	// 监视器正文与身份说明文本的 Dice 相似度低于阈值，此例由计费归因块识别放行。
	require.Less(t, validator.BestSimilarityScore(string(monitorPrompt)), clientmeta.ClaudeCodeSystemPromptThreshold)

	req := httptest.NewRequest(http.MethodPost, "http://example.com/v1/messages", nil)
	req.Header.Set("User-Agent", "claude-cli/2.1.162 (external, cli)")
	req.Header.Set("X-App", "cli")
	req.Header.Set("anthropic-beta", "claude-code-20250219")
	req.Header.Set("anthropic-version", "2023-06-01")

	// Claude Code 安全监视器子请求：不携带身份说明文本，但 system 数组携带计费归因块
	// cc_entrypoint=cli，应据此识别为 Claude Code 客户端。
	ok := validator.Validate(claudeCodeInputFixture(req), map[string]any{
		"model": "claude-3-5-haiku-20241022",
		"system": []any{
			map[string]any{
				"type": "text",
				"text": "x-anthropic-billing-header: cc_version=2.1.162.884; cc_entrypoint=cli; cch=d8726;",
			},
			map[string]any{
				"type": "text",
				"text": string(monitorPrompt),
			},
		},
		"metadata": map[string]any{
			"user_id": claudeCodeMetadataUserIDJSON,
		},
	})
	require.True(t, ok)
}

func TestClaudeCodeValidator_SecurityMonitorWithoutBillingBlock(t *testing.T) {
	monitorPrompt, err := os.ReadFile("testdata/security_monitor_system_prompt.txt")
	require.NoError(t, err)

	validHeaders := map[string]string{
		"User-Agent":        "claude-cli/2.1.220 (external, cli)",
		"X-App":             "cli",
		"anthropic-beta":    "claude-code-20250219",
		"anthropic-version": "2023-06-01",
	}
	validBody := func(prompt string) map[string]any {
		return map[string]any{
			"model": "claude-haiku-4-5-20251001",
			"system": []any{
				map[string]any{"type": "text", "text": prompt},
			},
			"metadata": map[string]any{"user_id": claudeCodeMetadataUserIDJSON},
		}
	}
	// CLI 会在安全监视器提示词之外追加随会话变化的上下文块。
	sessionContext := "\n\n## Session Context\n\n- **User identity**: testuser\n" +
		"- **Working directory**: /home/testuser/project\n- **Platform**: linux"

	tests := []struct {
		name       string
		headers    map[string]string
		body       map[string]any
		wantAccept bool
	}{
		{
			name:       "official classifier request",
			headers:    validHeaders,
			body:       validBody(string(monitorPrompt)),
			wantAccept: true,
		},
		{
			name: "non-Claude user agent",
			headers: map[string]string{
				"User-Agent":        "curl/8.0.0",
				"X-App":             "cli",
				"anthropic-beta":    "claude-code-20250219",
				"anthropic-version": "2023-06-01",
			},
			body: validBody(string(monitorPrompt)),
		},
		{
			name: "missing X-App",
			headers: map[string]string{
				"User-Agent":        validHeaders["User-Agent"],
				"anthropic-beta":    validHeaders["anthropic-beta"],
				"anthropic-version": validHeaders["anthropic-version"],
			},
			body: validBody(string(monitorPrompt)),
		},
		{
			name: "missing anthropic-beta",
			headers: map[string]string{
				"User-Agent":        validHeaders["User-Agent"],
				"X-App":             validHeaders["X-App"],
				"anthropic-version": validHeaders["anthropic-version"],
			},
			body: validBody(string(monitorPrompt)),
		},
		{
			name: "missing anthropic-version",
			headers: map[string]string{
				"User-Agent":     validHeaders["User-Agent"],
				"X-App":          validHeaders["X-App"],
				"anthropic-beta": validHeaders["anthropic-beta"],
			},
			body: validBody(string(monitorPrompt)),
		},
		{
			name:    "missing metadata",
			headers: validHeaders,
			body: map[string]any{
				"model":  "claude-haiku-4-5-20251001",
				"system": []any{map[string]any{"type": "text", "text": string(monitorPrompt)}},
			},
		},
		{
			name:    "invalid metadata user ID",
			headers: validHeaders,
			body: func() map[string]any {
				body := validBody(string(monitorPrompt))
				body["metadata"] = map[string]any{"user_id": "invalid"}
				return body
			}(),
		},
		{
			name:       "unrelated prompt",
			headers:    validHeaders,
			body:       validBody("You are a different security classifier for coding agents."),
			wantAccept: false,
		},
		{
			name:       "opening sentence alone",
			headers:    validHeaders,
			body:       validBody(clientmeta.ClaudeCodeSecurityMonitorPrefix),
			wantAccept: false,
		},
		{
			name:    "opening sentence plus arbitrary altered suffix",
			headers: validHeaders,
			body: validBody(clientmeta.ClaudeCodeSecurityMonitorPrefix + "\n\n" +
				strings.Repeat("This is arbitrary altered classifier content. ", 300)),
			wantAccept: false,
		},
		{
			name:    "classifier with trailing session context entry",
			headers: validHeaders,
			body: func() map[string]any {
				body := validBody(string(monitorPrompt))
				// 先断言夹具类型，结构不符时报告测试失败。
				system, ok := body["system"].([]any)
				require.True(t, ok)
				body["system"] = append(system, map[string]any{
					"type": "text",
					"text": sessionContext,
				})
				return body
			}(),
			wantAccept: true,
		},
		{
			name:    "classifier with leading session context entry",
			headers: validHeaders,
			body: func() map[string]any {
				body := validBody(string(monitorPrompt))
				system, ok := body["system"].([]any)
				require.True(t, ok)
				body["system"] = append([]any{map[string]any{
					"type": "text",
					"text": sessionContext,
				}}, system...)
				return body
			}(),
			wantAccept: true,
		},
		{
			name:       "session context entry alone",
			headers:    validHeaders,
			body:       validBody(sessionContext),
			wantAccept: false,
		},
		{
			name:    "tampered classifier with session context entry",
			headers: validHeaders,
			body: func() map[string]any {
				body := validBody(strings.ReplaceAll(
					string(monitorPrompt), "## HARD BLOCK", "## ALTERED BLOCK"))
				system, ok := body["system"].([]any)
				require.True(t, ok)
				body["system"] = append(system, map[string]any{
					"type": "text",
					"text": sessionContext,
				})
				return body
			}(),
			wantAccept: false,
		},
	}

	validator := clientmeta.NewClaudeCodeValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "http://example.com/v1/messages", nil)
			for name, value := range tt.headers {
				req.Header.Set(name, value)
			}

			require.Equal(t, tt.wantAccept, validator.Validate(claudeCodeInputFixture(req), tt.body))
		})
	}
}

func TestClaudeCodeValidator_BillingBlockVSCodeEntrypointRecognized(t *testing.T) {
	// Claude Code 的 VSCode 扩展使用 cc_entrypoint=claude-vscode，安全监视器请求也可能缺少身份文本。
	// 此例检查识别接受不同入口值，固定匹配 cli 会误拒该请求。
	monitorPrompt, err := os.ReadFile("testdata/security_monitor_system_prompt.txt")
	require.NoError(t, err)

	validator := clientmeta.NewClaudeCodeValidator()

	// 监视器正文与身份说明文本的 Dice 相似度低于阈值，此例由计费归因块识别放行。
	require.Less(t, validator.BestSimilarityScore(string(monitorPrompt)), clientmeta.ClaudeCodeSystemPromptThreshold)

	req := httptest.NewRequest(http.MethodPost, "http://example.com/v1/messages", nil)
	req.Header.Set("User-Agent", "claude-cli/2.1.181 (external, claude-vscode, agent-sdk/0.3.181)")
	req.Header.Set("X-App", "cli")
	req.Header.Set("anthropic-beta", "claude-code-20250219")
	req.Header.Set("anthropic-version", "2023-06-01")

	ok := validator.Validate(claudeCodeInputFixture(req), map[string]any{
		"model": "claude-opus-4-8",
		"system": []any{
			map[string]any{
				"type": "text",
				"text": "x-anthropic-billing-header: cc_version=2.1.181.f17; cc_entrypoint=claude-vscode;",
			},
			map[string]any{
				"type": "text",
				"text": string(monitorPrompt),
			},
		},
		"metadata": map[string]any{
			"user_id": claudeCodeMetadataUserIDJSON,
		},
	})
	require.True(t, ok)
}

func TestClaudeCodeValidator_BillingBlockWithoutEntrypointFallsThrough(t *testing.T) {
	validator := clientmeta.NewClaudeCodeValidator()
	req := httptest.NewRequest(http.MethodPost, "http://example.com/v1/messages", nil)
	req.Header.Set("User-Agent", "claude-cli/2.1.162 (external, cli)")
	req.Header.Set("X-App", "cli")
	req.Header.Set("anthropic-beta", "claude-code-20250219")
	req.Header.Set("anthropic-version", "2023-06-01")

	// 计费块前缀命中但完全没有 cc_entrypoint= 字段，且无身份 prose：
	// 不应凭前缀放行，应落回 Dice 检查并失败。验证 cc_entrypoint= 字段的存在仍是必要条件。
	ok := validator.Validate(claudeCodeInputFixture(req), map[string]any{
		"model": "claude-3-5-haiku-20241022",
		"system": []any{
			map[string]any{
				"type": "text",
				"text": "x-anthropic-billing-header: cc_version=2.1.162.884; cch=d8726;",
			},
			map[string]any{
				"type": "text",
				"text": "Some unrelated system prompt that does not resemble Claude Code.",
			},
		},
		"metadata": map[string]any{
			"user_id": claudeCodeMetadataUserIDJSON,
		},
	})
	require.False(t, ok)
}

func TestClaudeCodeValidator_BillingBlockStillRequiresClaudeCodeUA(t *testing.T) {
	validator := clientmeta.NewClaudeCodeValidator()
	req := httptest.NewRequest(http.MethodPost, "http://example.com/v1/messages", nil)
	req.Header.Set("User-Agent", "curl/8.0.0")
	req.Header.Set("X-App", "cli")
	req.Header.Set("anthropic-beta", "claude-code-20250219")
	req.Header.Set("anthropic-version", "2023-06-01")

	// 计费块无法绕过 User-Agent 校验：非 claude-cli 客户端在第一步即被拒。
	ok := validator.Validate(claudeCodeInputFixture(req), map[string]any{
		"model": "claude-3-5-haiku-20241022",
		"system": []any{
			map[string]any{
				"type": "text",
				"text": "x-anthropic-billing-header: cc_version=2.1.162.884; cc_entrypoint=cli; cch=d8726;",
			},
		},
	})
	require.False(t, ok)
}

// TestClaudeCodeValidator_BillingBlockRecognizedWithoutCCH 验证新版 Claude Code CLI 已取消 cch=... 签名字段，billing block 形如
// `x-anthropic-billing-header: cc_version=...; cc_entrypoint=cli;`（无 cch）。
// 检测使用计费前缀和 cc_entrypoint 字段，缺少 cch 和身份文本的子请求仍可识别。
// 本测试同时覆盖 buildBillingAttributionText 注入的计费块格式。
func TestClaudeCodeValidator_BillingBlockRecognizedWithoutCCH(t *testing.T) {
	monitorPrompt, err := os.ReadFile("testdata/security_monitor_system_prompt.txt")
	require.NoError(t, err)

	validator := clientmeta.NewClaudeCodeValidator()
	require.Less(t, validator.BestSimilarityScore(string(monitorPrompt)), clientmeta.ClaudeCodeSystemPromptThreshold)

	req := httptest.NewRequest(http.MethodPost, "http://example.com/v1/messages", nil)
	req.Header.Set("User-Agent", "claude-cli/2.1.162 (external, cli)")
	req.Header.Set("X-App", "cli")
	req.Header.Set("anthropic-beta", "claude-code-20250219")
	req.Header.Set("anthropic-version", "2023-06-01")

	ok := validator.Validate(claudeCodeInputFixture(req), map[string]any{
		"model": "claude-3-5-haiku-20241022",
		"system": []any{
			map[string]any{
				"type": "text",
				// 注意：无 cch 段，对齐新版 CLI 与本仓新的注入格式。
				"text": "x-anthropic-billing-header: cc_version=2.1.162.884; cc_entrypoint=cli;",
			},
			map[string]any{
				"type": "text",
				"text": string(monitorPrompt),
			},
		},
		"metadata": map[string]any{
			"user_id": claudeCodeMetadataUserIDJSON,
		},
	})
	require.True(t, ok, "无 cch 的新版 billing block 仍应被识别为 Claude Code")
}

// TestClaudeCodeValidator_NoCCHBlockStillRequiresClaudeCodeUA 检查缺少 cch 的请求仍通过 claude-cli UA 校验。
// 其他 UA 即使携带 billing block，也在第一步被拒绝。
func TestClaudeCodeValidator_NoCCHBlockStillRequiresClaudeCodeUA(t *testing.T) {
	validator := clientmeta.NewClaudeCodeValidator()
	req := httptest.NewRequest(http.MethodPost, "http://example.com/v1/messages", nil)
	req.Header.Set("User-Agent", "curl/8.0.0")
	req.Header.Set("X-App", "cli")
	req.Header.Set("anthropic-beta", "claude-code-20250219")
	req.Header.Set("anthropic-version", "2023-06-01")

	ok := validator.Validate(claudeCodeInputFixture(req), map[string]any{
		"model": "claude-3-5-haiku-20241022",
		"system": []any{
			map[string]any{
				"type": "text",
				"text": "x-anthropic-billing-header: cc_version=2.1.162.884; cc_entrypoint=cli;",
			},
		},
	})
	require.False(t, ok)
}

func TestClaudeCodeValidator_MessagesPathRejectsNonClaudeCodeUA(t *testing.T) {
	validator := clientmeta.NewClaudeCodeValidator()
	req := httptest.NewRequest(http.MethodPost, "http://example.com/v1/messages", nil)
	req.Header.Set("User-Agent", "curl/8.0.0")
	req.Header.Set("X-App", "claude-code")
	req.Header.Set("anthropic-beta", "claude-code-20250219")
	req.Header.Set("anthropic-version", "2023-06-01")

	ok := validator.Validate(claudeCodeInputFixture(req), map[string]any{
		"model":  "claude-opus-4-8",
		"stream": true,
		"system": []any{
			map[string]any{
				"type": "text",
				"text": "You are Claude Code, Anthropic's official CLI for Claude.",
			},
		},
		"metadata": map[string]any{
			"user_id": "user_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa_account__session_aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		},
	})
	require.False(t, ok)
}

func TestClaudeCodeValidator_MessagesPathWithoutSystemPromptStillRejected(t *testing.T) {
	validator := clientmeta.NewClaudeCodeValidator()
	req := httptest.NewRequest(http.MethodPost, "http://example.com/v1/messages", nil)
	req.Header.Set("User-Agent", "claude-cli/2.1.156 (Claude Code)")
	req.Header.Set("X-App", "claude-code")
	req.Header.Set("anthropic-beta", "claude-code-20250219")
	req.Header.Set("anthropic-version", "2023-06-01")

	ok := validator.Validate(claudeCodeInputFixture(req), map[string]any{
		"model":  "claude-opus-4-8",
		"stream": true,
		"messages": []any{
			map[string]any{"role": "user", "content": "hello"},
		},
		"metadata": map[string]any{
			"user_id": "user_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa_account__session_aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		},
	})
	require.False(t, ok)
}

func TestClaudeCodeValidator_NonMessagesPathUAOnly(t *testing.T) {
	validator := clientmeta.NewClaudeCodeValidator()
	req := httptest.NewRequest(http.MethodPost, "http://example.com/v1/models", nil)
	req.Header.Set("User-Agent", "claude-cli/1.2.3 (darwin; arm64)")

	ok := validator.Validate(claudeCodeInputFixture(req), nil)
	require.True(t, ok)
}

func TestExtractVersion(t *testing.T) {
	v := clientmeta.NewClaudeCodeValidator()
	tests := []struct {
		ua   string
		want string
	}{
		{"claude-cli/2.1.22 (darwin; arm64)", "2.1.22"},
		{"claude-cli/1.0.0", "1.0.0"},
		{"Claude-CLI/3.10.5 (linux; x86_64)", "3.10.5"}, // 大小写不敏感
		{"curl/8.0.0", ""},                              // 非 Claude CLI
		{"", ""},                                        // 空字符串
		{"claude-cli/", ""},                             // 无版本号
		{"claude-cli/2.1.22-beta", "2.1.22"},            // 带后缀仍提取主版本号
	}
	for _, tt := range tests {
		got := v.ExtractVersion(tt.ua)
		require.Equal(t, tt.want, got, "ExtractVersion(%q)", tt.ua)
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"2.1.0", "2.1.0", 0},   // 相等
		{"2.1.1", "2.1.0", 1},   // patch 更大
		{"2.0.0", "2.1.0", -1},  // minor 更小
		{"3.0.0", "2.99.99", 1}, // major 更大
		{"1.0.0", "2.0.0", -1},  // major 更小
		{"0.0.1", "0.0.0", 1},   // patch 差异
		{"", "1.0.0", -1},       // 空字符串 vs 正常版本
		{"v2.1.0", "2.1.0", 0},  // v 前缀处理
	}
	for _, tt := range tests {
		got := clientmeta.CompareVersions(tt.a, tt.b)
		require.Equal(t, tt.want, got, "CompareVersions(%q, %q)", tt.a, tt.b)
	}
}

// claudeCodeInputFixture 将 HTTP 测试数据转换为客户端识别输入。
func claudeCodeInputFixture(r *http.Request) clientmeta.ClaudeCodeValidationInput {
	bypass, _ := requeststate.IsMaxTokensOneHaikuRequestFromContext(r.Context())
	return clientmeta.ClaudeCodeValidationInput{Path: r.URL.Path, UserAgent: r.Header.Get("User-Agent"), XApp: r.Header.Get("X-App"), AnthropicBeta: r.Header.Get("anthropic-beta"), AnthropicVersion: r.Header.Get("anthropic-version"), MaxTokensOneHaiku: bypass}
}
