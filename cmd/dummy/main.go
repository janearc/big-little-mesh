// Command dummy is good-citizen's reference service. It is two things at once:
// the canonical example a new citizen copies, and the harness we prove
// emit/consume with -- so a new good-citizen release is validated against the
// dummy, never against a production service (delightd, paling, ...). It has its
// own service id so its heartbeats and events are distinguishable on the bus.
//
// It does the minimum a citizen does: emit a heartbeat, watch an inbox, and
// react to new files (here, just log them; a real citizen would process them).
// Kafka is best-effort -- with no broker reachable it runs with emission off.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/janearc/blm/citizen"
	"github.com/janearc/blm/emit"
	observabilityproto "github.com/janearc/blm/proto/observability/v1"
	"github.com/janearc/blm/watcher"
)

// serviceID is the dummy's identity on the bus -- deliberately not a real
// service name, so its telemetry never masquerades as a production citizen's.
const serviceID = "good-citizen-dummy"

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	inbox := getenv("DUMMY_INBOX", "/tmp/good-citizen-dummy/inbox")
	brokers := strings.Split(getenv("KAFKA_BROKERS", "kafka:9092"), ",")
	srURL := getenv("SCHEMA_REGISTRY_URL", "http://schema-registry:8081")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Emission is best-effort: a down broker disables it but never stops the
	// service -- the same stance every citizen takes.
	var pub *emit.Publisher
	if p, err := emit.New(ctx, brokers, srURL); err != nil {
		log.Warn("kafka emission disabled", "err", err)
	} else {
		pub = p
		defer pub.Close()
		log.Info("kafka emission ready")
		go citizen.Heartbeat(ctx, pub, serviceID, observabilityproto.Schema, 15*time.Second, log)
	}

	// Watch the inbox for new files and react. A real citizen would process the
	// file and emit a domain event; the dummy just logs, which is enough to prove
	// the watch loop end to end.
	w := watcher.New(watcher.NewFilesOracle(inbox), 5*time.Second, func(ctx context.Context, r watcher.Result) error {
		log.Info("new inputs detected", "count", len(r.Items), "items", r.Items)
		return nil
	}, log)
	go w.Run(ctx)

	log.Info("good-citizen dummy running", "service", serviceID, "inbox", inbox)
	<-ctx.Done()
	log.Info("good-citizen dummy stopped")
}
