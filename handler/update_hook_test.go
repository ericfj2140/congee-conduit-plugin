package handler

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunUpdateHookAppliesSchema(t *testing.T) {
	dir := t.TempDir()
	if err := RunUpdateHook(t.Context(), dir, `{"index_backend":"turso"}`); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "conduit-index.db")); err != nil {
		t.Fatalf("expected index db: %v", err)
	}
}
