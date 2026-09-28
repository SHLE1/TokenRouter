package app

import (
	"log"
	"runtime"
	"runtime/debug"

	"github.com/TokenFlux/TokenRouter/internal/app/lifecycle"
	"github.com/TokenFlux/TokenRouter/internal/ops"
)

func provideOpsErrorQueue(manager *lifecycle.Manager) *ops.ErrorLogQueue {
	q := ops.NewErrorLogQueue(ops.ErrorLogQueueOptions{Processors: func() int { return runtime.GOMAXPROCS(0) }, Logf: log.Printf, Stack: debug.Stack})
	manager.Register(lifecycle.Hook{Name: "OpsErrorLogWorkers", StartOrder: 924, StopOrder: 76, Stop: q.Shutdown})
	return q
}
