package bridge

import (
	"crypto/rand"
	"time"
)

// testRuntime 为协议转换测试提供时间、随机数和诊断回调。
func testRuntime() Runtime { return Runtime{Now: time.Now, ReadRandom: rand.Read} }
