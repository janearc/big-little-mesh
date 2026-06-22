package scheduler

import (
	"context"
	"testing"
	"time"
)

// TestEvery_FiresOnInterval proves a job registered with Every actually runs
// under the running scheduler, and that Stop halts cleanly.
func TestEvery_FiresOnInterval(t *testing.T) {
	s, err := New(nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.Start(ctx)
	defer s.Stop()

	fired := make(chan struct{}, 1)
	if err := s.Every("tick", 10*time.Millisecond, func(ctx context.Context) error {
		select {
		case fired <- struct{}{}:
		default:
		}
		return nil
	}); err != nil {
		t.Fatalf("Every: %v", err)
	}

	select {
	case <-fired:
	case <-time.After(2 * time.Second):
		t.Fatal("scheduled job never fired")
	}
}

// TestCron_RejectsBadExpression pins that a malformed cron expression surfaces
// as an error at registration rather than silently never firing.
func TestCron_RejectsBadExpression(t *testing.T) {
	s, err := New(nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := s.Cron("bad", "not a cron", func(context.Context) error { return nil }); err == nil {
		t.Error("expected error for malformed cron expression")
	}
}
