package dto_test

import "github.com/TokenFlux/TokenRouter/internal/usage/httpapi/dto"

func requestTypeStringPtr(v *int16) *string { return dto.RequestTypeStringPtr(v) }
