package cache

import (
	"context"
	"time"
)

type Cache interface {
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

type RedisCache struct {
	client RedisClient
}

func NewCache(client RedisClient) Cache {
	return &RedisCache{
		client: client,
	}
}

// Basic operations
func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key)
}

func (r *RedisCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl)
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	return r.client.Delete(ctx, key)
}

// Map-like operations
func (r *RedisCache) HSet(ctx context.Context, key, field, value string) error {
	return r.client.HSet(ctx, key, field, value)
}

func (r *RedisCache) HGet(ctx context.Context, key, field string) (string, error) {
	return r.client.HGet(ctx, key, field)
}

func (r *RedisCache) HDel(ctx context.Context, key string, fields ...string) error {
	return r.client.HDel(ctx, key, fields...)
}

// Set-like operations
func (r *RedisCache) SAdd(ctx context.Context, key string, members ...string) error {
	return r.client.SAdd(ctx, key, members...)
}

func (r *RedisCache) SRem(ctx context.Context, key string, members ...string) error {
	return r.client.SRem(ctx, key, members...)
}

func (r *RedisCache) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.client.SMembers(ctx, key)
}

func (r *RedisCache) SIsMember(ctx context.Context, key, member string) (bool, error) {
	return r.client.SIsMember(ctx, key, member)
}
