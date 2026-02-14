package worker

import (
	"math"
	"math/rand"
	"time"
)

// BackoffConfig controls exponential backoff behavior.
type BackoffConfig struct {
	Base   time.Duration // e.g., 500ms
	Max    time.Duration // e.g., 30s
	Jitter float64       // e.g., 0.20 => ±20%
}

// Next returns the delay for a given attempt (attempt starts at 1).
func (c BackoffConfig) Next(attempt int) time.Duration {
	if attempt <= 1 {
		return jittered(c.Base, c.Jitter)
	}

	// base * 2^(attempt-1), capped
	exp := float64(c.Base) * math.Pow(2, float64(attempt-1))
	delay := time.Duration(exp)
	if c.Max > 0 && delay > c.Max {
		delay = c.Max
	}
	return jittered(delay, c.Jitter)
}

func jittered(d time.Duration, jitter float64) time.Duration {
	if d <= 0 {
		return 0
	}
	if jitter <= 0 {
		return d
	}
	// factor in [1-j, 1+j]
	f := 1 + (rand.Float64()*2-1)*jitter
	return time.Duration(float64(d) * f)
}
