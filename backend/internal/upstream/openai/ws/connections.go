package ws

import (
	"errors"
	"sync"

	"github.com/TokenFlux/TokenRouter/internal/upstream/openai"
)

// OpenAIWSConnections 持有唯一按需连接池与共享拨号器，关闭后不能重新建池。
type OpenAIWSConnections struct {
	Options    *WSPoolOptions
	dialer     openai.WSClientDialer
	dialerOnce sync.Once
	pool       *WSConnPool
	poolOnce   sync.Once
	poolMu     sync.Mutex
	closed     bool
}

// NewOpenAIWSConnections 保存配置和拨号器，连接池在首次使用时启动。
func NewOpenAIWSConnections(options *WSPoolOptions, dialer openai.WSClientDialer, pools ...*WSConnPool) *OpenAIWSConnections {
	value := &OpenAIWSConnections{Options: options, dialer: dialer}
	if len(pools) > 0 {
		value.pool = pools[0]
	}
	return value
}

func (s *OpenAIWSConnections) Pool() *WSConnPool {
	if s == nil {
		return nil
	}
	s.poolMu.Lock()
	defer s.poolMu.Unlock()
	if s.closed {
		return s.pool
	}
	s.poolOnce.Do(func() {
		if s.pool == nil {
			s.pool = NewWSConnPool(s.Options)
			s.pool.Start()
		}
	})
	return s.pool
}

func (s *OpenAIWSConnections) Dialer() openai.WSClientDialer {
	if s == nil {
		return nil
	}
	s.dialerOnce.Do(func() {
		if s.dialer == nil {
			s.dialer = openai.NewDefaultWSClientDialer()
		}
	})
	return s.dialer
}

func (s *OpenAIWSConnections) InvalidateProvider(providerID int64) {
	if pool := s.Pool(); pool != nil {
		pool.ClearProvider(providerID)
	}
}

// Close 关闭连接池并标记停止，后续 Pool 调用返回已关闭的池。
func (s *OpenAIWSConnections) Close() {
	if s == nil {
		return
	}
	s.poolMu.Lock()
	s.closed = true
	pool := s.pool
	s.poolMu.Unlock()
	if pool != nil {
		pool.Close()
	}
}

// UpdateOptions 更新待创建或已创建的池，参数发布不会主动创建连接。
func (s *OpenAIWSConnections) UpdateOptions(options WSPoolOptions) error {
	s.poolMu.Lock()
	defer s.poolMu.Unlock()
	if s.closed {
		return errors.New("responses websocket connections are stopped")
	}
	s.Options = &options
	if s.pool != nil {
		return s.pool.UpdateOptions(options)
	}
	return nil
}
