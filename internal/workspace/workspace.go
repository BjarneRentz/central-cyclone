package workspace

import (
	"central-cyclone/internal/analyzer"
	"central-cyclone/internal/gittool"
	"fmt"
	"os"
	"path/filepath"
)

const (
	workspacePath = "workfolder"
	repoFolder    = "repos"
	sbomFolder    = "sboms"
)

type localWorkspace struct {
	path       string
	reposPath  string
	sbomsPath  string
	gitCloner  gittool.Cloner
	analyzer   analyzer.Analyzer
	fs         FSHelper
	namer      SBOMNamer
	repoMapper RepoURLMapper
}

type Workspace interface {
	Clear() error
	CloneRepoToWorkspace(repoUrl string) (string, error)
	AnalyzeRepoForTarget(repoUrl string, projectType string) (string, error)
}

func (w localWorkspace) CloneRepoToWorkspace(repoUrl string) (string, error) {
	folderName, err := w.repoMapper.GetFolderName(repoUrl)
	if err != nil {
		return "", fmt.Errorf("failed to get folder name from repo URL: %w", err)
	}
	targetDir := filepath.Join(w.reposPath, folderName)

	if err := w.fs.CreateFolderIfNotExists(targetDir); err != nil {
		return "", fmt.Errorf("failed to create target dir: %w", err)
	}

	err = w.gitCloner.CloneRepoToDir(repoUrl, targetDir)
	if err != nil {
		return "", fmt.Errorf("failed to clone repo: %w", err)
	}
	return targetDir, nil
}

func (w localWorkspace) Clear() error {
	if err := w.fs.RemoveAll(w.reposPath); err != nil {
		return fmt.Errorf("failed to clear repos directory: %w", err)
	}

	if err := w.fs.RemoveAll(w.sbomsPath); err != nil {
		return fmt.Errorf("failed to clear sboms directory: %w", err)
	}
	return nil
}

func (w localWorkspace) AnalyzeRepoForTarget(repoUrl string, projectType string) (string, error) {
	repoFolder, err := w.repoMapper.GetFolderName(repoUrl)
	if err != nil {
		return "", fmt.Errorf("failed to get folder name from repo URL: %w", err)
	}
	repoPath := filepath.Join(w.reposPath, repoFolder)

	sbomPath := w.namer.GenerateSBOMPath(w.sbomsPath, repoFolder, projectType)

	err = w.analyzer.AnalyzeProject(repoPath, projectType, sbomPath)
	if err != nil {
		return "", fmt.Errorf("failed to analyze project: %w", err)
	}

	return sbomPath, nil
}

func CreateLocalWorkspace() (Workspace, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}
	fullWorkFolderPath := filepath.Join(homeDir, ".central-cyclone", workspacePath)

	fs := LocalFSHelper{}
	if err := fs.CreateFolderIfNotExists(fullWorkFolderPath); err != nil {
		return nil, fmt.Errorf("could not create workfolder: %w", err)
	}

	fullReposPath := filepath.Join(fullWorkFolderPath, repoFolder)
	if err := fs.CreateFolderIfNotExists(fullReposPath); err != nil {
		return nil, fmt.Errorf("could not create repos folder: %w", err)
	}

	fullSbomsPath := filepath.Join(fullWorkFolderPath, sbomFolder)
	if err := fs.CreateFolderIfNotExists(fullSbomsPath); err != nil {
		return nil, fmt.Errorf("could not create sboms folder: %w", err)
	}
	return localWorkspace{
		path:       fullWorkFolderPath,
		reposPath:  fullReposPath,
		sbomsPath:  fullSbomsPath,
		gitCloner:  gittool.LocalGitCloner{},
		analyzer:   analyzer.CdxgenAnalyzer{},
		fs:         fs,
		namer:      DefaultSBOMNamer{},
		repoMapper: DefaultRepoMapper{},
	}, nil
}
