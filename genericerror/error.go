package genericerror

import (
	"fmt"
	"strings"
)

var (
	ErrImutableField    = fmt.Errorf("attempted to update an immutable field")
	ErrAccessRevoked    = fmt.Errorf("access has been revoked")
	ErrAccessPending    = fmt.Errorf("access is pending approval")
	ErrPermissionDenied = fmt.Errorf("permission denied")
)

func ErrRequired(field string) error {
	return fmt.Errorf("%s is required", field)
}

func ValidateRequired(fields map[string]any) error {
	for name, value := range fields {
		if strings.TrimSpace(fmt.Sprint(value)) == "" {
			return ErrRequired(name)
		}
	}

	return nil
}
