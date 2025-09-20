package redis
// Package redis provides high-performance Redis cache implementation
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/RichardKnop/go-oauth2-server/storage"
	"github.com/go-redis/redis/v8"
)

// RedisCache implements high-performance Redis caching
type RedisCache struct {
	client  redis.UniversalClient
	metrics storage.MetricsProvider
	prefix  string
}

// RedisConfig defines Redis-specific configuration
type RedisConfig struct {
	// Single Redis instance
	Addr     string `json:"addr"`
	Password string `json:"password"`
	DB       int    `json:"db"`

	// Redis Cluster
	Addrs []string `json:"addrs,omitempty"`

	// Performance settings
	PoolSize     int           `json:"pool_size"`
	MinIdleConns int           `json:"min_idle_conns"`
	MaxRetries   int           `json:"max_retries"`
	DialTimeout  time.Duration `json:"dial_timeout"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`

	// Cache settings
	KeyPrefix string `json:"key_prefix"`
}

// NewRedisCache creates a new high-performance Redis cache
func NewRedisCache(config *RedisConfig, metrics storage.MetricsProvider) (*RedisCache, error) {
	var client redis.UniversalClient

	if len(config.Addrs) > 0 {
		// Redis Cluster configuration for high availability
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        config.Addrs,
			Password:     config.Password,
			PoolSize:     config.PoolSize,
			MinIdleConns: config.MinIdleConns,
			MaxRetries:   config.MaxRetries,
			DialTimeout:  config.DialTimeout,
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
		})
	} else {
		// Single Redis instance
		client = redis.NewClient(&redis.Options{
			Addr:         config.Addr,
			Password:     config.Password,
			DB:           config.DB,
			PoolSize:     config.PoolSize,
			MinIdleConns: config.MinIdleConns,
			MaxRetries:   config.MaxRetries,
			DialTimeout:  config.DialTimeout,
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
		})
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client:  client,
		metrics: metrics,
		prefix:  config.KeyPrefix,
	}, nil
}

// Set stores a value in cache with TTL
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	start := time.Now()
	defer func() {
		r.metrics.RecordCacheOperation("set", true, time.Since(start))
	}()

	fullKey := r.getFullKey(key)
	
	// Serialize the value
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	// Store in Redis
	if err := r.client.Set(ctx, fullKey, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache value: %w", err)
	}

	return nil
}

// Get retrieves a value from cache
func (r *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	start := time.Now()
	
	fullKey := r.getFullKey(key)
	
	// Get from Redis
	data, err := r.client.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			r.metrics.RecordCacheOperation("get", false, time.Since(start))
			return fmt.Errorf("cache miss")
		}
		return fmt.Errorf("failed to get cache value: %w", err)
	}

	// Deserialize the value
	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	r.metrics.RecordCacheOperation("get", true, time.Since(start))
	return nil
}

// Delete removes a value from cache
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	start := time.Now()
	defer func() {
		r.metrics.RecordCacheOperation("delete", true, time.Since(start))
	}()

	fullKey := r.getFullKey(key)
	
	if err := r.client.Del(ctx, fullKey).Err(); err != nil {
		return fmt.Errorf("failed to delete cache value: %w", err)
	}

	return nil
}

// SetMulti stores multiple values in cache with TTL
func (r *RedisCache) SetMulti(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	start := time.Now()
	defer func() {
		r.metrics.RecordCacheOperation("set_multi", true, time.Since(start))
	}()

	// Use pipeline for batch operations
	pipe := r.client.Pipeline()
	
	for key, value := range items {
		fullKey := r.getFullKey(key)
		
		// Serialize the value
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
		}
		
		pipe.Set(ctx, fullKey, data, ttl)
	}
	
	// Execute pipeline
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to execute set multi: %w", err)
	}

	return nil
}

// GetMulti retrieves multiple values from cache
func (r *RedisCache) GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error) {
	start := time.Now()
	defer func() {
		r.metrics.RecordCacheOperation("get_multi", true, time.Since(start))
	}()

	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	// Prepare full keys
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = r.getFullKey(key)
	}

	// Use pipeline for batch operations
	pipe := r.client.Pipeline()
	commands := make([]*redis.StringCmd, len(fullKeys))
	
	for i, fullKey := range fullKeys {
		commands[i] = pipe.Get(ctx, fullKey)
	}
	
	// Execute pipeline
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to execute get multi: %w", err)
	}

	// Process results
	result := make(map[string]interface{})
	for i, cmd := range commands {
		if cmd.Err() == nil {
			var value interface{}
			if err := json.Unmarshal([]byte(cmd.Val()), &value); err == nil {
				result[keys[i]] = value
			}
		}
	}

	return result, nil
}

// DeleteMulti removes multiple values from cache
func (r *RedisCache) DeleteMulti(ctx context.Context, keys []string) error {
	start := time.Now()
	defer func() {
		r.metrics.RecordCacheOperation("delete_multi", true, time.Since(start))
	}()

	if len(keys) == 0 {
		return nil
	}

	// Prepare full keys
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = r.getFullKey(key)
	}

	// Delete in batch
	if err := r.client.Del(ctx, fullKeys...).Err(); err != nil {
		return fmt.Errorf("failed to delete multi: %w", err)
	}

	return nil
}

// FlushAll clears all cache entries
func (r *RedisCache) FlushAll(ctx context.Context) error {
	start := time.Now()
	defer func() {
		r.metrics.RecordCacheOperation("flush_all", true, time.Since(start))
	}()

	// Use pattern matching to delete only our prefixed keys
	pattern := r.getFullKey("*")
	
	// Scan and delete in batches to avoid blocking Redis
	iter := r.client.Scan(ctx, 0, pattern, 1000).Iterator()
	for iter.Next(ctx) {
		if err := r.client.Del(ctx, iter.Val()).Err(); err != nil {
			return fmt.Errorf("failed to delete key %s: %w", iter.Val(), err)
		}
	}
	
	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan keys: %w", err)
	}

	return nil
}

// Stats returns cache statistics
func (r *RedisCache) Stats(ctx context.Context) (*storage.CacheStats, error) {
	start := time.Now()
	defer func() {
		r.metrics.RecordCacheOperation("stats", true, time.Since(start))
	}()

	info, err := r.client.Info(ctx, "stats", "memory", "keyspace").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis info: %w", err)
	}

	// Parse Redis info for statistics
	// This is a simplified implementation - you'd parse the actual info string
	stats := &storage.CacheStats{
		Hits:     0, // Would parse from keyspace_hits
		Misses:   0, // Would parse from keyspace_misses
		Keys:     0, // Would parse from db0:keys
		Memory:   0, // Would parse from used_memory
		HitRatio: 0.0,
	}

	// Calculate hit ratio
	if stats.Hits+stats.Misses > 0 {
		stats.HitRatio = float64(stats.Hits) / float64(stats.Hits+stats.Misses)
	}

	return stats, nil
}

// Close closes the Redis connection
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// getFullKey returns the full cache key with prefix
func (r *RedisCache) getFullKey(key string) string {
	if r.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", r.prefix, key)
}