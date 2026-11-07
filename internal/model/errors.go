package model

import "errors"
var (
	ErrNotFound = errors.New("url not found")
	ErrCodeExists = errors.New("Code already exists")
	ErrInvalidUrl = errors.New("Invalid url")
	ErrInvalidAlias = errors.New("Invalid url alias")
	ErrExpired = errors.New("Url has expired")
)
