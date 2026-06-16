package config

import (
	"central-cyclone/internal/analyzer"
	"fmt"
)

type ConfigProvider struct {
	settings           *Settings
	applicationRepoMap map[string]string
}

func NewConfigProvider(settings *Settings) *ConfigProvider {
	provider := &ConfigProvider{
		settings:           settings,
		applicationRepoMap: make(map[string]string),
	}

	provider.initAppRepoMap()

	return provider
}

func (c *ConfigProvider) initAppRepoMap() {

	for _, appRepo := range c.settings.ApplicationRepos {
		for _, app := range appRepo.Applications {
			c.applicationRepoMap[app] = appRepo.RepoUrl
		}
	}
}

func (c *ConfigProvider) GetApplicationRepo(applicationName string) (string, error) {
	repoUrl, exists := c.applicationRepoMap[applicationName]
	if !exists {
		return "", fmt.Errorf("no repository found for application: %s", applicationName)
	}
	return repoUrl, nil
}

func (c *ConfigProvider) getApplication(applicationName string) *Application {
	for _, app := range c.settings.Applications {
		if app.Name == applicationName {
			return &app
		}
	}
	return nil
}

func (c *ConfigProvider) GetScanTargetForApplication(applicationName, env string) (*analyzer.ScanTarget, error) {
	applicationConfig := c.getApplication(applicationName)
	if applicationConfig == nil {
		return nil, fmt.Errorf("application config not found for application: %s", applicationName)
	}

	for _, project := range applicationConfig.Projects {
		if project.Environment == env {
			return &analyzer.ScanTarget{
				ProjectId:   *project.ProjectId,
				ProjectType: applicationConfig.Type,
				Directory:   applicationConfig.RepoPath,
			}, nil
		}
	}
	return nil, fmt.Errorf("project config not found for application: %s and environment: %s", applicationName, env)
}

// GetGitOpsRefreshInterval returns the configured refresh interval in minutes.
// If not configured, it returns the default value of 10 minutes.
func (c *ConfigProvider) GetGitOpsRefreshInterval() int {
	if c.settings.GitOps.RefreshInterval == nil {
		return 10
	}
	return *c.settings.GitOps.RefreshInterval
}
