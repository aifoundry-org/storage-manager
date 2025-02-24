package huggingface

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/aifoundry-org/storage-manager/pkg/download"
)

var _ download.Downloader = &downloader{}

type downloader struct {
	authToken string
	model     string
	file      string
}

func New(ref *url.URL, creds string) (*downloader, error) {
	// parse the name of the file and the model name from the URL
	file := path.Base(ref.Path)
	model := path.Dir(ref.Path)
	model = strings.TrimLeft(model, "/")
	return &downloader{model: model, file: file, authToken: creds}, nil
}

func (d *downloader) Info() (*RepoInfo, error) {
	u := fmt.Sprintf("https://huggingface.co/api/models/%s/revision/main?blobs=true", d.model)
	// get the info about the repo and its files
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	// Set the authorization header
	if d.authToken == "" {
		req.Header.Set("Authorization", "Bearer "+d.authToken)
	}
	// Use an HTTP client to send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download %s: %s", u, resp.Status)
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, resp.Body); err != nil {
		return nil, err
	}
	var info RepoInfo
	if err := json.Unmarshal(buf.Bytes(), &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func (d *downloader) Download() ([]download.KeyReader, error) {
	info, err := d.Info()
	if err != nil {
		return nil, err
	}
	// see if our file is in the info
	var (
		size int64
		key  string
	)
	for _, f := range info.Siblings {
		if f.Name == d.file {
			size = f.Size
			key = fmt.Sprintf("sha256:%s", f.LFS.Sha256)
		}
	}

	// get the info about the repo and its files
	u := fmt.Sprintf("https://huggingface.co/%s/resolve/%s/%s", d.model, info.CommitHash, d.file)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	// Set the authorization header
	if d.authToken == "" {
		req.Header.Set("Authorization", "Bearer "+d.authToken)
	}
	// Use an HTTP client to send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download %s: %s", u, resp.Status)
	}

	return []download.KeyReader{{Key: key, Size: size, Reader: resp.Body}}, nil
}
