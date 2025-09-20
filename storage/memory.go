package storage

import (
	"context"
	"sync"
	"time"
	
	"github.com/RichardKnop/go-oauth2-server/models"
	"golang.org/x/crypto/bcrypt"
)

// MemoryStorage provides a simple in-memory storage implementation for development/testing
type MemoryStorage struct {
	mu           sync.RWMutex
	clients      map[string]*models.OauthClient
	users        map[string]*models.OauthUser
	usersByID    map[string]*models.OauthUser
	accessTokens map[string]*models.OauthAccessToken
	refreshTokens map[string]*models.OauthRefreshToken
	authCodes    map[string]*models.OauthAuthorizationCode
	scopes       map[string]*models.OauthScope
}

// NewMemoryStorage creates a new in-memory storage instance
func NewMemoryStorage() Storage {
	return &MemoryStorage{
		clients:       make(map[string]*models.OauthClient),
		users:         make(map[string]*models.OauthUser),
		usersByID:     make(map[string]*models.OauthUser),
		accessTokens:  make(map[string]*models.OauthAccessToken),
		refreshTokens: make(map[string]*models.OauthRefreshToken),
		authCodes:     make(map[string]*models.OauthAuthorizationCode),
		scopes:        make(map[string]*models.OauthScope),
	}
}

// Client operations
func (m *MemoryStorage) GetClient(ctx context.Context, clientID string) (*models.OauthClient, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if client, exists := m.clients[clientID]; exists {
		return client, nil
	}
	return nil, ErrClientNotFound
}

func (m *MemoryStorage) CreateClient(ctx context.Context, client *models.OauthClient) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	client.CreatedAt = time.Now().UTC()
	m.clients[client.Key] = client
	return nil
}

func (m *MemoryStorage) UpdateClient(ctx context.Context, client *models.OauthClient) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.clients[client.Key]; !exists {
		return ErrClientNotFound
	}
	m.clients[client.Key] = client
	return nil
}

func (m *MemoryStorage) DeleteClient(ctx context.Context, clientID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clients, clientID)
	return nil
}

// User operations
func (m *MemoryStorage) GetUser(ctx context.Context, username string) (*models.OauthUser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if user, exists := m.users[username]; exists {
		return user, nil
	}
	return nil, ErrUserNotFound
}

func (m *MemoryStorage) GetUserByID(ctx context.Context, userID string) (*models.OauthUser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if user, exists := m.usersByID[userID]; exists {
		return user, nil
	}
	return nil, ErrUserNotFound
}

func (m *MemoryStorage) CreateUser(ctx context.Context, user *models.OauthUser) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	user.CreatedAt = time.Now().UTC()
	m.users[user.Username] = user
	m.usersByID[user.ID] = user
	return nil
}

func (m *MemoryStorage) AuthenticateUser(ctx context.Context, username, password string) (*models.OauthUser, error) {
	user, err := m.GetUser(ctx, username)
	if err != nil {
		return nil, err
	}
	
	// Check password - handle sql.NullString
	var hashedPassword string
	if user.Password.Valid {
		hashedPassword = user.Password.String
	} else {
		return nil, ErrInvalidCredentials
	}
	
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	
	return user, nil
}

// Token operations
func (m *MemoryStorage) StoreAccessToken(ctx context.Context, token *models.OauthAccessToken) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	token.CreatedAt = time.Now().UTC()
	m.accessTokens[token.Token] = token
	return nil
}

func (m *MemoryStorage) GetAccessToken(ctx context.Context, tokenStr string) (*models.OauthAccessToken, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if token, exists := m.accessTokens[tokenStr]; exists {
		if token.ExpiresAt.Before(time.Now().UTC()) {
			return nil, ErrTokenExpired
		}
		return token, nil
	}
	return nil, ErrTokenNotFound
}

func (m *MemoryStorage) DeleteAccessToken(ctx context.Context, tokenStr string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.accessTokens, tokenStr)
	return nil
}

func (m *MemoryStorage) CleanupExpiredTokens(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	now := time.Now().UTC()
	for token, accessToken := range m.accessTokens {
		if accessToken.ExpiresAt.Before(now) {
			delete(m.accessTokens, token)
		}
	}
	
	for token, refreshToken := range m.refreshTokens {
		if refreshToken.ExpiresAt.Before(now) {
			delete(m.refreshTokens, token)
		}
	}
	
	return nil
}

// Refresh token operations
func (m *MemoryStorage) StoreRefreshToken(ctx context.Context, token *models.OauthRefreshToken) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	token.CreatedAt = time.Now().UTC()
	m.refreshTokens[token.Token] = token
	return nil
}

func (m *MemoryStorage) GetRefreshToken(ctx context.Context, tokenStr string) (*models.OauthRefreshToken, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if token, exists := m.refreshTokens[tokenStr]; exists {
		if token.ExpiresAt.Before(time.Now().UTC()) {
			return nil, ErrTokenExpired
		}
		return token, nil
	}
	return nil, ErrTokenNotFound
}

func (m *MemoryStorage) DeleteRefreshToken(ctx context.Context, tokenStr string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.refreshTokens, tokenStr)
	return nil
}

// Authorization code operations
func (m *MemoryStorage) StoreAuthorizationCode(ctx context.Context, code *models.OauthAuthorizationCode) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	code.CreatedAt = time.Now().UTC()
	m.authCodes[code.Code] = code
	return nil
}

func (m *MemoryStorage) GetAuthorizationCode(ctx context.Context, codeStr string) (*models.OauthAuthorizationCode, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if code, exists := m.authCodes[codeStr]; exists {
		if code.ExpiresAt.Before(time.Now().UTC()) {
			return nil, ErrCodeExpired
		}
		return code, nil
	}
	return nil, ErrCodeNotFound
}

func (m *MemoryStorage) DeleteAuthorizationCode(ctx context.Context, codeStr string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.authCodes, codeStr)
	return nil
}

// Scope operations
func (m *MemoryStorage) GetScope(ctx context.Context, scope string) (*models.OauthScope, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if scopeObj, exists := m.scopes[scope]; exists {
		return scopeObj, nil
	}
	return nil, ErrScopeNotFound
}

func (m *MemoryStorage) GetDefaultScope(ctx context.Context) (string, error) {
	return "read", nil // Default scope for development
}

// Batch operations (simplified stubs)
func (m *MemoryStorage) BatchGetTokens(ctx context.Context, tokens []string) ([]*models.OauthAccessToken, error) {
	var result []*models.OauthAccessToken
	for _, token := range tokens {
		if t, err := m.GetAccessToken(ctx, token); err == nil {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *MemoryStorage) BatchDeleteTokens(ctx context.Context, tokens []string) error {
	for _, token := range tokens {
		m.DeleteAccessToken(ctx, token)
	}
	return nil
}

// Health check
func (m *MemoryStorage) HealthCheck(ctx context.Context) error {
	return nil // Always healthy for memory storage
}

// Close
func (m *MemoryStorage) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Clear all maps
	m.clients = nil
	m.users = nil
	m.usersByID = nil
	m.accessTokens = nil
	m.refreshTokens = nil
	m.authCodes = nil
	m.scopes = nil
	
	return nil
}