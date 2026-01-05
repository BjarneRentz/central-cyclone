package upload

import (
	"bytes"
	"central-cyclone/internal/sbom"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type DependencyTrackUploader struct {
	serverURL string
	apiKey    string
}

func (uploader DependencyTrackUploader) UploadSBOM(sbom sbom.Sbom) error {
	url := uploader.serverURL + "/api/v1/bom"
	encodedSbom, err := getEncodedSbom(sbom)
	if err != nil {
		return err
	}

	req, err := createRequest(url, uploader.apiKey, sbom.ProjectId, encodedSbom)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload SBOM: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	slog.Info("⬆️  Uploaded SBOM successfully")

	return nil
}

func getEncodedSbom(sbom sbom.Sbom) (string, error) {
	encodedSbom := base64.StdEncoding.EncodeToString(sbom.Data)
	return encodedSbom, nil
}

func createRequest(url, apiKey, projectId, encodedSbom string) (*http.Request, error) {
	jsonBody := fmt.Appendf(nil, `{"project":"%s", "bom":"%s"}`, projectId, encodedSbom)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", apiKey)

	return req, nil
}
