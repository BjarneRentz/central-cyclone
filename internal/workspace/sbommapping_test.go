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
		ProjectId:   "org_repo",
		ProjectType: "go",
	}
	got := namer.GenerateSBOMPath(sbomsDir, sbom)
	want := filepath.Join(sbomsDir, "sbom_org_repo.json")
	if got != want {
		t.Errorf("GenerateSBOMPath() = %q, want %q", got, want)
	}
}
