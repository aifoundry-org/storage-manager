package download

import (
	"io"
)

// KeyReader represents a key and a reader. If the key is blank, then the receiving system may decide
// how to calculate a unique key.
type KeyReader struct {
	Key    string
	Size   int64
	Reader io.ReadCloser
}
type Downloader interface {
	Download() ([]KeyReader, error)
}
