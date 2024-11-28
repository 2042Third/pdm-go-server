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

	// Count operations
	Incr(ctx context.Context, key string) (int64, error)
	IncrBy(ctx context.Context, key string, value int64) (int64, error)
	Decr(ctx context.Context, key string) (int64, error)
	DecrBy(ctx context.Context, key string, value int64) (int64, error)
	CountKeys(ctx context.Context, pattern string) (int64, error)

	// Key operations
	Exists(ctx context.Context, key string) (bool, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Map-like operations
	HSet(ctx context.Context, key, field, value string) error
	HGet(ctx context.Context, key, field string) (string, error)
	HDel(ctx context.Context, key string, fields ...string) error
	HExists(ctx context.Context, key, field string) (bool, error)
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	HLen(ctx context.Context, key string) (int64, error)

	// Set-like operations
	SAdd(ctx context.Context, key string, members ...string) error
	SRem(ctx context.Context, key string, members ...string) error
	SMembers(ctx context.Context, key string) ([]string, error)
	SIsMember(ctx context.Context, key, member string) (bool, error)
	SCard(ctx context.Context, key string) (int64, error)
	SetNX(ctx context.Context, key string, value string, ttl time.Duration) (bool, error)

	// List operations
	LPush(ctx context.Context, key string, values ...string) error
	RPush(ctx context.Context, key string, values ...string) error
	LPop(ctx context.Context, key string) (string, error)
	RPop(ctx context.Context, key string) (string, error)
	LLen(ctx context.Context, key string) (int64, error)
	LRange(ctx context.Context, key string, start, stop int64) ([]string, error)
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

func (c *DefaultRedisClient) CountKeys(ctx context.Context, pattern string) (int64, error) {
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return 0, err
	}
	return int64(len(keys)), nil
}

func (c *DefaultRedisClient) SetNX(ctx context.Context, key string, value string, ttl time.Duration) (bool, error) {
	success, err := c.client.SetNX(ctx, key, value, ttl).Result()
	if err != nil {
		return false, err
	}
	return success, nil
}

func (c *DefaultRedisClient) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

func (c *DefaultRedisClient) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.IncrBy(ctx, key, value).Result()
}

func (c *DefaultRedisClient) Decr(ctx context.Context, key string) (int64, error) {
	return c.client.Decr(ctx, key).Result()
}

func (c *DefaultRedisClient) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.DecrBy(ctx, key, value).Result()
}

func (c *DefaultRedisClient) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

func (c *DefaultRedisClient) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return c.client.Expire(ctx, key, ttl).Err()
}

func (c *DefaultRedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, key).Result()
}

func (c *DefaultRedisClient) HExists(ctx context.Context, key, field string) (bool, error) {
	result, err := c.client.HExists(ctx, key, field).Result()
	if err != nil {
		return false, err
	}
	return result, nil
}

func (c *DefaultRedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	result, err := c.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *DefaultRedisClient) HLen(ctx context.Context, key string) (int64, error) {
	result, err := c.client.HLen(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (c *DefaultRedisClient) SCard(ctx context.Context, key string) (int64, error) {
	return c.client.SCard(ctx, key).Result()
}

func (c *DefaultRedisClient) LPush(ctx context.Context, key string, values ...string) error {
	return c.client.LPush(ctx, key, values).Err()
}

func (c *DefaultRedisClient) RPush(ctx context.Context, key string, values ...string) error {
	return c.client.RPush(ctx, key, values).Err()
}

func (c *DefaultRedisClient) LPop(ctx context.Context, key string) (string, error) {
	val, err := c.client.LPop(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // Key not found
	}
	return val, err
}

func (c *DefaultRedisClient) RPop(ctx context.Context, key string) (string, error) {
	val, err := c.client.RPop(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // Key not found
	}
	return val, err
}

func (c *DefaultRedisClient) LLen(ctx context.Context, key string) (int64, error) {
	return c.client.LLen(ctx, key).Result()
}

func (c *DefaultRedisClient) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.client.LRange(ctx, key, start, stop).Result()
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
