package dt

import (
	"central-cyclone/internal/config"
	"context"
	"errors"
	"log/slog"

	dtrack "github.com/DependencyTrack/client-go"
)

type ProjectSyncer struct {
	Client Client
}

func (ps *ProjectSyncer) SyncProjects(ctx context.Context, projects []config.Project) error {

	for _, proj := range projects {
		maybeProject, err := ps.Client.GetProject(ctx, proj.Name, proj.Version)
		if err != nil {
			var apiErr *dtrack.APIError
			if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
				// Project not found in Dependency-Track -> create it
				newProject := dtrack.Project{
					Name:     proj.Name,
					Version:  proj.Version,
					Active:   true,
					IsLatest: &proj.IsLatest,
				}
				createdProject, err := ps.Client.CreateProject(ctx, newProject)
				if err != nil {
					slog.Error("Could not create project in Dependency-Track", "project-name", proj.Name, "version", proj.Version, "error", err)
					continue
				}
				slog.Info("Created new project in Dependency-Track", "project-name", createdProject.Name, "version", createdProject.Version)
				continue
			}

			slog.Warn("Could not perform project lookup", "project-name", proj.Name, "version", proj.Version, "error", err)
			continue
		}

		if maybeProject.Name == "" {
			newProject := dtrack.Project{
				Name:    proj.Name,
				Version: proj.Version,
			}
			createdProject, err := ps.Client.CreateProject(ctx, newProject)
			if err != nil {
				slog.Error("Could not create project in Dependency-Track", "project-name", proj.Name, "version", proj.Version, "error", err)
				continue
			}
			slog.Info("Created new project in Dependency-Track", "project-name", createdProject.Name, "version", createdProject.Version)

		} else {
			slog.Info("Project already exists in Dependency-Track", "project-name", proj.Name, "version", proj.Version)
		}
	}
	return nil

}
