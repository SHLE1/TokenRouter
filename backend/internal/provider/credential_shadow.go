// ResolveCredentialRecord 校验凭据母提供商；只解一层，保留旧读取及错误顺序。
package provider

import (
	"context"
	"fmt"
)

// resolveCredentialProvider 解析影子提供商到其母提供商，用于凭据/Token 透传。
// - 普通提供商（非影子）：直接返回自身。
// - 影子提供商：通过 repo 取母提供商，校验母提供商存在且为 OpenAI OAuth 类型，否则返回错误。
// 凭据取得、额度查询和用量探针共同使用该解析入口，不复制母提供商校验规则。
func ResolveCredentialRecord(ctx context.Context, read func(context.Context, int64) (*Record, error), provider *Record) (*Record, error) {
	if provider == nil || !provider.IsCredentialShadow() {
		return provider, nil
	}
	parent, err := read(ctx, *provider.ParentProviderID)
	if err != nil {
		return nil, fmt.Errorf("resolve spark shadow parent %d: %w", *provider.ParentProviderID, err)
	}
	if parent == nil {
		return nil, fmt.Errorf("spark shadow parent %d not found", *provider.ParentProviderID)
	}
	// 防御:创建路径已禁二级影子(G6),此处再挡一层——畸形数据/手工 DB 写出的影子→影子链
	// 会让凭据解析停在无凭据的一级影子(只解一层),fail-closed 比静默返回坏母更安全(外审第6轮)。
	if parent.IsCredentialShadow() {
		return nil, fmt.Errorf("spark shadow parent %d is itself a shadow", parent.ID)
	}
	if !parent.IsOpenAIOAuth() {
		return nil, fmt.Errorf("spark shadow parent %d is not OpenAI OAuth", parent.ID)
	}
	return parent, nil
}
