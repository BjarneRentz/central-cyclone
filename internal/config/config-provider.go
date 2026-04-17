package config

import "fmt"

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
