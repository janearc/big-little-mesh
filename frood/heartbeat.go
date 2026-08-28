// Package frood is the high-level surface a service wires up to become a frood of
// the mesh: the heartbeat it emits, and (as the package grows) the watch/secrets/
// scheduler glue. It sits on top of the emit package so a service gets fleet
// observability for free by calling one function.
package frood

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/janearc/big-little-mesh/emit"
	observabilityv1 "github.com/janearc/big-little-mesh/gen/go/observability/v1"
	observabilityproto "github.com/janearc/big-little-mesh/proto/observability/v1"
)

const (
	// TopicObservability is the fleet heartbeat topic. Code is truth: this is
	// observability.events, not observability.heartbeat.
	TopicObservability = "observability.events"
	// SubjectHeartbeat is the RecordNameStrategy subject for the heartbeat schema. It is the
	// canonical observabilityproto constant -- the SAME one admin.FleetSubjects provisions under --
	// so the subject the producer emits and the subject the provisioner registers cannot drift.
	SubjectHeartbeat = observabilityproto.SubjectServiceHealthHeartbeat
)

// Heartbeat emits a ServiceHealthHeartbeat for serviceName every interval via
// pub, until ctx is cancelled. Best-effort: a publish failure is logged, never
// fatal -- a frood whose telemetry is down keeps doing its job. A nil pub is a
// no-op (emission disabled), so a caller can hold one unconditionally and let a
// missing broker be silent. Blocks; run it in a goroutine.
//
// proto/observability/v1.Schema); it is registered once per subject.
// Heartbeat supplies its own schema. It used to take a schemaText parameter,
// and that parameter had exactly one correct value -- the observability
// contract this function publishes -- so its only reachable behavior was the
// failure flipr hit: pass any other schema and a FULL_TRANSITIVE registry
// refuses every beat, best-effort, silently, and the citizen never appears in
// the authorized table. A parameter whose wrong values are all silent is not
// flexibility; callers that passed the old argument now fail to COMPILE,
// which is the notice.
func Heartbeat(ctx context.Context, pub *emit.Publisher, serviceName string, interval time.Duration, log *slog.Logger) {
	schemaText := observabilityproto.Schema
	if interval <= 0 {
		interval = 15 * time.Second
	}
	if log == nil {
		log = slog.Default()
	}
	start := time.Now()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			hb := &observabilityv1.ServiceHealthHeartbeat{
				ServiceName:    serviceName,
				CurrentState:   observabilityv1.HealthState_HEALTH_STATE_GREEN,
				UptimeSeconds:  uint32(time.Since(start).Seconds()),
				Timestamp:      timestamppb.Now(),
				IdempotencyKey: fmt.Sprintf("%s-hb-%d", serviceName, time.Now().UnixNano()),
			}
			if err := pub.Publish(ctx, TopicObservability, SubjectHeartbeat, schemaText, serviceName, hb); err != nil {
				log.Error("heartbeat emit failed", "service", serviceName, "err", err)
			}
		}
	}
}
