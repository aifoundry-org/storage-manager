package http

import (
	"net/http"
	"net/url"

	"github.com/aifoundry-org/storage-manager/pkg/download"
)

var _ download.Downloader = &downloader{}

type downloader struct {
	ref   *url.URL
	creds string
}

func New(ref *url.URL, creds string) (*downloader, error) {
	return &downloader{ref, creds}, nil
}

func (d *downloader) Download() ([]download.KeyReader, error) {
	resp, err := http.Get(d.ref.String())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, err
	}
	return []download.KeyReader{{Reader: resp.Body}}, nil
}
