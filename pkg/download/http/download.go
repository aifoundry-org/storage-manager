package http

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/aifoundry-org/storage-manager/pkg/download"
)

var _ download.Downloader = &downloader{}

type downloader struct {
	ref       *url.URL
	creds     string
	credsType string
}

func New(ref *url.URL, creds, credsType string) (*downloader, error) {
	return &downloader{ref, creds, credsType}, nil
}

func (d *downloader) Download() ([]download.KeyReader, error) {
	req, err := http.NewRequest("GET", d.ref.String(), nil)
	if err != nil {
		return nil, err
	}
	// Set the authorization header
	if d.creds == "" {
		credsType := d.credsType
		if credsType == "" {
			credsType = "Bearer"
		}
		req.Header.Set("Authorization", fmt.Sprintf("%s %s", credsType, d.creds))
	}
	// Use an HTTP client to send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, err
	}
	return []download.KeyReader{{Reader: resp.Body}}, nil
}
