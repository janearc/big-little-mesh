// The registry-aware half of consuming: resolve a frame's schema id to its
// SUBJECT, so a consumer decides what a record IS from the registry's answer
// rather than from unknown-field residue. Under RecordNameStrategy the
// subject is the fully-qualified message name (lease.v1.LeaseVerdict,
// observability.v1.ServiceHealthHeartbeat), which is exactly the dispatch
// key a mixed-subject topic needs.
package consume

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Resolver caches schema-id -> subject lookups against one schema registry.
// Ids are immutable in the registry, so a hit is cached for the process
// lifetime; errors are never cached -- a registry outage is transient and
// the caller decides its own degraded behavior (hall-monitor falls back to
// its legacy discriminator, for instance).
type Resolver struct {
	url  string
	http *http.Client

	mu    sync.RWMutex
	cache map[int32]string
}

// NewResolver builds a resolver against the registry's base URL. The timeout
// mirrors emit's publisher: 5s, aggressive on purpose -- a slow registry
// must degrade the caller, not stall it.
func NewResolver(srURL string) *Resolver {
	return &Resolver{
		url:   strings.TrimRight(srURL, "/"),
		http:  &http.Client{Timeout: 5 * time.Second},
		cache: map[int32]string{},
	}
}

// Subject resolves one schema id. GET /schemas/ids/{id}/versions answers
// [{"subject":..., "version":...}]; the first row's subject is the answer
// (an id registered under several subjects is not a case this mesh produces
// -- RecordNameStrategy gives each message name its own subject).
func (r *Resolver) Subject(ctx context.Context, id int32) (string, error) {
	r.mu.RLock()
	s, ok := r.cache[id]
	r.mu.RUnlock()
	if ok {
		return s, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/schemas/ids/%d/versions", r.url, id), nil)
	if err != nil {
		return "", err
	}
	resp, err := r.http.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("registry answered %d for schema id %d", resp.StatusCode, id)
	}
	var rows []struct {
		Subject string `json:"subject"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return "", err
	}
	if len(rows) == 0 || rows[0].Subject == "" {
		return "", fmt.Errorf("schema id %d has no subject", id)
	}

	r.mu.Lock()
	r.cache[id] = rows[0].Subject
	r.mu.Unlock()
	return rows[0].Subject, nil
}
