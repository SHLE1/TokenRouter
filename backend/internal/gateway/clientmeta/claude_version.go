package clientmeta

import (
	"strings"

	"golang.org/x/mod/semver"
)

// IsSupportedClaudeCLIVersion 校验运维设置的版本覆盖值。
// 版本需要为三段纯数字（如 2.1.251），并且不低于内置基线 CLICurrentVersion。
// -local、-dev、+build 等后缀会被 fingerprintUserAgentPattern 拒绝，写入持久指纹后可能持续触发上游 429，系统没有指纹重置入口。
// 降低版本基线也会降低身份服务检查客户端主版本超前时使用的基准。
func IsSupportedClaudeCLIVersion(version, minimum string) bool {
	version = strings.TrimSpace(version)
	if version == "" {
		return false
	}
	// Canonical 相等检查排除 v1.2 等省略段版本，再分别检查预发布和构建元数据。
	canonical := "v" + version
	if !semver.IsValid(canonical) || semver.Canonical(canonical) != canonical {
		return false
	}
	if semver.Prerelease(canonical) != "" || semver.Build(canonical) != "" {
		return false
	}
	return semver.Compare(canonical, "v"+minimum) >= 0
}
