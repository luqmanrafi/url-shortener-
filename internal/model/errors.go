package model

import "errors"
var (
	ErrNotFound = errors.New("url not found")
	ErrCodeExists = errors.New("Code already exists")
	ErrInvalidUrl = errors.New("Invalid url")
)