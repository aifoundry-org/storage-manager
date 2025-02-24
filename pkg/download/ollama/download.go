package ollama

import (
	"fmt"
	"net/url"

	"github.com/aifoundry-org/storage-manager/pkg/download"

	"oras.land/oras-go/v2/registry/remote"
)

var _ download.Downloader = &downloader{}

type downloader struct {
	repo      *remote.Repository
	ref       string
	creds     string
	credsType string
}

func New(ref *url.URL, creds, credsType string) (*downloader, error) {
	return nil, fmt.Errorf("not implemented")
}

func (d *downloader) Download() ([]download.KeyReader, error) {
	return nil, nil
}
