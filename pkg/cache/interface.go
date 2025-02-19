package cache

import (
	"io"
)

// Cache represents an implementation of a storage cache.
// Cache has a few concepts: blobs (or content), which are referenced via a unique key, and names (or tags),
// which are unique names that point to another name. Thus you can have 10 blobs each with its own key,
// and then have a name that points to one of those keys. This allows you to build a hierarchy of content.
type Cache interface {
	// These methods provide direct control over content
	// Get the content from the cache
	Get(key string) (io.ReadCloser, error)
	// Exists check if the content exists in the cache
	Exists(key string) (bool, error)
	// Delete the content from the cache
	Delete(key string) error
	// Put the content in the cache
	Put(key string, size int64, r io.ReadCloser) error

	// These methods provide control over aliases or names
	// Name alias a key to a name, will replace if already there
	Name(key, name string) error
	// Unname remove the alias from a key
	Unname(name string) error
	// Resolve a name to a key
	Resolve(name string) (string, error)

	// This method is used to clean up unreferenced keys
	GC() error
}
