package cache

import "errors"

var (
	ErrCacheMiss = errors.New("cache miss")
	ErrCacheSet  = errors.New("failed to set cache")
	ErrCacheDel  = errors.New("failed to delete cache")
)
