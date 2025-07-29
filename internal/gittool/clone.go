package gittool

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func CloneRepo(repoURL string) (string, error) {

	ex, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}
	execDir := filepath.Dir(ex)

	// 2. Construct the full path for the work folder
	fullWorkFolderPath := filepath.Join(execDir, "workfolder")

	if _, err := os.Stat(fullWorkFolderPath); os.IsNotExist(err) {
		fmt.Printf("Creating work directory: %s\n", fullWorkFolderPath)
		if err := os.MkdirAll(fullWorkFolderPath, 0755); err != nil {
			return "", fmt.Errorf("failed to create work folder '%s': %w", fullWorkFolderPath, err)
		}
	} else if err != nil {
		return "", fmt.Errorf("failed to check work folder '%s': %w", fullWorkFolderPath, err)
	} else {
		fmt.Printf("Work directory '%s' already exists.\n", fullWorkFolderPath)

	}

	// Git will clone directly into this directory (because we use "." as the destination)
	originalDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}
	if err := os.Chdir(fullWorkFolderPath); err != nil {
		return "", fmt.Errorf("failed to change into work folder '%s': %w", fullWorkFolderPath, err)
	}

	// Ensure we change back to the original directory when the function exits
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			log.Printf("Warning: Failed to change back to original directory: %v", err)
		}
	}()

	// The "." means clone into the current directory, which is now `fullWorkFolderPath`
	cmd := exec.Command("git", "clone", repoURL, ".")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Executing command: %s %s (in directory: %s)\n", cmd.Path, cmd.Args, fullWorkFolderPath)
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("git clone failed: %w", err)
	}

	fmt.Printf("Successfully cloned repository to: %s\n", fullWorkFolderPath)
	return fullWorkFolderPath, nil
}
