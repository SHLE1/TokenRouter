package provider

// ProviderRefreshPlatformPolicy 组合刷新资格判断和错误快照。
func ProviderRefreshPlatformPolicy() RefreshPlatformPolicy {
	return RefreshPlatformPolicy{
		Eligibility: GrokOAuthRequestProviderEligibilityError,
		MissingRefreshToken: func() error {
			return ErrGrokOAuthRefreshTokenMissing
		},
		SnapshotError: WithGrokCredentialFailureSnapshot,
		ConfigurationError: func(err error) error {
			return &ProviderConfigurationRefreshError{Cause: err}
		},
		ContainmentError: func(err error) error {
			return &ProviderCycleContainmentRefreshError{Cause: err}
		},
	}
}
