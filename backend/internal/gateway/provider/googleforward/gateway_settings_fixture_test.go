package googleforward_test

import (
	gatewayprovider "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	gatewaytestkit "github.com/TokenFlux/TokenRouter/internal/gateway/testkit"
	"github.com/TokenFlux/TokenRouter/internal/settings"
)

// newExecutionReadersFixture 将测试替身绑定到执行时的设置读取接口。
func newExecutionReadersFixture(repo settings.Repository) *gatewayprovider.RuntimeReaders {
	if repo != nil {
		repo = settings.New(repo)
	}
	value := gatewaytestkit.RuntimeReaders(repo)
	return value
}
