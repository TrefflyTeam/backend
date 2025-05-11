package models

import "time"

type RateLimitResult struct {
	Allowed    bool
	Remaining  int
	ResetAt    time.Time
}
