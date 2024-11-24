package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisClient interface {
	// Basic operations
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error

	// Map-like operations
	HSet(ctx context.Context, key, field, value string) error
	HGet(ctx context.Context, key, field string) (string, error)
	HDel(ctx context.Context, key string, fields ...string) error

	// Set-like operations
	SAdd(ctx context.Context, key string, members ...string) error
	SRem(ctx context.Context, key string, members ...string) error
	SMembers(ctx context.Context, key string) ([]string, error)
	SIsMember(ctx context.Context, key, member string) (bool, error)
}

type DefaultRedisClient struct {
	client *redis.Client
}

func NewRedisClient(cfg *CacheConfig) RedisClient {
	options := &redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	}
	return &DefaultRedisClient{
		client: redis.NewClient(options),
	}
}

// Basic operations
func (c *DefaultRedisClient) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // Key not found
	}
	return val, err
}

func (c *DefaultRedisClient) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *DefaultRedisClient) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// Map-like operations
func (c *DefaultRedisClient) HSet(ctx context.Context, key, field, value string) error {
	return c.client.HSet(ctx, key, field, value).Err()
}

func (c *DefaultRedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	val, err := c.client.HGet(ctx, key, field).Result()
	if err == redis.Nil {
		return "", nil // Field not found
	}
	return val, err
}

func (c *DefaultRedisClient) HDel(ctx context.Context, key string, fields ...string) error {
	return c.client.HDel(ctx, key, fields...).Err()
}

// Set-like operations
func (c *DefaultRedisClient) SAdd(ctx context.Context, key string, members ...string) error {
	// Convert []string to []interface{}
	args := make([]interface{}, len(members))
	for i, member := range members {
		args[i] = member
	}
	return c.client.SAdd(ctx, key, args...).Err()
}

func (c *DefaultRedisClient) SRem(ctx context.Context, key string, members ...string) error {
	// Convert []string to []interface{}
	args := make([]interface{}, len(members))
	for i, member := range members {
		args[i] = member
	}
	return c.client.SRem(ctx, key, args...).Err()
}

func (c *DefaultRedisClient) SMembers(ctx context.Context, key string) ([]string, error) {
	return c.client.SMembers(ctx, key).Result()
}

func (c *DefaultRedisClient) SIsMember(ctx context.Context, key, member string) (bool, error) {
	return c.client.SIsMember(ctx, key, member).Result()
}
