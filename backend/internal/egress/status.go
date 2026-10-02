package egress

// StatusActive 和 StatusExpired 是持久化的代理状态值。
const (
	StatusActive  = "active"
	StatusExpired = "expired"
)

// Diagnostics 由装配注入原日志出口，核心不安装或持有日志后端。
type Diagnostics struct {
	Logf func(component, format string, args ...any)
}

func (d Diagnostics) Log(component, format string, args ...any) {
	if d.Logf != nil {
		d.Logf(component, format, args...)
	}
}

func diagnosticsOption(options []Diagnostics) Diagnostics {
	if len(options) > 0 {
		return options[0]
	}
	return Diagnostics{}
}
