package billing

import "github.com/TokenFlux/TokenRouter/internal/pkg/apperror"

var ErrImageTaskPricingMissing = apperror.New(apperror.CategoryBadRequest, "BATCH_IMAGE_SETTLEMENT_PRICING_MISSING", "batch image settlement pricing is missing")
