package upload

import (
	"central-cyclone/internal/config"
	"central-cyclone/internal/sbom"
	"fmt"
	"os"
)

type Uploader interface {
	UploadSBOM(sbom sbom.Sbom) error
}

func CreateDependencyTrackUploader(settings *config.Settings) (Uploader, error) {
	apiKey := os.Getenv("DEPENDENCYTRACK_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("DEPENDENCYTRACK_API_KEY environment variable is not set")
	}

	return DependencyTrackUploader{
		serverURL: settings.DependencyTrack.Url,
		apiKey:    apiKey,
	}, nil

}
