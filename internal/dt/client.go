package dt

import (
	"central-cyclone/internal/config"
	"context"
	"fmt"
	"os"

	dtrack "github.com/DependencyTrack/client-go"
)

type DTrackClient struct {
	client *dtrack.Client
}

func NewDTrackClient(dtrackConfig *config.DependencyTrackConfig) (*DTrackClient, error) {
	apiKey := os.Getenv("DEPENDENCYTRACK_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("DEPENDENCYTRACK_API_KEY environment variable is not set")
	}
	client, err := dtrack.NewClient(dtrackConfig.Url, dtrack.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	return &DTrackClient{client: client}, nil
}

func (client *DTrackClient) CreateProject(ctx context.Context, project dtrack.Project) (dtrack.Project, error) {
	createdProject, err := client.client.Project.Create(ctx, project)

	if err != nil {
		return dtrack.Project{}, fmt.Errorf("Could not create DependencyTrack project: %w", err)
	}
	return createdProject, nil
}

func (client *DTrackClient) GetProject(ctx context.Context, name string, version string) (dtrack.Project, error) {
	project, err := client.client.Project.Lookup(ctx, name, version)

	if err != nil {
		return dtrack.Project{}, fmt.Errorf("Could not lookup project: %w", err)
	}
	return project, nil
}
