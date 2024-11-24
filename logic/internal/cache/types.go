package cache

type CacheEntry struct {
	Key   string
	Value string
	TTL   int64 // Time-to-live in seconds
}
