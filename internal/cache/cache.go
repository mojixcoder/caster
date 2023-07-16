package cache

// Cache is the cache algorithm and can be implemented by various algorithms.
type Cache interface {
	// Get gets a key from cache.
	Get(key string) (any, error)

	// Set sets a key-value pair to the cache.
	Set(key string, val any) error

	// Flush flushes the cache.
	Flush() error
}
