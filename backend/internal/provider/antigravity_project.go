package provider

import "strings"

const antigravityProjectIDFallbackCredentialKey = "antigravity_project_id"

func ResolveAntigravityProjectID(provider *Record, missing error) (string, error) {
	if provider == nil {
		return "", missing
	}
	if projectID := strings.TrimSpace(provider.GetCredential("project_id")); projectID != "" {
		return projectID, nil
	}
	if projectID := strings.TrimSpace(provider.GetCredential(antigravityProjectIDFallbackCredentialKey)); projectID != "" {
		return projectID, nil
	}
	if projectID := strings.TrimSpace(provider.GetExtraString(antigravityProjectIDFallbackCredentialKey)); projectID != "" {
		return projectID, nil
	}
	return "", missing
}
