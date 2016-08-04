package client

// Cacheable represents an item that can be stored in a serialized cache
type Cacheable interface {
	// BuildCache builds the set of data to cache
	BuildCache() map[string]string

	// ApplyCache applies a set of cached data
	ApplyCache(c map[string]string)
}
