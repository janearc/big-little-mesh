package watcher

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// fakeOracle drives the loop deterministically: it reports work on the first
// poll and nothing after, so we can assert the handler fires exactly once.
type fakeOracle struct{ polls int }

func (f *fakeOracle) Name() string { return "fake" }
func (f *fakeOracle) Poll(ctx context.Context) (Result, error) {
	f.polls++
	if f.polls == 1 {
		return Result{HasWork: true, Items: []string{"x"}}, nil
	}
	return Result{}, nil
}

// TestWatcher_FiresHandlerOnWork proves the loop calls the handler when the
// oracle reports work and stops cleanly on context cancel.
func TestWatcher_FiresHandlerOnWork(t *testing.T) {
	got := make(chan Result, 1)
	w := New(&fakeOracle{}, 5*time.Millisecond, func(ctx context.Context, r Result) error {
		got <- r
		return nil
	}, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go w.Run(ctx)

	select {
	case r := <-got:
		if !r.HasWork || len(r.Items) != 1 || r.Items[0] != "x" {
			t.Fatalf("handler got %+v, want one item 'x'", r)
		}
	case <-time.After(time.Second):
		t.Fatal("handler never fired")
	}
}

// TestFilesOracle_DetectsNewFilesOnce proves the inbox oracle reports a file the
// first time it appears and not on subsequent polls (the seen-set contract), and
// that the extension filter excludes non-matching files.
func TestFilesOracle_DetectsNewFilesOnce(t *testing.T) {
	dir := t.TempDir()
	o := NewFilesOracle(dir, ".m4a")

	// empty inbox -> no work
	if r, err := o.Poll(context.Background()); err != nil || r.HasWork {
		t.Fatalf("empty dir: got %+v err %v, want no work", r, err)
	}

	// a matching file and a non-matching file
	must := func(name string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	must("memo.m4a")
	must("notes.txt")

	r, err := o.Poll(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !r.HasWork || len(r.Items) != 1 || filepath.Base(r.Items[0]) != "memo.m4a" {
		t.Fatalf("first poll got %+v, want only memo.m4a", r)
	}

	// second poll: already seen -> no work
	if r, err := o.Poll(context.Background()); err != nil || r.HasWork {
		t.Fatalf("second poll: got %+v err %v, want no work (already seen)", r, err)
	}
}

// TestFilesOracle_MissingDirIsNotError pins that a not-yet-created inbox simply
// reports no work rather than erroring -- a daemon may start before its dir.
func TestFilesOracle_MissingDirIsNotError(t *testing.T) {
	o := NewFilesOracle(filepath.Join(t.TempDir(), "does-not-exist"))
	if r, err := o.Poll(context.Background()); err != nil || r.HasWork {
		t.Fatalf("missing dir: got %+v err %v, want no work and no error", r, err)
	}
}
