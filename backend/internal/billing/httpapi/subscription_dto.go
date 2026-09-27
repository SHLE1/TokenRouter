// 订阅接口共享套餐、用户订阅和批量分配结果的 DTO。
package httpapi

import (
	billingdto "github.com/TokenFlux/TokenRouter/internal/billing/httpapi/dto"
)

type SubscriptionPlan = billingdto.SubscriptionPlan

type SubscriptionPlanGroup = billingdto.SubscriptionPlanGroup

type UserSubscription = billingdto.UserSubscription

type AdminUserSubscription = billingdto.AdminUserSubscription

type BulkAssignResult = billingdto.BulkAssignResult

type UserSummaryResponse = billingdto.UserSummaryResponse
