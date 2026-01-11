package workspace

import (
	"central-cyclone/internal/models"
	"path/filepath"
	"testing"
)

func TestDefaultSBOMNamer_GenerateSBOMPath(t *testing.T) {
	namer := DefaultSBOMNamer{}
	sbomsDir := "/path/to/sboms"
	sbom := models.Sbom{
		ProjectFolderName: "org_repo",
		ProjectType:       "go",
	}
	got := namer.GenerateSBOMPath(sbomsDir, sbom)
	want := filepath.Join(sbomsDir, "org_repo_sbom_go.json")
	if got != want {
		t.Errorf("GenerateSBOMPath() = %q, want %q", got, want)
	}
}

func TestDefaultSBOMNamer_ParseFilename(t *testing.T) {
	tests := []struct {
		name            string
		filename        string
		wantFolderName  string
		wantProjectType string
		wantErr         bool
	}{
		{
			name:            "valid filename with go project",
			filename:        "org_repo_sbom_go.json",
			wantFolderName:  "org_repo",
			wantProjectType: "go",
			wantErr:         false,
		},
		{
			name:            "valid filename with npm project",
			filename:        "my_project_sbom_npm.json",
			wantFolderName:  "my_project",
			wantProjectType: "npm",
			wantErr:         false,
		},
		{
			name:            "valid filename with python project",
			filename:        "data_science_sbom_python.json",
			wantFolderName:  "data_science",
			wantProjectType: "python",
			wantErr:         false,
		},
		{
			name:            "filename with underscore in folder name",
			filename:        "my_org_my_repo_sbom_java.json",
			wantFolderName:  "my_org_my_repo",
			wantProjectType: "java",
			wantErr:         false,
		},
		{
			name:            "missing .json extension",
			filename:        "org_repo_sbom_go",
			wantFolderName:  "org_repo",
			wantProjectType: "go",
			wantErr:         false,
		},
		{
			name:            "invalid format - missing _sbom_ separator",
			filename:        "org_repo_go.json",
			wantFolderName:  "",
			wantProjectType: "",
			wantErr:         true,
		},
		{
			name:            "invalid format - multiple _sbom_ separators",
			filename:        "org_repo_sbom_type_sbom_extra.json",
			wantFolderName:  "",
			wantProjectType: "",
			wantErr:         true,
		},
		{
			name:            "invalid format - only filename with no structure",
			filename:        "random.json",
			wantFolderName:  "",
			wantProjectType: "",
			wantErr:         true,
		},
		{
			name:            "empty filename",
			filename:        "",
			wantFolderName:  "",
			wantProjectType: "",
			wantErr:         true,
		},
	}

	namer := DefaultSBOMNamer{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFolderName, gotProjectType, err := namer.ParseFilename(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFilename() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotFolderName != tt.wantFolderName {
				t.Errorf("ParseFilename() folderName = %q, want %q", gotFolderName, tt.wantFolderName)
			}
			if gotProjectType != tt.wantProjectType {
				t.Errorf("ParseFilename() projectType = %q, want %q", gotProjectType, tt.wantProjectType)
			}
		})
	}
}
