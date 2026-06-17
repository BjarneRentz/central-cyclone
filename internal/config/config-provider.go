package config

import (
	"central-cyclone/internal/analyzer"
	"fmt"
)

type ConfigProvider struct {
	settings           *Settings
	applicationRepoMap map[string]string
	applicationMap     map[string]*Application
	projectMap         map[string]*Project
}

func NewConfigProvider(settings *Settings) (*ConfigProvider, error) {
	provider := &ConfigProvider{
		settings:           settings,
		applicationRepoMap: make(map[string]string),
		applicationMap:     make(map[string]*Application),
		projectMap:         make(map[string]*Project),
	}

	// Validate and build lookup maps
	if err := provider.validateAndBuildMaps(); err != nil {
		return nil, err
	}

	return provider, nil
}

// validateAndBuildMaps validates configuration relationships and builds optimized lookup maps.
// Returns an error if any validation fails.
func (c *ConfigProvider) validateAndBuildMaps() error {
	// Build application repo map
	for _, appRepo := range c.settings.ApplicationRepos {
		for _, app := range appRepo.Applications {
			c.applicationRepoMap[app] = appRepo.RepoUrl
		}
	}

	// Build application map and project map
	for i := range c.settings.Applications {
		app := &c.settings.Applications[i]
		c.applicationMap[app.Name] = app
		for j := range app.Projects {
			project := &app.Projects[j]
			// Use format: appName:environment as key
			key := fmt.Sprintf("%s:%s", app.Name, project.Environment)
			c.projectMap[key] = project
		}
	}

	// Validate GitOps configuration
	for _, gitOpsRepo := range c.settings.GitOpsRepos {
		for _, gitOpsApp := range gitOpsRepo.GitOpsApplications {
			appName := gitOpsApp.ApplicationName

			// Check that application has a corresponding  entry for the gitops applicaton
			if _, exists := c.applicationMap[appName]; !exists {
				return fmt.Errorf("GitOps application '%s' has no corresponding Application entry in config", appName)
			}

			// Check that application has a corresponding ApplicationRepo entry
			if _, exists := c.applicationRepoMap[appName]; !exists {
				return fmt.Errorf("GitOps application '%s' has no corresponding ApplicationRepo entry in config", appName)
			}

			// Check that every environment has a matching Project
			for _, versionIdentifier := range gitOpsApp.VersionIdentifiers {
				key := fmt.Sprintf("%s:%s", appName, versionIdentifier.Environment)
				if _, exists := c.projectMap[key]; !exists {
					return fmt.Errorf(
						"GitOps application '%s' with environment '%s' has no matching Project in Application '%s' projects",
						appName, versionIdentifier.Environment, appName,
					)
				}
			}
		}
	}

	return nil
}

func (c *ConfigProvider) GetApplicationRepo(applicationName string) (string, error) {
	repoUrl, exists := c.applicationRepoMap[applicationName]
	if !exists {
		return "", fmt.Errorf("no repository found for application: %s", applicationName)
	}
	return repoUrl, nil
}

func (c *ConfigProvider) getApplication(applicationName string) *Application {
	return c.applicationMap[applicationName]
}

func (c *ConfigProvider) GetScanTargetForApplication(applicationName, env string) (*analyzer.ScanTarget, error) {
	applicationConfig := c.getApplication(applicationName)
	if applicationConfig == nil {
		return nil, fmt.Errorf("application config not found for application: %s", applicationName)
	}

	key := fmt.Sprintf("%s:%s", applicationName, env)
	project := c.projectMap[key]
	if project == nil {
		return nil, fmt.Errorf("project config not found for application: %s and environment: %s", applicationName, env)
	}

	return &analyzer.ScanTarget{
		ProjectId:   *project.ProjectId,
		ProjectType: applicationConfig.Type,
		Directory:   applicationConfig.RepoPath,
	}, nil
}

// GetGitOpsRefreshInterval returns the configured refresh interval in minutes.
// If not configured, it returns the default value of 10 minutes.
func (c *ConfigProvider) GetGitOpsRefreshInterval() int {
	if c.settings.GitOps.RefreshInterval == nil {
		return 10
	}
	return *c.settings.GitOps.RefreshInterval
}
