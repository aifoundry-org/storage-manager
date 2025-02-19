package parser

import (
	"net/url"

	"github.com/aifoundry-org/storage-manager/pkg/download"
	"github.com/aifoundry-org/storage-manager/pkg/download/http"
	"github.com/aifoundry-org/storage-manager/pkg/download/huggingface"
	"github.com/aifoundry-org/storage-manager/pkg/download/oci"
)

func Parse(source string) (download.Downloader, error) {
	u, err := url.Parse(source)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "http", "https":
		return http.New(u)
	case "oci":
		return oci.New(u)
	case "hf", "huggingface":
		return huggingface.New(u)
	default:
		return nil, &download.ErrUnsupportedScheme{}
	}
}
