package store

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisClientInterface represents a go-redis/redis client
type RedisClientInterface interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	TTL(ctx context.Context, key string) *redis.DurationCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	Set(ctx context.Context, key string, values interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	FlushAll(ctx context.Context) *redis.StatusCmd
	SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd
	SMembers(ctx context.Context, key string) *redis.StringSliceCmd
}

const (
	// RedisType represents the storage type as a string value
	RedisType = "redis"
	// RedisTagPattern represents the tag pattern to be used as a key in specified storage
	RedisTagPattern = "gocache_tag_%s"
)

// RedisStore is a store for Redis
type RedisStore struct {
	client  RedisClientInterface
	options *Options
}

// NewRedis creates a new store to Redis instance(s)
func NewRedis(client RedisClientInterface, options *Options) *RedisStore {
	if options == nil {
		options = &Options{}
	}
	if options.Ctx == nil {
		options.Ctx = context.Background()
	}

	return &RedisStore{
		client:  client,
		options: options,
	}
}

// Get returns data stored from a given key
func (s *RedisStore) Get(key interface{}) (interface{}, error) {
	return s.client.Get(s.options.Ctx, key.(string)).Result()
}

// GetWithTTL returns data stored from a given key and its corresponding TTL
func (s *RedisStore) GetWithTTL(key interface{}) (interface{}, time.Duration, error) {
	object, err := s.client.Get(s.options.Ctx, key.(string)).Result()
	if err != nil {
		return nil, 0, err
	}

	ttl, err := s.client.TTL(s.options.Ctx, key.(string)).Result()
	if err != nil {
		return nil, 0, err
	}

	return object, ttl, err
}

// Set defines data in Redis for given key identifier
func (s *RedisStore) Set(key interface{}, value interface{}, options *Options) error {
	if options == nil {
		options = s.options
	}
	if options.Ctx == nil {
		options.Ctx = context.Background()
	}

	err := s.client.Set(options.Ctx, key.(string), value, options.ExpirationValue()).Err()
	if err != nil {
		return err
	}

	if tags := options.TagsValue(); len(tags) > 0 {
		s.setTags(key, tags)
	}

	return nil
}

func (s *RedisStore) setTags(key interface{}, tags []string) {
	for _, tag := range tags {
		tagKey := fmt.Sprintf(RedisTagPattern, tag)
		s.client.SAdd(s.options.Ctx, tagKey, key.(string))
		s.client.Expire(s.options.Ctx, tagKey, 720*time.Hour)
	}
}

// Delete removes data from Redis for given key identifier
func (s *RedisStore) Delete(key interface{}) error {
	_, err := s.client.Del(s.options.Ctx, key.(string)).Result()
	return err
}

// Invalidate invalidates some cache data in Redis for given options
func (s *RedisStore) Invalidate(options InvalidateOptions) error {
	if tags := options.TagsValue(); len(tags) > 0 {
		for _, tag := range tags {
			tagKey := fmt.Sprintf(RedisTagPattern, tag)
			cacheKeys, err := s.client.SMembers(s.options.Ctx, tagKey).Result()
			if err != nil {
				continue
			}

			for _, cacheKey := range cacheKeys {
				s.Delete(cacheKey)
			}

			s.Delete(tagKey)
		}
	}

	return nil
}

// GetType returns the store type
func (s *RedisStore) GetType() string {
	return RedisType
}

// Clear resets all data in the store
func (s *RedisStore) Clear() error {
	if err := s.client.FlushAll(s.options.Ctx).Err(); err != nil {
		return err
	}

	return nil
}
