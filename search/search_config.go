package search

type SearchConfig struct {
	MaxWorkers int
}

func DefaultSearchConfig() *SearchConfig {
	return &SearchConfig{
		MaxWorkers: 8,
	}
}
