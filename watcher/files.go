package watcher

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// FilesOracle reports files that have appeared in a directory and not yet been
// seen. It is the "new inputs in an inbox" oracle -- magpie's m4a watch is
// exactly this. Detection only; the handler does the actual processing and I/O.
//
// "Seen" is tracked in memory. On restart the set is empty, so already-present
// files report as new again -- consumers must make processing idempotent (the
// bento model's job: a file whose output already exists is a no-op). Persisting
// the seen set is a deliberate later step, noted rather than half-built.
type FilesOracle struct {
	dir  string
	exts map[string]bool // lowercased, with leading dot; empty = match all

	mu   sync.Mutex
	seen map[string]bool
}

// NewFilesOracle watches dir for new files. exts filters by extension (e.g.
// ".m4a", ".wav"); pass none to match every file. Existing files are NOT marked
// seen at construction -- the first poll reports them, which is what you want
// for an inbox that may already hold work when the daemon starts.
func NewFilesOracle(dir string, exts ...string) *FilesOracle {
	set := map[string]bool{}
	for _, e := range exts {
		if e == "" {
			continue
		}
		if !strings.HasPrefix(e, ".") {
			e = "." + e
		}
		set[strings.ToLower(e)] = true
	}
	return &FilesOracle{dir: dir, exts: set, seen: map[string]bool{}}
}

// Name implements Oracle.
func (o *FilesOracle) Name() string { return "files:" + o.dir }

// Poll lists the directory (non-recursive), returns any matching files not seen
// before, and records them. A missing directory is not an error -- an inbox that
// does not exist yet simply has no work; it reports no work until it appears.
func (o *FilesOracle) Poll(ctx context.Context) (Result, error) {
	entries, err := os.ReadDir(o.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return Result{}, nil
		}
		return Result{}, err
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	var fresh []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if len(o.exts) > 0 && !o.exts[strings.ToLower(filepath.Ext(name))] {
			continue
		}
		full := filepath.Join(o.dir, name)
		if o.seen[full] {
			continue
		}
		o.seen[full] = true
		fresh = append(fresh, full)
	}
	sort.Strings(fresh) // deterministic order for the handler and tests
	return Result{HasWork: len(fresh) > 0, Items: fresh}, nil
}
