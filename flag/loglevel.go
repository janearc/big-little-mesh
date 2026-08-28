// Package flag is the fleet's flipr consumption for TUNING values -- today,
// the log level. One home, per the no-copies rule: every Go service gets
// "log.level through flipr" by starting one poller, and the resolution
// precedence (service's own flag over the _global scope) is flipr's,
// server-side, so this package never reimplements it.
//
// A DELIBERATE DIVERGENCE FROM THE RULED GATE SEMANTICS, stated rather than
// slipped in: the reference client's rule -- flipr unreachable means STOP --
// is for GATES on expensive actions, where acting on a stale value is the
// harm. The log level is TUNING: the harm inverts. A service that stops (or
// wedges its boot) because the tuning store is away has turned a knob into a
// dependency, and hm in particular must run on a fresh cluster where flipr
// has not arrived yet. So: unreachable KEEPS the current level and says so
// once on the transition, never per poll. The cache-is-not-a-fallback ruling
// stays true for gates; this is not a gate.
package flag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// LogLevelKey is the flag key, named by flipr's own SetDebugFromFlag: the one
// knob, "log.level", in the service's namespace or _global.
const LogLevelKey = "log.level"

// ParseLevel maps flag values to slog levels. Unknown returns an error so the
// caller can refuse loudly rather than guess.
func ParseLevel(v string) (slog.Level, error) {
	switch v {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	}
	return 0, fmt.Errorf("unknown log level %q (debug|info|warn|error)", v)
}

// PollLogLevel follows the log.level flag for service, applying changes to lv
// until ctx ends. base is flipr's address BY NAME. Blocks; run it as a
// goroutine. interval <= 0 defaults to 15s.
func PollLogLevel(ctx context.Context, base, service string, lv *slog.LevelVar, log *slog.Logger, interval time.Duration) {
	if interval <= 0 {
		interval = 15 * time.Second
	}
	client := &http.Client{Timeout: 3 * time.Second}
	var lastApplied string
	var down bool

	tick := time.NewTicker(interval)
	defer tick.Stop()
	for {
		val, err := getFlag(ctx, client, base, service, LogLevelKey)
		switch {
		case err != nil:
			if !down {
				log.Warn("flipr unreachable; log level unchanged",
					"base", base, "err", err, "held_level", lv.Level().String())
				down = true
			}
		case val == "":
			// declared nowhere, service or global: the boot default stands.
			down = false
		default:
			down = false
			if val != lastApplied {
				lvl, perr := ParseLevel(val)
				if perr != nil {
					log.Warn("log.level flag holds an unknown value; keeping current",
						"value", val, "err", perr)
					lastApplied = val // do not re-warn every poll for the same bad value
					break
				}
				lv.Set(lvl)
				lastApplied = val
				// INFO would be invisible after a flip to warn; the one line
				// that explains the change must outrank the change.
				log.Warn("log level set from flipr", "level", val)
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
		}
	}
}

// getFlag asks flipr for one key in service@_. protojson over plain HTTP,
// matching the reference client; the server merges the _global scope, so a
// missing service flag returns the global one without this code knowing.
// A "no flag" answer is ("", nil): absence is a state, not an outage.
func getFlag(ctx context.Context, c *http.Client, base, service, key string) (string, error) {
	body, _ := json.Marshal(map[string]string{"service": service, "version": "_", "key": key})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		base+"/flipr.v1.FliprService/GetFlag", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var out struct {
		Flag struct {
			Value struct {
				StringValue string `json:"stringValue"`
			} `json:"value"`
		} `json:"flag"`
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("flipr answered %d and unparseable json: %w", resp.StatusCode, err)
	}
	if out.Error != "" {
		return "", nil // "no flag X in svc@_": absence, by design not an error here
	}
	return out.Flag.Value.StringValue, nil
}
