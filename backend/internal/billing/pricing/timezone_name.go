package pricing

import (
	"fmt"
	"strings"
)

func ValidateTimezoneName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("timezone is required")
	}
	if name == "Local" {
		return fmt.Errorf("local is not a supported timezone")
	}
	return nil
}
