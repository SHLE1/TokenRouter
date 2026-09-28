package moderation

import "github.com/TokenFlux/TokenRouter/internal/notification"

func buildContentModerationAccountDisabledEmailBody(site string, v *ContentModerationLog, c *ContentModerationConfig) string {
	return notification.BuildContentModerationAccountDisabledEmailBody(site, riskLog(v), riskPolicy(c))
}
