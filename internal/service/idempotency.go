package service

import "strings"

// NormalizeIdempotencyKey trims and returns nil if empty.
func NormalizeIdempotencyKey(k string) *string {
	k = strings.TrimSpace(k)
	if k == "" {
		return nil
	}
	return &k
}
