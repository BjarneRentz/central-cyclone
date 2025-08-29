package upload

type Uploader interface {
	UploadSBOM(sbomPath, projectId string) error
}
