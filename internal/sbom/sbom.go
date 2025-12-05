package sbom

type Sbom struct {
	ProjectId         string
	ProjectType       string
	ProjectFolderName string
	Data              []byte
}
