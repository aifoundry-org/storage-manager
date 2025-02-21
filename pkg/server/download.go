package server

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/opencontainers/go-digest"
)

type removeCloser struct {
	io.ReadCloser
	path string
}

func (r *removeCloser) Close() error {
	if err := r.ReadCloser.Close(); err != nil {
		return err
	}
	return os.RemoveAll(r.path)
}

func downloadAndHash(r io.ReadCloser) (key string, size int64, reader io.ReadCloser, err error) {
	dir, err := os.MkdirTemp("", "nekko-storage-manager-download")
	if err != nil {
		return key, size, nil, err
	}
	p := path.Join(dir, "download")
	f, err := os.Create(p)
	if err != nil {
		return key, size, nil, err
	}

	digester := sha256.New()
	multi := io.MultiWriter(digester, f)
	n, err := io.Copy(multi, r)
	if err != nil {
		return key, size, nil, err
	}
	if n == 0 {
		return key, size, nil, fmt.Errorf("no content to copy")
	}

	if _, err := f.Seek(0, 0); err != nil {
		return key, size, nil, fmt.Errorf("could not seek to the beginning of the file: %v", err)
	}
	key = digest.NewDigestFromEncoded(digest.SHA256, fmt.Sprintf("%x", digester.Sum(nil))).String()
	return key, size, &removeCloser{f, dir}, nil
}
