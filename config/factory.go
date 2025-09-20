package config

import (
	"time"
)

var (
	configLoaded   bool
	dialTimeout    = 5 * time.Second
	contextTimeout = 5 * time.Second
	reloadDelay    = time.Second * 10
)

// Cnf ...
// Let's start with some sensible defaults
var Cnf = &Config{
	Database: DatabaseConfig{
		Type:         "postgres",
		Host:         "localhost",
		Port:         5432,
		User:         "go_oauth2_server",
		Password:     "password",
		DatabaseName: "go_oauth2_server",
		MaxIdleConns: 5,
		MaxOpenConns: 5,
	},
	Oauth: OauthConfig{
		AccessTokenLifetime:  3600,    // 1 hour
		RefreshTokenLifetime: 1209600, // 14 days
		AuthCodeLifetime:     3600,    // 1 hour
	},
	Session: SessionConfig{
		Secret:   "test_secret",
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HTTPOnly: true,
	},
	IsDevelopment: true,
}

// NewConfig creates a configuration - simplified version for library use
func NewConfig(mustLoadOnce bool, keepReloading bool, backendType string) *Config {
	// For library use, we only support simple configuration
	configLoaded = true
	return Cnf
}

// SetConfig allows setting configuration directly for library use
func SetConfig(config *Config) {
	Cnf = config
	configLoaded = true
}
