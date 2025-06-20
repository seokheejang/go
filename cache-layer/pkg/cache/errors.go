package cache

import "errors"

// Common errors
var (
	ErrNotFound = errors.New("cache: key not found")
)
