package storage

import "errors"

// Common storage errors
var (
	ErrClientNotFound  = errors.New("oauth client not found")
	ErrUserNotFound    = errors.New("oauth user not found")
	ErrTokenNotFound   = errors.New("oauth token not found")
	ErrTokenExpired    = errors.New("oauth token expired")
	ErrCodeNotFound    = errors.New("authorization code not found")
	ErrCodeExpired     = errors.New("authorization code expired")
	ErrScopeNotFound   = errors.New("oauth scope not found")
	ErrRoleNotFound    = errors.New("oauth role not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)