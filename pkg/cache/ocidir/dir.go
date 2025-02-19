package ocidir

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/aifoundry-org/storage-manager/pkg/cache"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content/oci"
	oraserrdefs "oras.land/oras-go/v2/errdef"
)

type cacheOCIDir struct {
	dir   string
	cache *oci.Store
}

var _ cache.Cache = &cacheOCIDir{}

func New(cacheDir string) (*cacheOCIDir, error) {
	p, err := oci.New(cacheDir)
	if err != nil {
		return nil, fmt.Errorf("could not initialize cache at path %s: %v", cacheDir, err)
	}
	return &cacheOCIDir{
		cache: p,
		dir:   cacheDir,
	}, nil
}

// Get content from the cache
func (c *cacheOCIDir) Get(key string) (io.ReadCloser, error) {
	ctx := context.Background()
	desc, err := c.cache.Resolve(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("could not resolve %s: %v", key, err)
	}
	return c.cache.Fetch(ctx, desc)
}

// Exists check if content for a given key exists in the cache
func (c *cacheOCIDir) Exists(key string) (bool, error) {
	ctx := context.Background()
	desc, err := c.cache.Resolve(ctx, key)
	if err != nil && errors.Is(err, oraserrdefs.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("could not resolve %s: %v", key, err)
	}
	return c.cache.Exists(ctx, desc)
}

// Delete content from the cache
func (c *cacheOCIDir) Delete(key string) error {
	ctx := context.Background()
	desc, err := c.cache.Resolve(ctx, key)
	if err != nil && errors.Is(err, oraserrdefs.ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("could not resolve %s: %v", key, err)
	}
	defer c.cache.SaveIndex()
	exists, err := c.cache.Exists(ctx, desc)
	// if it does not exist, just remove the tag
	if (err != nil && errors.Is(err, oraserrdefs.ErrNotFound)) || !exists {
		return c.cache.Untag(ctx, key)
	}
	if err := c.cache.Delete(ctx, desc); err != nil {
		return fmt.Errorf("could not delete %s: %v", key, err)
	}
	return c.cache.Untag(ctx, key)
}

// Put content in the cache. If the key is not provided it will be generated from the content.
func (c *cacheOCIDir) Put(key string, size int64, r io.ReadCloser) error {
	ctx := context.Background()
	var desc ocispec.Descriptor
	if key == "" || size <= 0 {
		// we need a temporary directory to store the content
		tmpDir, err := os.MkdirTemp(c.dir, "content")
		if err != nil {
			return fmt.Errorf("could not create temporary directory: %v", err)
		}
		defer os.RemoveAll(tmpDir)
		digester := sha256.New()
		fileName := fmt.Sprintf("%s/%s", tmpDir, "content")
		f, err := os.Create(fileName)
		if err != nil {
			return fmt.Errorf("could not create temporary file: %v", err)
		}
		defer f.Close()
		multi := io.MultiWriter(digester, f)
		n, err := io.Copy(multi, r)
		if err != nil {
			return fmt.Errorf("could not copy content: %v", err)
		}
		if n == 0 {
			return fmt.Errorf("no content to copy")
		}
		newKey := digest.NewDigestFromEncoded(digest.SHA256, fmt.Sprintf("%x", digester.Sum(nil))).String()
		if newKey != key {
			return fmt.Errorf("key mismatch: %s != %s", newKey, key)
		}
		desc = ocispec.Descriptor{
			Digest: digest.Digest(key),
			Size:   n,
		}
		if _, err := f.Seek(0, 0); err != nil {
			return fmt.Errorf("could not seek to the beginning of the file: %v", err)
		}
		r = f
	} else {
		desc = ocispec.Descriptor{
			Digest: digest.Digest(key),
			Size:   size,
		}
		defer r.Close()
	}
	if err := c.cache.Push(ctx, desc, r); err != nil {
		return fmt.Errorf("could not put %s: %v", key, err)
	}
	defer c.cache.SaveIndex()
	return c.cache.Tag(ctx, desc, key)
}

// Name alias a key to a name
func (c *cacheOCIDir) Name(key, name string) error {
	ctx := context.Background()
	desc := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageConfig,
		Digest:    digest.Digest(key),
	}
	return c.cache.Tag(ctx, desc, name)
}

// Unname remove the alias from a key
func (c *cacheOCIDir) Unname(name string) error {
	ctx := context.Background()
	return c.cache.Untag(ctx, name)
}

// Resolve a name to a key
func (c *cacheOCIDir) Resolve(name string) (string, error) {
	ctx := context.Background()
	desc, err := c.cache.Resolve(ctx, name)
	if err != nil {
		return "", fmt.Errorf("could not resolve %s: %v", name, err)
	}
	return desc.Digest.String(), nil
}

// GC clean up unreferenced keys
func (c *cacheOCIDir) GC() error {
	ctx := context.Background()
	if err := c.cache.SaveIndex(); err != nil {
		return fmt.Errorf("could not save index: %v", err)
	}
	return c.cache.GC(ctx)
}
