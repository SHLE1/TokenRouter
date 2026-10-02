package media

import "time"

// DefaultRealtimeDialTimeout 设置下游升级前的上游握手超时，会话时长由连接生命周期控制。
const DefaultRealtimeDialTimeout = 12 * time.Second
