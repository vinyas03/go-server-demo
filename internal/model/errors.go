package model

import "errors"

var (
	ErrNotFound      = errors.New("resource not found")
	ErrInvalidInput  = errors.New("invalid input")
	ErrAlreadyExists = errors.New("resource already exists")
	ErrUnauthorized  = errors.New("unauthorized: missing user credentials")
	ErrForbidden     = errors.New("forbidden: insufficient permissions")
)
