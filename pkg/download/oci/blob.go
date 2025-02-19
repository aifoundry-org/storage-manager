package oci

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	dockermanifest "github.com/docker/distribution/manifest/manifestlist"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry/remote"
)

// pullBlobAndList pulls down a blob based on the provided descriptor.
// - if the descriptor is an index, it will parse the index and return a list of children descriptors.
// - if the descriptor is a manifest, it will parse the manifest return the list of config and layers.
func pullBlobAndList(desc ocispec.Descriptor, repo *remote.Repository) (io.ReadCloser, []ocispec.Descriptor, error) {
	var (
		rc       io.ReadCloser
		children []ocispec.Descriptor
	)
	ctx := context.Background()
	// get a readcloser for the content
	blobRC, err := repo.Fetch(ctx, desc)
	if err != nil {
		return nil, nil, fmt.Errorf("could not fetch content: %v", err)
	}

	// if it is an index or a manifest, parse it and get its children
	switch desc.MediaType {
	case dockermanifest.MediaTypeManifestList:
		b, err := io.ReadAll(blobRC)
		if err != nil {
			return nil, nil, fmt.Errorf("could not read content: %v", err)
		}
		var list dockermanifest.ManifestList
		if err := json.Unmarshal(b, &list); err != nil {
			return nil, nil, fmt.Errorf("could not unmarshal list: %v", err)
		}
		// add to the stuff to read
		rc = io.NopCloser(bytes.NewReader(b))
		manifests := list.Manifests
		for _, m := range manifests {
			b, err := json.Marshal(m)
			if err != nil {
				return nil, nil, fmt.Errorf("could not marshal manifest: %v", err)
			}
			var child ocispec.Descriptor
			if err := json.Unmarshal(b, &child); err != nil {
				return nil, nil, fmt.Errorf("could not unmarshal manifest: %v", err)
			}
			children = append(children, child)

		}
	case ocispec.MediaTypeImageIndex:
		b, err := io.ReadAll(blobRC)
		if err != nil {
			return nil, nil, fmt.Errorf("could not read content: %v", err)
		}
		var index ocispec.Index
		if err := json.Unmarshal(b, &index); err != nil {
			return nil, nil, fmt.Errorf("could not unmarshal index: %v", err)
		}
		// add to the stuff to read
		rc = io.NopCloser(bytes.NewReader(b))
		children = index.Manifests
	case ocispec.MediaTypeImageManifest:
		b, err := io.ReadAll(blobRC)
		if err != nil {
			return nil, nil, fmt.Errorf("could not read content: %v", err)
		}
		var image ocispec.Manifest
		if err := json.Unmarshal(b, &image); err != nil {
			return nil, nil, fmt.Errorf("could not unmarshal image: %v", err)
		}
		// add to the stuff to read
		rc = io.NopCloser(bytes.NewReader(b))
		children = image.Layers
		children = append(children, image.Config)
	default:
		rc = blobRC
	}
	return rc, children, nil
}
