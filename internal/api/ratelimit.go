package api

import (
	"context"
	"time"
)

// RateLimiter implements a simple token bucket rate limiter.
type RateLimiter struct {
	tokens   chan struct{}
	interval time.Duration
	quit     chan struct{}
}

// NewRateLimiter creates a token bucket that allows rate requests per interval.
// Tokens are pre-filled and refilled at a steady rate.
func NewRateLimiter(rate int, interval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		tokens:   make(chan struct{}, rate),
		interval: interval,
		quit:     make(chan struct{}),
	}

	// Pre-fill tokens.
	for range rate {
		rl.tokens <- struct{}{}
	}

	// Refill goroutine: add one token every interval/rate.
	refillInterval := interval / time.Duration(rate)
	go func() {
		ticker := time.NewTicker(refillInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				select {
				case rl.tokens <- struct{}{}:
				default:
					// Bucket full, discard.
				}
			case <-rl.quit:
				return
			}
		}
	}()

	return rl
}

// Wait blocks until a token is available or ctx is cancelled.
func (r *RateLimiter) Wait(ctx context.Context) error {
	select {
	case <-r.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Close stops the refill goroutine.
func (r *RateLimiter) Close() {
	close(r.quit)
}
