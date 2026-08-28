package consume

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// TestParseFrame_ReturnsId: same frames StripFrame reads, with the schema id
// kept -- the registry-aware consumer's whole reason to exist.
func TestParseFrame_ReturnsId(t *testing.T) {
	payload := []byte("hello-protobuf")
	id, got, err := ParseFrame(frame(7, []byte{0x00}, payload))
	if err != nil {
		t.Fatalf("ParseFrame: %v", err)
	}
	if id != 7 {
		t.Errorf("id = %d, want 7", id)
	}
	if !bytes.Equal(got, payload) {
		t.Errorf("payload = %q, want %q", got, payload)
	}
}

// TestParseFrame_RefusesBareBytes: raw JSON on the bus must not parse -- the
// 164MB-loop lesson; the frame check is the contract's front door.
func TestParseFrame_RefusesBareBytes(t *testing.T) {
	if _, _, err := ParseFrame([]byte(`{"msg":"raw json"}`)); err == nil {
		t.Fatal("bare JSON parsed as an SR frame")
	}
}

// TestResolver_CachesForever: ids are immutable in the registry, so the
// second lookup must not touch the wire.
func TestResolver_CachesForever(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.URL.Path != "/schemas/ids/42/versions" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`[{"subject":"lease.v1.LeaseVerdict","version":1}]`))
	}))
	defer srv.Close()

	r := NewResolver(srv.URL)
	for range 3 {
		s, err := r.Subject(context.Background(), 42)
		if err != nil {
			t.Fatalf("Subject: %v", err)
		}
		if s != "lease.v1.LeaseVerdict" {
			t.Fatalf("subject = %q", s)
		}
	}
	if hits.Load() != 1 {
		t.Fatalf("registry hit %d times; the cache leaked", hits.Load())
	}
}

// TestResolver_ErrorsAreNotCached: an outage answer must not poison the
// cache -- the next call after recovery resolves.
func TestResolver_ErrorsAreNotCached(t *testing.T) {
	var fail atomic.Bool
	fail.Store(true)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if fail.Load() {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(`[{"subject":"observability.v1.ServiceHealthHeartbeat","version":3}]`))
	}))
	defer srv.Close()

	r := NewResolver(srv.URL)
	if _, err := r.Subject(context.Background(), 9); err == nil {
		t.Fatal("expected the outage to surface as an error")
	}
	fail.Store(false)
	s, err := r.Subject(context.Background(), 9)
	if err != nil {
		t.Fatalf("post-recovery Subject: %v", err)
	}
	if s != "observability.v1.ServiceHealthHeartbeat" {
		t.Fatalf("subject = %q", s)
	}
}
