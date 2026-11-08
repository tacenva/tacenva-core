package genericerror

import "errors"

var (
	ErrEmptyValue    = errors.New("Value cannot be empty")
	ErrImutableField = errors.New("Attempted to update an immutable field")
	ErrAccessRevoked = errors.New("access has been revoked")
	ErrAccessPending = errors.New("access is pending approval")
)
