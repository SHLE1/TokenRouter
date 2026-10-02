package app

import (
	"log/slog"
	"os"

	"github.com/TokenFlux/TokenRouter/internal/billing"
	paymentadapter "github.com/TokenFlux/TokenRouter/internal/payment/provider"
	settingscore "github.com/TokenFlux/TokenRouter/internal/settings"

	"github.com/TokenFlux/TokenRouter/internal/config"
	"github.com/TokenFlux/TokenRouter/internal/payment"

	paymentpostgres "github.com/TokenFlux/TokenRouter/internal/payment/postgres"
)

// payment.EncryptionKey 是 AES-256 支付密钥类型，长度为 32 字节。
// 独立类型供 Wire 区分其他 []byte 参数。

// providePaymentEncryptionKey 从配置中的 TOTP 加密密钥派生支付密钥。
// 密钥为空时返回 nil，依赖加密的支付功能关闭。
// 非空密钥的十六进制格式或长度无效时返回错误，应用启动失败。
func providePaymentEncryptionKey(cfg *config.Config) (payment.EncryptionKey, error) {
	if cfg == nil {
		slog.Warn("payment encryption key not configured — encrypted payment config and resume signing will be unavailable")
		return nil, nil
	}
	key, warning, err := payment.ConfiguredEncryptionKey(cfg.Totp.EncryptionKey, cfg.Totp.EncryptionKeyConfigured)
	if warning != "" {
		slog.Warn(warning)
	}
	return key, err
}

// providePaymentRegistry 创建支付提供方注册表。
// 应用启动后在运行期间登记提供方。
func providePaymentRegistry() *payment.Registry {
	return payment.NewRegistry()
}

// providePaymentLoadBalancer 使用 Ent 客户端创建 DefaultLoadBalancer。
func providePaymentLoadBalancer(store *paymentpostgres.InstanceStore, key payment.EncryptionKey) *payment.DefaultLoadBalancer {
	return payment.NewDefaultLoadBalancer(store, []byte(key), payment.SelectionRuntime{Observe: paymentSelectionLog})
}

func paymentSelectionLog(level, message string, attrs ...any) {
	if level == "warn" {
		slog.Warn(message, attrs...)
	} else {
		slog.Info(message, attrs...)
	}
}

func providePaymentConfigCore(store *paymentpostgres.InstanceStore, settings settingscore.Repository, key payment.EncryptionKey, plans *billing.Plans) *payment.ConfigService {
	return payment.NewConfigService(store, settings, []byte(key), plans, payment.ConfigurationRuntime{CreateProvider: paymentadapter.CreateProvider, LookupEnv: os.LookupEnv, Warn: slog.Warn})
}
