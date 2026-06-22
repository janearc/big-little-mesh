// Package edge registers a service with the fleet's edge so a public hostname
// reaches it through cloudflared -> traefik. The single cloudflared tunnel (its
// token a Kube Secret) and the public DNS are provisioned once at the edge; an
// individual service only needs traefik to know how to route its host. For a
// bare-metal daemon (off the cluster network, e.g. magpie/paling, which must run
// on the host for Metal) traefik can reach it only via a dynamic-config route the
// service installs itself -- exactly what RegisterRoute writes, codifying the
// route services currently hand-roll. In-cluster services use a static traefik
// IngressRoute manifest instead and do not need this.
//
// Provisioning the public hostname/DNS through the Cloudflare API (using the
// cloudflare-api token via the secret package) is a follow-up: it needs the zone
// and tunnel ids, and is an edge-level one-time step, not per-service.
package edge

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RegisterRoute writes a traefik file-provider route into dynamicDir so traefik
// serves host from serviceURL on the web entrypoint. name keys the router, the
// service, and the filename. Idempotent: re-registering overwrites the same file,
// so a service can call this unconditionally at startup.
func RegisterRoute(dynamicDir, name, host, serviceURL string) error {
	if dynamicDir == "" || name == "" || host == "" || serviceURL == "" {
		return fmt.Errorf("dynamicDir, name, host, and serviceURL are all required")
	}
	if err := os.MkdirAll(dynamicDir, 0o755); err != nil {
		return fmt.Errorf("create traefik dynamic dir: %w", err)
	}
	path := filepath.Join(dynamicDir, name+".yml")
	if err := os.WriteFile(path, []byte(routeYAML(name, host, serviceURL)), 0o644); err != nil {
		return fmt.Errorf("write route %s: %w", path, err)
	}
	return nil
}

// routeYAML renders the traefik file-provider document: a Host router on the web
// entrypoint pointing at a loadBalancer service. TLS terminates at the edge; the
// tunnel forwards cleartext to web, so the route only needs the web entrypoint.
func routeYAML(name, host, serviceURL string) string {
	var b strings.Builder
	b.WriteString("http:\n  routers:\n")
	fmt.Fprintf(&b, "    %s:\n", name)
	fmt.Fprintf(&b, "      rule: \"Host(`%s`)\"\n", host)
	b.WriteString("      entryPoints: [web]\n")
	fmt.Fprintf(&b, "      service: %s\n", name)
	b.WriteString("  services:\n")
	fmt.Fprintf(&b, "    %s:\n      loadBalancer:\n        servers:\n          - url: \"%s\"\n", name, serviceURL)
	return b.String()
}
