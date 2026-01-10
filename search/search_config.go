package search

// SearchConfig holds configuration for search operations.
//
// MaxWorkers specifies the maximum number of concurrent workers to use
// when fetching search results. A higher number can speed up searches but
// may increase resource usage. Default is 8.
type SearchConfig struct {
	MaxWorkers int
}

// DefaultSearchConfig returns a SearchConfig with default settings.
func DefaultSearchConfig() *SearchConfig {
	return &SearchConfig{
		MaxWorkers: 8,
	}
}
