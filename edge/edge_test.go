package edge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRegisterRoute(t *testing.T) {
	dir := t.TempDir()
	if err := RegisterRoute(dir, "magpie", "magpie.example.com", "http://host.k3d.internal:8092"); err != nil {
		t.Fatalf("RegisterRoute: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "magpie.yml"))
	if err != nil {
		t.Fatalf("route file not written: %v", err)
	}
	s := string(got)
	for _, want := range []string{
		"Host(`magpie.example.com`)",
		"http://host.k3d.internal:8092",
		"entryPoints: [web]",
		"service: magpie",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("route YAML missing %q; got:\n%s", want, s)
		}
	}
}

func TestRegisterRoute_RequiresArgs(t *testing.T) {
	if err := RegisterRoute("", "n", "h", "u"); err == nil {
		t.Error("expected error for empty dynamicDir")
	}
	if err := RegisterRoute(t.TempDir(), "n", "", "u"); err == nil {
		t.Error("expected error for empty host")
	}
}
