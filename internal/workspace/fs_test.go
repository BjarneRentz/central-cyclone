package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocalFSHelper_CreateFolderIfNotExists(t *testing.T) {
	fs := LocalFSHelper{}
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "test", "nested")

	// First creation should work
	if err := fs.CreateFolderIfNotExists(testPath); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Directory should exist
	if _, err := os.Stat(testPath); os.IsNotExist(err) {
		t.Error("Directory was not created")
	}

	// Second creation should not error
	if err := fs.CreateFolderIfNotExists(testPath); err != nil {
		t.Errorf("CreateFolderIfNotExists on existing dir failed: %v", err)
	}
}

func TestLocalFSHelper_ListFiles(t *testing.T) {
	fs := LocalFSHelper{}
	tempDir := t.TempDir()

	// Create test files and directories
	files := []string{"file1.txt", "file2.txt"}
	dirs := []string{"dir1", "dir2"}

	for _, f := range files {
		path := filepath.Join(tempDir, f)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	for _, d := range dirs {
		path := filepath.Join(tempDir, d)
		if err := os.Mkdir(path, 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}
	}

	// List and verify
	entries, err := fs.ListFiles(tempDir)
	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}

	if len(entries) != len(files)+len(dirs) {
		t.Errorf("Expected %d entries, got %d", len(files)+len(dirs), len(entries))
	}

	// Verify error on non-existent directory
	if _, err := fs.ListFiles(filepath.Join(tempDir, "nonexistent")); err == nil {
		t.Error("Expected error for non-existent directory")
	}
}

func TestLocalFSHelper_RemoveAll(t *testing.T) {
	fs := LocalFSHelper{}
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "test")

	// Create a directory with content
	if err := os.MkdirAll(filepath.Join(testPath, "nested"), 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testPath, "file.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Remove and verify
	if err := fs.RemoveAll(testPath); err != nil {
		t.Fatalf("RemoveAll failed: %v", err)
	}

	if _, err := os.Stat(testPath); !os.IsNotExist(err) {
		t.Error("Directory still exists after RemoveAll")
	}

	// Removing non-existent path should not error
	if err := fs.RemoveAll(testPath); err != nil {
		t.Errorf("RemoveAll on non-existent path failed: %v", err)
	}
}

func TestResolvePath(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		parts    []string
		expected string
	}{
		{
			name:     "single part",
			base:     "/base",
			parts:    []string{"folder"},
			expected: filepath.Join("/base", "folder"),
		},
		{
			name:     "multiple parts",
			base:     "/base",
			parts:    []string{"folder1", "folder2", "folder3"},
			expected: filepath.Join("/base", "folder1", "folder2", "folder3"),
		},
		{
			name:     "no parts",
			base:     "/base",
			parts:    []string{},
			expected: "/base",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolvePath(tt.base, tt.parts...)
			if result != tt.expected {
				t.Errorf("ResolvePath() = %v, want %v", result, tt.expected)
			}
		})
	}
}
