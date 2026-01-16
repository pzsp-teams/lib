package search

// SearchConfig configures how search operations are executed.
//
// MaxWorkers defines the maximum number of concurrent workers used to fetch
// and enrich search results. Higher values can speed up searches but increase
// resource usage and pressure on the upstream API.
type SearchConfig struct {
	MaxWorkers int
}

// DefaultSearchConfig returns a SearchConfig initialized with sane defaults.
func DefaultSearchConfig() *SearchConfig {
	return &SearchConfig{
		MaxWorkers: 8,
	}
}
