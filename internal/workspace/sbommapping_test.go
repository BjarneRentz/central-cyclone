package workspace

import (
	"central-cyclone/internal/config"
	"path/filepath"
	"testing"
)

func TestDefaultSBOMNamer_GenerateSBOMPath(t *testing.T) {
	namer := DefaultSBOMNamer{}
	sbomsDir := "/path/to/sboms"
	got := namer.GenerateSBOMPath(sbomsDir, "org_repo", "go")
	want := filepath.Join(sbomsDir, "org_repo_sbom_go.json")
	if got != want {
		t.Errorf("GenerateSBOMPath() = %q, want %q", got, want)
	}
}

func TestSBOMNamer_MapSBOMToProject(t *testing.T) {
	settings := &config.Settings{
		Repositories: []config.Repo{
			{
				Url: "https://github.com/org/repo",
				Targets: []config.RepoTarget{
					{ProjectId: "proj-123", Type: "go"},
					{ProjectId: "proj-456", Type: "npm"},
				},
			},
		},
	}

	tests := []struct {
		name       string
		folderName string
		projType   string
		wantId     string
		wantFound  bool
	}{
		{
			name:       "matching repo and type",
			folderName: "org_repo",
			projType:   "go",
			wantId:     "proj-123",
			wantFound:  true,
		},
		{
			name:       "matching repo different type",
			folderName: "org_repo",
			projType:   "npm",
			wantId:     "proj-456",
			wantFound:  true,
		},
		{
			name:       "unknown repo",
			folderName: "unknown_repo",
			projType:   "go",
			wantId:     "",
			wantFound:  false,
		},
		{
			name:       "known repo unknown type",
			folderName: "org_repo",
			projType:   "unknown",
			wantId:     "",
			wantFound:  false,
		},
	}
	namer := DefaultSBOMNamer{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotId, gotFound := namer.MapSBOMToProject(settings, tt.folderName, tt.projType)
			if gotFound != tt.wantFound {
				t.Errorf("MapSBOMToProject() found = %v, want %v", gotFound, tt.wantFound)
			}
			if gotId != tt.wantId {
				t.Errorf("MapSBOMToProject() projectId = %q, want %q", gotId, tt.wantId)
			}
		})
	}

}
