package model

import "errors"

var (
	ErrNotFound       = errors.New("not found")
	ErrInvalidInput   = errors.New("invalid input")
	ErrDuplicateKey   = errors.New("idempotency key reused with different request body")
	ErrInternalServer = errors.New("internal server error")
)
