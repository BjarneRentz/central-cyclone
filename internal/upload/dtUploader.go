package upload

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
)

type DependencyTrackUploader struct {
	serverURL string
	apiKey    string
}

func (uploader DependencyTrackUploader) UploadSBOM(sbomPath, projectId string) error {
	url := uploader.serverURL + "/api/v1/bom"
	encodedSbom, err := getEncodedSbom(sbomPath)
	if err != nil {
		return err
	}

	req, err := createRequest(url, uploader.apiKey, projectId, encodedSbom)
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
	return nil
}

func getEncodedSbom(sbomPath string) (string, error) {
	sbomData, err := os.ReadFile(sbomPath)
	if err != nil {
		return "", fmt.Errorf("failed to read SBOM file: %v", err)
	}
	encodedSbom := base64.StdEncoding.EncodeToString(sbomData)
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
