package api

import (
	"context"
	"testing"
	"time"
)

func TestRateLimiter_BurstThenBlock(t *testing.T) {
	rl := NewRateLimiter(10, time.Second)
	defer rl.Close()

	ctx := context.Background()

	// 10 rapid calls should succeed immediately.
	for range 10 {
		if err := rl.Wait(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	// 11th should block; use short timeout to prove it.
	ctx2, cancel := context.WithTimeout(ctx, 20*time.Millisecond)
	defer cancel()

	err := rl.Wait(ctx2)
	if err == nil {
		t.Fatal("expected 11th call to block and timeout, but it succeeded")
	}
	if err != context.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got: %v", err)
	}
}

func TestRateLimiter_Refill(t *testing.T) {
	rl := NewRateLimiter(2, 200*time.Millisecond)
	defer rl.Close()

	ctx := context.Background()

	// Drain both tokens.
	_ = rl.Wait(ctx)
	_ = rl.Wait(ctx)

	// Wait for refill (200ms / 2 = 100ms per token, wait a bit extra).
	time.Sleep(150 * time.Millisecond)

	ctx2, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	if err := rl.Wait(ctx2); err != nil {
		t.Fatalf("expected refill token available, got: %v", err)
	}
}

func TestRateLimiter_ContextCancelled(t *testing.T) {
	rl := NewRateLimiter(1, time.Second)
	defer rl.Close()

	ctx := context.Background()
	_ = rl.Wait(ctx) // drain

	ctx2, cancel := context.WithCancel(ctx)
	cancel()

	if err := rl.Wait(ctx2); err != context.Canceled {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}
