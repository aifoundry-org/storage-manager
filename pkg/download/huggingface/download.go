package huggingface

import (
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/aifoundry-org/storage-manager/pkg/download"

	"github.com/gomlx/go-huggingface/hub"
)

var _ download.Downloader = &downloader{}

type downloader struct {
	repo *hub.Repo
	file string
}

func New(ref *url.URL) (*downloader, error) {
	// parse the name of the file and the model name from the URL
	file := path.Base(ref.Path)
	model := path.Base(path.Dir(ref.Path))
	repo := hub.New(model)
	if repo == nil {
		return nil, fmt.Errorf("invalid huggingface reference %s", ref.String())
	}
	return &downloader{repo: repo, file: file}, nil
}

func (d *downloader) Download() ([]download.KeyReader, error) {
	u, err := d.repo.FileURL(d.file)
	if err != nil {
		return nil, err
	}
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download %s: %s", u, resp.Status)
	}
	return []download.KeyReader{{Reader: resp.Body}}, nil
}
