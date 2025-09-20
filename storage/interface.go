// Package storage provides pluggable backend interfaces for enterprise OAuth2 SDK
package storage

import (
	"context"
	"time"

	"github.com/RichardKnop/go-oauth2-server/models"
)

// Storage defines the interface for OAuth2 data persistence
// Implementations must be thread-safe and support high concurrency
type Storage interface {
	// Client operations
	GetClient(ctx context.Context, clientID string) (*models.OauthClient, error)
	CreateClient(ctx context.Context, client *models.OauthClient) error
	UpdateClient(ctx context.Context, client *models.OauthClient) error
	DeleteClient(ctx context.Context, clientID string) error

	// User operations  
	GetUser(ctx context.Context, username string) (*models.OauthUser, error)
	GetUserByID(ctx context.Context, userID string) (*models.OauthUser, error)
	CreateUser(ctx context.Context, user *models.OauthUser) error
	AuthenticateUser(ctx context.Context, username, password string) (*models.OauthUser, error)

	// Token operations
	StoreAccessToken(ctx context.Context, token *models.OauthAccessToken) error
	GetAccessToken(ctx context.Context, tokenStr string) (*models.OauthAccessToken, error)
	DeleteAccessToken(ctx context.Context, tokenStr string) error
	CleanupExpiredTokens(ctx context.Context) error

	// Refresh token operations
	StoreRefreshToken(ctx context.Context, token *models.OauthRefreshToken) error
	GetRefreshToken(ctx context.Context, tokenStr string) (*models.OauthRefreshToken, error)
	DeleteRefreshToken(ctx context.Context, tokenStr string) error

	// Authorization code operations
	StoreAuthorizationCode(ctx context.Context, code *models.OauthAuthorizationCode) error
	GetAuthorizationCode(ctx context.Context, codeStr string) (*models.OauthAuthorizationCode, error)
	DeleteAuthorizationCode(ctx context.Context, codeStr string) error

	// Scope operations
	GetScope(ctx context.Context, scope string) (*models.OauthScope, error)
	GetDefaultScope(ctx context.Context) (string, error)
	
	// Batch operations for performance
	BatchGetTokens(ctx context.Context, tokens []string) ([]*models.OauthAccessToken, error)
	BatchDeleteTokens(ctx context.Context, tokens []string) error

	// Health and maintenance
	HealthCheck(ctx context.Context) error
	Close() error
}

// CacheProvider defines caching interface for performance optimization
type CacheProvider interface {
	// Basic cache operations
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
	
	// Batch operations
	SetMulti(ctx context.Context, items map[string]interface{}, ttl time.Duration) error
	GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error)
	DeleteMulti(ctx context.Context, keys []string) error
	
	// Cache management
	FlushAll(ctx context.Context) error
	Stats(ctx context.Context) (*CacheStats, error)
	Close() error
}

// CacheStats provides cache performance metrics
type CacheStats struct {
	Hits        int64   `json:"hits"`
	Misses      int64   `json:"misses"`
	Keys        int64   `json:"keys"`
	Memory      int64   `json:"memory_bytes"`
	HitRatio    float64 `json:"hit_ratio"`
}

// MetricsProvider defines interface for performance monitoring
type MetricsProvider interface {
	// Performance metrics
	RecordTokenGeneration(clientID, grantType string, duration time.Duration)
	RecordTokenValidation(valid bool, duration time.Duration)
	RecordDatabaseQuery(operation string, duration time.Duration, success bool)
	RecordCacheOperation(operation string, hit bool, duration time.Duration)
	
	// Business metrics
	IncrementActiveTokens(clientID string)
	DecrementActiveTokens(clientID string)
	RecordRateLimit(clientID string, limited bool)
	
	// System metrics
	RecordMemoryUsage(bytes int64)
	RecordGoroutineCount(count int)
	RecordRequestCount(endpoint, method, status string)
}

// StorageConfig provides configuration for storage backends
type StorageConfig struct {
	// Primary storage configuration
	Primary StorageBackend `json:"primary"`
	
	// Cache configuration
	Cache *CacheConfig `json:"cache,omitempty"`
	
	// Performance settings
	Performance *PerformanceConfig `json:"performance,omitempty"`
	
	// Monitoring settings
	Monitoring *MonitoringConfig `json:"monitoring,omitempty"`
}

// StorageBackend defines backend-specific configuration
type StorageBackend struct {
	Type   string                 `json:"type"`   // "postgres", "redis", "mongodb"
	Config map[string]interface{} `json:"config"` // Backend-specific settings
}

// CacheConfig defines caching configuration
type CacheConfig struct {
	Provider string                 `json:"provider"` // "redis", "memory"
	TTL      time.Duration          `json:"ttl"`
	Config   map[string]interface{} `json:"config"`
}

// PerformanceConfig defines performance optimization settings
type PerformanceConfig struct {
	// Connection pooling
	MaxOpenConnections int           `json:"max_open_connections"`
	MaxIdleConnections int           `json:"max_idle_connections"`
	ConnMaxLifetime    time.Duration `json:"connection_max_lifetime"`
	
	// Worker pools
	WorkerPoolSize int `json:"worker_pool_size"`
	QueueSize      int `json:"queue_size"`
	
	// Circuit breaker
	CircuitBreaker *CircuitBreakerConfig `json:"circuit_breaker,omitempty"`
}

// CircuitBreakerConfig defines circuit breaker settings
type CircuitBreakerConfig struct {
	Enabled          bool          `json:"enabled"`
	FailureThreshold int           `json:"failure_threshold"`
	RecoveryTimeout  time.Duration `json:"recovery_timeout"`
	MaxRequests      int           `json:"max_requests"`
}

// MonitoringConfig defines monitoring and observability settings
type MonitoringConfig struct {
	Enabled   bool   `json:"enabled"`
	Provider  string `json:"provider"` // "prometheus", "datadog"
	Namespace string `json:"namespace"`
	Subsystem string `json:"subsystem"`
}

// Factory creates storage instances based on configuration
type Factory interface {
	CreateStorage(config StorageConfig) (Storage, error)
	CreateCache(config CacheConfig) (CacheProvider, error)
	CreateMetrics(config MonitoringConfig) (MetricsProvider, error)
}