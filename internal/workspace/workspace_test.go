package workspace

import (
	"os"
	"testing"
)

func TestWorkspaceHandler_Clear_NonExistent_Throws(t *testing.T) {
	w := localWorkspace{path: "./nonexistentdir"}
	err := w.Clear()
	if err == nil {
		t.Error("expected error for non-existent directory, got nil")
	}
}

func TestWorkspaceHandler_Clear_Empty(t *testing.T) {
	dir := t.TempDir()
	w := localWorkspace{path: dir}
	if err := w.Clear(); err != nil {
		t.Errorf("unexpected error clearing empty dir: %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Errorf("failed to read dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty dir, found %d entries", len(entries))
	}
}
