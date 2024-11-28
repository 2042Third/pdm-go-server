package cache

import (
	"context"
	"encoding/json"
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

	GetJSON(ctx context.Context, key string, dest interface{}) error
	SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	GetWithDefault(ctx context.Context, key string, defaultValue string) (string, error)
	SetNX(ctx context.Context, key string, value string, ttl time.Duration) (bool, error)
	IncrWithReset(ctx context.Context, key string, resetAfter time.Duration) (int64, error)
	Exists(ctx context.Context, key string) (interface{}, interface{})
	Incr(ctx context.Context, key string) (interface{}, interface{})
	Expire(ctx context.Context, key string, after time.Duration) interface{}
	CountKeys(ctx context.Context, pattern string) (int64, error)
}

type RedisCache struct {
	client RedisClient
}

func NewCache(client RedisClient) *RedisCache {
	return &RedisCache{
		client: client,
	}
}

// Additional helper methods
func (r *RedisCache) CountKeys(ctx context.Context, pattern string) (int64, error) {
	return r.client.CountKeys(ctx, pattern)
}

func (r *RedisCache) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := r.Get(ctx, key)
	if err != nil {
		return err
	}
	if data == "" {
		return nil
	}
	return json.Unmarshal([]byte(data), dest)
}

func (r *RedisCache) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.Set(ctx, key, string(data), ttl)
}

func (r *RedisCache) GetWithDefault(ctx context.Context, key string, defaultValue string) (string, error) {
	value, err := r.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if value == "" {
		return defaultValue, nil
	}
	return value, nil
}

func (r *RedisCache) SetNX(ctx context.Context, key string, value string, ttl time.Duration) (bool, error) {
	success, err := r.client.SetNX(ctx, key, value, ttl)
	if err != nil {
		return false, err
	}
	return success, nil
}

func (r *RedisCache) IncrWithReset(ctx context.Context, key string, resetAfter time.Duration) (int64, error) {
	count, err := r.client.Incr(ctx, key)
	if err != nil {
		return 0, err
	}

	if count == 1 {
		err = r.client.Expire(ctx, key, resetAfter)
		if err != nil {
			return count, err
		}
	}

	return count, nil
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

func (r *RedisCache) Exists(ctx context.Context, key string) (interface{}, interface{}) {
	return r.client.Exists(ctx, key)
}

func (r *RedisCache) Incr(ctx context.Context, key string) (interface{}, interface{}) {
	return r.client.Incr(ctx, key)
}

func (r *RedisCache) Expire(ctx context.Context, key string, after time.Duration) interface{} {
	return r.client.Expire(ctx, key, after)
}
