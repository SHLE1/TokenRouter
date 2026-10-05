package routing

import (
	"strings"

	"github.com/TokenFlux/TokenRouter/internal/pkg/apperror"
	"github.com/TokenFlux/TokenRouter/internal/pkg/locale"
	"github.com/TokenFlux/TokenRouter/internal/routing/accessview"
)

type (
	GroupCopy         = accessview.GroupCopy
	GroupLocalization = accessview.GroupLocalization
)

// ValidateGroupCopy 对用户展示名称执行独立校验。
func ValidateGroupCopy(copy GroupCopy) error {
	if strings.TrimSpace(copy.DisplayName) == "" || len([]rune(copy.DisplayName)) > 100 {
		return apperror.BadRequest("GROUP_DISPLAY_NAME_REQUIRED", "Group display name is required and must not exceed 100 characters.")
	}
	if len(copy.Description) > 65536 {
		return apperror.BadRequest("GROUP_DESCRIPTION_TOO_LONG", "Group description exceeds the allowed length.")
	}
	return nil
}

// GroupDisplay 返回用户文案，原文缺失时使用历史名称和描述。
func GroupDisplay(group *Group, language string) (GroupCopy, locale.Resolution) {
	if group.Localization.Revision == 0 {
		return GroupCopy{DisplayName: group.Name, Description: group.Description}, locale.Resolution{Fallback: true}
	}
	return locale.Content[GroupCopy](group.Localization).Resolve(language)
}

// GroupSearchTexts 返回用户可见名称和描述的有效语言版本。
func GroupSearchTexts(group *Group) []string {
	if group.Localization.Revision == 0 {
		return nil
	}
	return locale.SearchTexts(locale.Content[GroupCopy](group.Localization), func(copy GroupCopy) []string { return []string{copy.DisplayName, copy.Description} })
}
