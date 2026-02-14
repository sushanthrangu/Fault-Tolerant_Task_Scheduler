package domain

import "errors"

var (
	ErrInvalidInput = errors.New("invalid_input")
	ErrNotFound     = errors.New("not_found")
)
