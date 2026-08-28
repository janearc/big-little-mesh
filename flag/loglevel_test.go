package flag

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// fliprStub answers GetFlag with whatever value the test sets; "" means
// answer the no-flag error shape, "DOWN" means refuse the connection.
type fliprStub struct {
	val atomic.Value // string
	srv *httptest.Server
}

func newStub() *fliprStub {
	s := &fliprStub{}
	s.val.Store("")
	s.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := s.val.Load().(string)
		if v == "" {
			fmt.Fprintf(w, `{"error":"no flag \"log.level\" in x@_"}`)
			return
		}
		fmt.Fprintf(w, `{"flag":{"key":"log.level","value":{"stringValue":%q}}}`, v)
	}))
	return s
}

func poll(t *testing.T, s *fliprStub, lv *slog.LevelVar) (context.CancelFunc, func()) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		PollLogLevel(ctx, s.srv.URL, "hm", lv, slog.Default(), 10*time.Millisecond)
	}()
	return cancel, func() { <-done }
}

func waitLevel(t *testing.T, lv *slog.LevelVar, want slog.Level) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if lv.Level() == want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("level never reached %v (at %v)", want, lv.Level())
}

func TestFlagDrivesTheLevel(t *testing.T) {
	s := newStub()
	defer s.srv.Close()
	var lv slog.LevelVar
	lv.Set(slog.LevelWarn) // the boot default Jane named

	cancel, wait := poll(t, s, &lv)
	defer wait()
	defer cancel()

	// no flag anywhere: the default stands
	time.Sleep(50 * time.Millisecond)
	if lv.Level() != slog.LevelWarn {
		t.Fatalf("absent flag moved the level to %v", lv.Level())
	}
	// operator flips to debug
	s.val.Store("debug")
	waitLevel(t, &lv, slog.LevelDebug)
	// and back down
	s.val.Store("warn")
	waitLevel(t, &lv, slog.LevelWarn)
}

func TestUnknownValueKeepsCurrent(t *testing.T) {
	s := newStub()
	defer s.srv.Close()
	var lv slog.LevelVar
	lv.Set(slog.LevelInfo)
	cancel, wait := poll(t, s, &lv)
	defer wait()
	defer cancel()

	s.val.Store("shouting")
	time.Sleep(100 * time.Millisecond)
	if lv.Level() != slog.LevelInfo {
		t.Fatalf("unknown value moved the level to %v", lv.Level())
	}
}

func TestUnreachableKeepsCurrent(t *testing.T) {
	// The divergence under test: flipr GONE must hold the level, not stop
	// anything -- log level is tuning, not a gate.
	s := newStub()
	var lv slog.LevelVar
	lv.Set(slog.LevelWarn)
	cancel, wait := poll(t, s, &lv)
	defer wait()
	defer cancel()

	s.val.Store("debug")
	waitLevel(t, &lv, slog.LevelDebug)
	s.srv.Close() // flipr dies
	time.Sleep(100 * time.Millisecond)
	if lv.Level() != slog.LevelDebug {
		t.Fatalf("outage moved the level to %v", lv.Level())
	}
}

func TestParseLevel(t *testing.T) {
	for v, want := range map[string]slog.Level{
		"debug": slog.LevelDebug, "info": slog.LevelInfo,
		"warn": slog.LevelWarn, "warning": slog.LevelWarn, "error": slog.LevelError,
	} {
		got, err := ParseLevel(v)
		if err != nil || got != want {
			t.Fatalf("ParseLevel(%q) = %v, %v", v, got, err)
		}
	}
	if _, err := ParseLevel("loud"); err == nil {
		t.Fatal("unknown level parsed")
	}
}
