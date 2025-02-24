package huggingface

type RepoInfo struct {
	ID         string      `json:"id"`
	ModelID    string      `json:"model_id"`
	Author     string      `json:"author"`
	CommitHash string      `json:"sha"`
	Tags       []string    `json:"tags"`
	Siblings   []*FileInfo `json:"siblings"`
}

// FileInfo represents one of the model file, in the Info structure.
type FileInfo struct {
	Name   string `json:"rfilename"`
	BlobID string `json:"blobId,omitempty"`
	Size   int64  `json:"size,omitempty"`
	LFS    LFS
}

type LFS struct {
	Sha256 string `json:"sha256,omitempty"`
	Size   int64  `json:"size,omitempty"`
}
