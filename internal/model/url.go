package model

import "time"

type URL struct {
	Code       string
	Original   string
	ClickCount int
	CreatedAt  time.Time
	ExpiresAt *time.Time
}