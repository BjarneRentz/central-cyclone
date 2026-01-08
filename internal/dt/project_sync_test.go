package dt

import (
	"context"
	"testing"

	"central-cyclone/internal/config"

	dtrack "github.com/DependencyTrack/client-go"
)

// fakeClient is a test double implementing the Client interface.
type fakeClient struct {
	getProjectResp  dtrack.Project
	getProjectErr   error
	createdProjects []dtrack.Project
}

func (f *fakeClient) GetProject(ctx context.Context, name string, version string) (dtrack.Project, error) {
	return f.getProjectResp, f.getProjectErr
}

func (f *fakeClient) CreateProject(ctx context.Context, project dtrack.Project) (dtrack.Project, error) {
	f.createdProjects = append(f.createdProjects, project)
	// Simulate that the project was created and returned by the server.
	return project, nil
}

func TestSyncProjects_CreatesWhenProjectDoesNotExist(t *testing.T) {
	fc := &fakeClient{getProjectErr: &dtrack.APIError{StatusCode: 404}}
	ps := &ProjectSyncer{Client: fc}

	projects := []config.Project{{Name: "testapp", Version: "1.2.3"}}

	if err := ps.SyncProjects(context.Background(), projects); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fc.createdProjects) != 1 {
		t.Fatalf("expected 1 created project, got %d", len(fc.createdProjects))
	}

	if fc.createdProjects[0].Name != "testapp" {
		t.Fatalf("unexpected created project name: %s", fc.createdProjects[0].Name)
	}
	if fc.createdProjects[0].Version != "1.2.3" {
		t.Fatalf("unexpected created project version: %s", fc.createdProjects[0].Version)
	}
}

func TestSyncProjects_CreatesWhenLookupReturnsEmptyName(t *testing.T) {
	fc := &fakeClient{getProjectResp: dtrack.Project{}}
	ps := &ProjectSyncer{Client: fc}

	projects := []config.Project{{Name: "another", Version: "0.0.1"}}

	if err := ps.SyncProjects(context.Background(), projects); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fc.createdProjects) != 1 {
		t.Fatalf("expected 1 created project, got %d", len(fc.createdProjects))
	}

	if fc.createdProjects[0].Name != "another" {
		t.Fatalf("unexpected created project name: %s", fc.createdProjects[0].Name)
	}
	if fc.createdProjects[0].Version != "0.0.1" {
		t.Fatalf("unexpected created project version: %s", fc.createdProjects[0].Version)
	}
}

func TestSyncProjects_DoesNotCreateWhenProjectExists(t *testing.T) {
	fc := &fakeClient{getProjectResp: dtrack.Project{Name: "exist", Version: "1.0.0"}}
	ps := &ProjectSyncer{Client: fc}

	projects := []config.Project{{Name: "exist", Version: "1.0.0"}}

	if err := ps.SyncProjects(context.Background(), projects); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fc.createdProjects) != 0 {
		t.Fatalf("expected 0 created projects, got %d", len(fc.createdProjects))
	}
}
