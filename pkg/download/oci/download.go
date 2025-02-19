package oci

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/aifoundry-org/storage-manager/pkg/download"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"oras.land/oras-go/v2/registry/remote"
)

var _ download.Downloader = &downloader{}

type downloader struct {
	repo *remote.Repository
	ref  string
}

func New(ref *url.URL) (*downloader, error) {
	parts := strings.SplitN(ref.Path, ":", 2)
	repoName := parts[0]
	// trim leading / off repoName, just in case
	repoName = strings.TrimLeft(repoName, "/")
	refName := "latest"
	if len(parts) == 2 {
		refName = parts[1]
	}
	repo, err := remote.NewRepository(fmt.Sprintf("%s/%s", ref.Host, repoName))
	if err != nil {
		return nil, fmt.Errorf("could not create repository: %v", err)
	}
	return &downloader{repo: repo, ref: refName}, nil
}

func (d *downloader) Download() ([]download.KeyReader, error) {
	var readers []download.KeyReader
	ctx := context.Background()
	// resolve the tag to get the descriptor
	descriptor, err := d.repo.Resolve(ctx, d.ref)
	if err != nil {
		return nil, fmt.Errorf("could not resolve reference: %v", err)
	}

	// the root descriptor always is first
	children := []ocispec.Descriptor{descriptor}

	for len(children) != 0 {
		var newChildren []ocispec.Descriptor
		for _, desc := range children {
			rc, addChildren, err := pullBlobAndList(desc, d.repo)
			if err != nil {
				return nil, fmt.Errorf("could not pull blob and list: %v", err)
			}
			readers = append(readers, download.KeyReader{Key: desc.Digest.String(), Size: desc.Size, Reader: rc})
			newChildren = append(newChildren, addChildren...)
		}
		children = newChildren
	}
	return readers, nil
}
