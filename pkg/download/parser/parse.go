package parser

import (
	"net/url"

	"github.com/aifoundry-org/storage-manager/pkg/download"
	"github.com/aifoundry-org/storage-manager/pkg/download/http"
	"github.com/aifoundry-org/storage-manager/pkg/download/huggingface"
	"github.com/aifoundry-org/storage-manager/pkg/download/oci"
)

func Parse(source download.ContentSource) (download.Downloader, error) {
	u, err := url.Parse(source.URL)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "http", "https":
		return http.New(u, source.Credentials, source.CredentialsType)
	case "oci":
		return oci.New(u, source.Credentials, source.CredentialsType)
	case "hf", "huggingface":
		return huggingface.New(u, source.Credentials, source.CredentialsType)
	default:
		return nil, &download.ErrUnsupportedScheme{}
	}
}
