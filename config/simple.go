package config

import "time"

// SimpleConfig provides a basic configuration structure for the OAuth2 library
// This replaces the complex etcd/consul configuration system with a simple struct-based approach
type SimpleConfig struct {
	Database DatabaseConfig
	OAuth    OauthConfig
}

// NewSimpleConfig creates a new configuration with sensible defaults
func NewSimpleConfig() *SimpleConfig {
	return &SimpleConfig{
		Database: DatabaseConfig{
			Type:         "postgres",
			Host:         "localhost",
			Port:         5432,
			User:         "oauth2_user",
			Password:     "password",
			DatabaseName: "oauth2_db",
			MaxIdleConns: 5,
			MaxOpenConns: 5,
		},
		OAuth: OauthConfig{
			AccessTokenLifetime:  int(time.Hour.Seconds()),        // 1 hour
			RefreshTokenLifetime: int((24 * 14 * time.Hour).Seconds()), // 14 days
			AuthCodeLifetime:     int(time.Hour.Seconds()),        // 1 hour
		},
	}
}

// WithDatabase allows customizing database configuration
func (c *SimpleConfig) WithDatabase(dbConfig DatabaseConfig) *SimpleConfig {
	c.Database = dbConfig
	return c
}

// WithPostgreSQL sets up PostgreSQL database configuration
func (c *SimpleConfig) WithPostgreSQL(host string, port int, user, password, dbName string) *SimpleConfig {
	c.Database = DatabaseConfig{
		Type:         "postgres",
		Host:         host,
		Port:         port,
		User:         user,
		Password:     password,
		DatabaseName: dbName,
		MaxIdleConns: 5,
		MaxOpenConns: 5,
	}
	return c
}

// WithTokenLifetimes allows customizing token lifetimes (in seconds)
func (c *SimpleConfig) WithTokenLifetimes(accessToken, refreshToken, authCode int) *SimpleConfig {
	c.OAuth.AccessTokenLifetime = accessToken
	c.OAuth.RefreshTokenLifetime = refreshToken
	c.OAuth.AuthCodeLifetime = authCode
	return c
}

// ToConfig converts SimpleConfig to the original Config structure for internal use
func (c *SimpleConfig) ToConfig() *Config {
	return &Config{
		Database: c.Database,
		Oauth:    c.OAuth,
		Session: SessionConfig{
			Secret:   "library-default-secret", // Not used in library mode
			Path:     "/",
			MaxAge:   0,
			HTTPOnly: true,
		},
		IsDevelopment: false, // Library should be production-ready
	}
}