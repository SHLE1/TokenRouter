package provider

import (
	"context"
	"math/rand/v2"
	"net/http"
	"sync"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/provider"
	"github.com/TokenFlux/TokenRouter/internal/settings"
)

const (
	ollamaCloudUsageMaxSessionBytes = 16 * 1024
	ollamaCloudUsageConcurrency     = 4
	ollamaCloudUsageLeaderLockKey   = "ollama:cloud:usage:leader"
)

// 夹具保存测试需要的提供商行，查询返回独立副本。
type ollamaUsageRows struct {
	mu        sync.Mutex
	providers map[int64]*provider.Record
}

func (r *ollamaUsageRows) GetByID(_ context.Context, id int64) (*provider.Record, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	v := r.providers[id]
	if v == nil {
		return nil, provider.ErrProviderNotFound
	}
	copy := cloneOllamaUsageTestProvider(*v)
	return &copy, nil
}

// 设置替身提供读取器使用的两个键值操作。
type ollamaUsageSettings struct {
	settings.Repository
	mu     sync.Mutex
	values map[string]string
}

func (r *ollamaUsageSettings) GetValue(_ context.Context, key string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	v, ok := r.values[key]
	if !ok {
		return "", settings.ErrSettingNotFound
	}
	return v, nil
}

func (r *ollamaUsageSettings) Set(_ context.Context, key, value string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.values == nil {
		r.values = make(map[string]string)
	}
	r.values[key] = value
	return nil
}

type ollamaUsageTransport interface {
	Do(*http.Request, string, int64, int) (*http.Response, error)
}

// 夹具注入可调时钟和锁替身，查询、缓存和启停使用生产实现。
type ollamaUsageContract struct {
	*provider.OllamaCloudUsageService
	now       func() time.Time
	lockCache provider.CNMonitorLeader
}

func newOllamaUsageContract(repo provider.OllamaProviderReader, transport ollamaUsageTransport, source provider.RuntimeSettingsStore, cipher provider.OllamaSessionCipher, fixedKey bool) *ollamaUsageContract {
	s := &ollamaUsageContract{now: time.Now}
	options := provider.OllamaUsageOptions{
		EncryptionKeyConfigured: fixedKey,
		Now:                     func() time.Time { return s.now() },
		Jitter:                  rand.Int64N,
		InstanceID:              "ollama-contract",
		Lease: func(ctx context.Context, key, owner string, ttl time.Duration) (func(), bool) {
			return provider.AcquireSingletonLease(ctx, s.lockCache, nil, key, owner, ttl)
		},
	}
	if transport != nil {
		options.Fetch = OllamaUsageFetcher(transport.Do)
	}
	var config provider.OllamaUsageSettingsStore
	if source != nil {
		config = provider.NewRuntimeSettings(source, settings.ErrSettingNotFound)
	}
	s.OllamaCloudUsageService = provider.NewOllamaCloudUsageService(repo, config, cipher, options)
	return s
}

// 内存 leader 模拟同一键的持有者比较，Redis 行为由集成测试覆盖。
type ollamaUsageLeader struct {
	mu     sync.Mutex
	owners map[string]string
}

func (l *ollamaUsageLeader) TryAcquireLeaderLock(_ context.Context, key, owner string, _ time.Duration) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.owners == nil {
		l.owners = make(map[string]string)
	}
	if _, ok := l.owners[key]; ok {
		return false, nil
	}
	l.owners[key] = owner
	return true, nil
}

func (l *ollamaUsageLeader) ReleaseLeaderLock(_ context.Context, key, owner string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.owners[key] == owner {
		delete(l.owners, key)
	}
	return nil
}
