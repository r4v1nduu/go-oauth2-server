package postgres
// Package postgres provides high-performance PostgreSQL storage implementation
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/RichardKnop/go-oauth2-server/models"
	"github.com/RichardKnop/go-oauth2-server/storage"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

// PostgreSQLStorage implements high-performance PostgreSQL backend
type PostgreSQLStorage struct {
	db      *gorm.DB
	metrics storage.MetricsProvider
	cache   storage.CacheProvider
	config  *PostgreSQLConfig
}

// PostgreSQLConfig defines PostgreSQL-specific configuration
type PostgreSQLConfig struct {
	// Connection settings
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	SSLMode  string `json:"ssl_mode"`

	// Performance settings
	MaxOpenConnections int           `json:"max_open_connections"`
	MaxIdleConnections int           `json:"max_idle_connections"`
	ConnMaxLifetime    time.Duration `json:"connection_max_lifetime"`

	// Query optimization
	PrepareStatements bool `json:"prepare_statements"`
	QueryTimeout      time.Duration `json:"query_timeout"`
}

// NewPostgreSQLStorage creates a new high-performance PostgreSQL storage instance
func NewPostgreSQLStorage(config *PostgreSQLConfig, cache storage.CacheProvider, metrics storage.MetricsProvider) (*PostgreSQLStorage, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Database, config.Password, config.SSLMode)

	db, err := gorm.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Configure connection pool for high performance
	db.DB().SetMaxOpenConns(config.MaxOpenConnections)
	db.DB().SetMaxIdleConns(config.MaxIdleConnections)
	db.DB().SetConnMaxLifetime(config.ConnMaxLifetime)

	// Enable query logging in development
	db.LogMode(false) // Disable for production performance

	storage := &PostgreSQLStorage{
		db:      db,
		metrics: metrics,
		cache:   cache,
		config:  config,
	}

	return storage, nil
}

// GetClient retrieves a client with caching support
func (s *PostgreSQLStorage) GetClient(ctx context.Context, clientID string) (*models.OauthClient, error) {
	start := time.Now()
	defer func() {
		s.metrics.RecordDatabaseQuery("get_client", time.Since(start), true)
	}()

	// Try cache first
	if s.cache != nil {
		cacheKey := fmt.Sprintf("client:%s", clientID)
		var client models.OauthClient
		if err := s.cache.Get(ctx, cacheKey, &client); err == nil {
			s.metrics.RecordCacheOperation("get_client", true, time.Since(start))
			return &client, nil
		}
		s.metrics.RecordCacheOperation("get_client", false, time.Since(start))
	}

	// Query database
	var client models.OauthClient
	if err := s.db.Where("key = ?", clientID).First(&client).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	// Cache the result
	if s.cache != nil {
		cacheKey := fmt.Sprintf("client:%s", clientID)
		s.cache.Set(ctx, cacheKey, &client, 5*time.Minute)
	}

	return &client, nil
}

// CreateClient creates a new OAuth client
func (s *PostgreSQLStorage) CreateClient(ctx context.Context, client *models.OauthClient) error {
	start := time.Now()
	defer func() {
		s.metrics.RecordDatabaseQuery("create_client", time.Since(start), true)
	}()

	if err := s.db.Create(client).Error; err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Invalidate cache
	if s.cache != nil {
		cacheKey := fmt.Sprintf("client:%s", client.Key)
		s.cache.Delete(ctx, cacheKey)
	}

	return nil
}

// UpdateClient updates an existing OAuth client
func (s *PostgreSQLStorage) UpdateClient(ctx context.Context, client *models.OauthClient) error {
	start := time.Now()
	defer func() {
		s.metrics.RecordDatabaseQuery("update_client", time.Since(start), true)
	}()

	if err := s.db.Save(client).Error; err != nil {
		return fmt.Errorf("failed to update client: %w", err)
	}

	// Invalidate cache
	if s.cache != nil {
		cacheKey := fmt.Sprintf("client:%s", client.Key)
		s.cache.Delete(ctx, cacheKey)
	}

	return nil
}

// DeleteClient deletes an OAuth client
func (s *PostgreSQLStorage) DeleteClient(ctx context.Context, clientID string) error {
	start := time.Now()
	defer func() {
		s.metrics.RecordDatabaseQuery("delete_client", time.Since(start), true)
	}()

	if err := s.db.Where("key = ?", clientID).Delete(&models.OauthClient{}).Error; err != nil {
		return fmt.Errorf("failed to delete client: %w", err)
	}

	// Invalidate cache
	if s.cache != nil {
		cacheKey := fmt.Sprintf("client:%s", clientID)
		s.cache.Delete(ctx, cacheKey)
	}

	return nil
}

// GetUser retrieves a user with caching support
func (s *PostgreSQLStorage) GetUser(ctx context.Context, username string) (*models.OauthUser, error) {
	start := time.Now()
	defer func() {
		s.metrics.RecordDatabaseQuery("get_user", time.Since(start), true)
	}()

	// Try cache first
	if s.cache != nil {
		cacheKey := fmt.Sprintf("user:%s", username)
		var user models.OauthUser
		if err := s.cache.Get(ctx, cacheKey, &user); err == nil {
			s.metrics.RecordCacheOperation("get_user", true, time.Since(start))
			return &user, nil
		}
		s.metrics.RecordCacheOperation("get_user", false, time.Since(start))
	}

	// Query database
	var user models.OauthUser
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Cache the result
	if s.cache != nil {
		cacheKey := fmt.Sprintf("user:%s", username)
		s.cache.Set(ctx, cacheKey, &user, 5*time.Minute)
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (s *PostgreSQLStorage) GetUserByID(ctx context.Context, userID string) (*models.OauthUser, error) {
	start := time.Now()
	defer func() {
		s.metrics.RecordDatabaseQuery("get_user_by_id", time.Since(start), true)
	}()

	var user models.OauthUser
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

// CreateUser creates a new user
func (s *PostgreSQLStorage) CreateUser(ctx context.Context, user *models.OauthUser) error {
	start := time.Now()
	defer func() {
		s.metrics.RecordDatabaseQuery("create_user", time.Since(start), true)
	}()

	if err := s.db.Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// AuthenticateUser authenticates a user with username and password
func (s *PostgreSQLStorage) AuthenticateUser(ctx context.Context, username, password string) (*models.OauthUser, error) {
	start := time.Now()
	defer func() {
		s.metrics.RecordDatabaseQuery("authenticate_user", time.Since(start), true)
	}()

	user, err := s.GetUser(ctx, username)
	if err != nil {
		return nil, err
	}
	
	if user == nil {
		return nil, nil // User not found
	}

	// TODO: Implement password verification
	// This would typically involve bcrypt.CompareHashAndPassword
	
	return user, nil
}

// StoreAccessToken stores an access token with optimized indexing
func (s *PostgreSQLStorage) StoreAccessToken(ctx context.Context, token *models.OauthAccessToken) error {
	start := time.Now()
	defer func() {
		s.metrics.RecordDatabaseQuery("store_access_token", time.Since(start), true)
		s.metrics.IncrementActiveTokens(token.Client.Key)
	}()

	if err := s.db.Create(token).Error; err != nil {
		return fmt.Errorf("failed to store access token: %w", err)
	}

	// Cache the token for fast lookup
	if s.cache != nil {
		cacheKey := fmt.Sprintf("access_token:%s", token.Token)
		s.cache.Set(ctx, cacheKey, token, time.Until(token.ExpiresAt))
	}

	return nil
}

// GetAccessToken retrieves an access token with caching
func (s *PostgreSQLStorage) GetAccessToken(ctx context.Context, tokenStr string) (*models.OauthAccessToken, error) {
	start := time.Now()
	defer func() {
		s.metrics.RecordDatabaseQuery("get_access_token", time.Since(start), true)
	}()

	// Try cache first
	if s.cache != nil {
		cacheKey := fmt.Sprintf("access_token:%s", tokenStr)
		var token models.OauthAccessToken
		if err := s.cache.Get(ctx, cacheKey, &token); err == nil {
			s.metrics.RecordCacheOperation("get_access_token", true, time.Since(start))
			return &token, nil
		}
		s.metrics.RecordCacheOperation("get_access_token", false, time.Since(start))
	}

	// Query database with preloading for performance
	var token models.OauthAccessToken
	if err := s.db.Preload("Client").Preload("User").Where("token = ?", tokenStr).First(&token).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Cache the result if not expired
	if s.cache != nil && token.ExpiresAt.After(time.Now()) {
		cacheKey := fmt.Sprintf("access_token:%s", tokenStr)
		s.cache.Set(ctx, cacheKey, &token, time.Until(token.ExpiresAt))
	}

	return &token, nil
}

// DeleteAccessToken deletes an access token
func (s *PostgreSQLStorage) DeleteAccessToken(ctx context.Context, tokenStr string) error {
	start := time.Now()
	defer func() {
		s.metrics.RecordDatabaseQuery("delete_access_token", time.Since(start), true)
	}()

	if err := s.db.Where("token = ?", tokenStr).Delete(&models.OauthAccessToken{}).Error; err != nil {
		return fmt.Errorf("failed to delete access token: %w", err)
	}

	// Remove from cache
	if s.cache != nil {
		cacheKey := fmt.Sprintf("access_token:%s", tokenStr)
		s.cache.Delete(ctx, cacheKey)
	}

	return nil
}

// BatchGetTokens retrieves multiple tokens in a single query for performance
func (s *PostgreSQLStorage) BatchGetTokens(ctx context.Context, tokens []string) ([]*models.OauthAccessToken, error) {
	start := time.Now()
	defer func() {
		s.metrics.RecordDatabaseQuery("batch_get_tokens", time.Since(start), true)
	}()

	var accessTokens []*models.OauthAccessToken
	if err := s.db.Preload("Client").Preload("User").Where("token IN (?)", tokens).Find(&accessTokens).Error; err != nil {
		return nil, fmt.Errorf("failed to batch get tokens: %w", err)
	}

	return accessTokens, nil
}

// BatchDeleteTokens deletes multiple tokens in a single query
func (s *PostgreSQLStorage) BatchDeleteTokens(ctx context.Context, tokens []string) error {
	start := time.Now()
	defer func() {
		s.metrics.RecordDatabaseQuery("batch_delete_tokens", time.Since(start), true)
	}()

	if err := s.db.Where("token IN (?)", tokens).Delete(&models.OauthAccessToken{}).Error; err != nil {
		return fmt.Errorf("failed to batch delete tokens: %w", err)
	}

	// Remove from cache
	if s.cache != nil {
		var cacheKeys []string
		for _, token := range tokens {
			cacheKeys = append(cacheKeys, fmt.Sprintf("access_token:%s", token))
		}
		s.cache.DeleteMulti(ctx, cacheKeys)
	}

	return nil
}

// CleanupExpiredTokens removes expired tokens for database maintenance
func (s *PostgreSQLStorage) CleanupExpiredTokens(ctx context.Context) error {
	start := time.Now()
	defer func() {
		s.metrics.RecordDatabaseQuery("cleanup_expired_tokens", time.Since(start), true)
	}()

	now := time.Now()
	
	// Clean up access tokens
	if err := s.db.Where("expires_at < ?", now).Delete(&models.OauthAccessToken{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup expired access tokens: %w", err)
	}
	
	// Clean up refresh tokens
	if err := s.db.Where("expires_at < ?", now).Delete(&models.OauthRefreshToken{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup expired refresh tokens: %w", err)
	}
	
	// Clean up authorization codes
	if err := s.db.Where("expires_at < ?", now).Delete(&models.OauthAuthorizationCode{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup expired authorization codes: %w", err)
	}

	return nil
}

// HealthCheck verifies database connectivity
func (s *PostgreSQLStorage) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return s.db.DB().PingContext(ctx)
}

// Close closes the database connection
func (s *PostgreSQLStorage) Close() error {
	return s.db.Close()
}

// Additional methods would implement the remaining Storage interface methods...
// StoreRefreshToken, GetRefreshToken, DeleteRefreshToken
// StoreAuthorizationCode, GetAuthorizationCode, DeleteAuthorizationCode  
// GetScope, GetDefaultScope