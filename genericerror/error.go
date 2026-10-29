package genericerror

import "errors"

var (
	ErrEmptyValue    error = errors.New("Value cannot be empty")
	ErrImutableField error = errors.New("Attempted to update an immutable field")
)
