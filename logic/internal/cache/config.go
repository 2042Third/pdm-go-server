package cache

import "time"

type CacheConfig struct {
	Address  string        // Redis server address
	Password string        // Redis password
	DB       int           // Redis database number
	Timeout  time.Duration // Operation timeout
}
