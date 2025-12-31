package config

// CacheMode defines the caching strategy used by the application.
type CacheMode string

const (
	// CacheDisabled indicates that caching is turned off.
	CacheDisabled CacheMode = "DISABLED"

	// CacheSync indicates that cache operations are performed synchronously.
	CacheSync CacheMode = "SYNC"

	// CacheAsync indicates that cache operations are performed asynchronously.
	CacheAsync CacheMode = "ASYNC"
)

// CacheProvider defines the backend used for caching.
type CacheProvider string

const (
	// CacheProviderJSONFile indicates that json-file cache is used.
	CacheProviderJSONFile CacheProvider = "JSON_FILE"
)

// CacheConfig holds configuration for caching.
type CacheConfig struct {
	Mode     CacheMode
	Provider CacheProvider
	Path     *string
}
