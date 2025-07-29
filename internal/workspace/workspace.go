package workspace

import (
	"fmt"
	"os"
	"path/filepath"
)

const workspacePath = "workfolder"

type workspaceStruct struct {
	path string
}

// Removes all files and folders in the workspace directory
func (w workspaceStruct) Clear() error {
	entries, err := os.ReadDir(w.path)
	if err != nil {
		return fmt.Errorf("failed to read workspace directory: %w", err)
	}
	for _, entry := range entries {
		entryPath := filepath.Join(w.path, entry.Name())
		if err := os.RemoveAll(entryPath); err != nil {
			return fmt.Errorf("failed to remove '%s': %w", entryPath, err)
		}
	}
	return nil
}

type WorkspaceHandler interface {
	Clear() error
}

func CreateWorkspace() (WorkspaceHandler, error) {
	// Get the executable path
	ex, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}
	execDir := filepath.Dir(ex)

	// Construct the full path for the work folder
	fullWorkFolderPath := filepath.Join(execDir, workspacePath)

	if _, err := os.Stat(fullWorkFolderPath); os.IsNotExist(err) {
		fmt.Printf("Creating work directory: %s\n", fullWorkFolderPath)
		if err := os.MkdirAll(fullWorkFolderPath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create work folder '%s': %w", fullWorkFolderPath, err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to check work folder '%s': %w", fullWorkFolderPath, err)
	} else {
		fmt.Printf("Work directory '%s' already exists.\n", fullWorkFolderPath)
	}

	return workspaceStruct{
		path: fullWorkFolderPath,
	}, nil
}
