package sbom

type Sbom struct {
	ProjectId   string `json:"projectId"`
	ProjectType string `json:"projectType"`
	Data        string `json:"data"`
}
