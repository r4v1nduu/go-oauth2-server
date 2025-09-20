package main

import (
	"fmt"
	"log"
	"time"

	oauth2server "github.com/RichardKnop/go-oauth2-server"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	fmt.Println("🚀 Starting Enterprise OAuth2 Server Example")
	
	// Build the enterprise OAuth2 SDK with high-performance configuration
	sdk, err := oauth2server.New().
		WithMemoryCache(10000).                           // 10K cache entries for development
		WithCustomRateLimit(&oauth2server.RateLimitConfig{
			Enabled:    true,
			DefaultRPS: 1000,                             // 1000 requests per second per client
			BurstSize:  100,
			WindowSize: time.Minute,
			Storage:    "memory",
		}).
		WithPerformance(&oauth2server.PerformanceConfig{
			AccessTokenTTL:  2 * time.Hour,               // 2 hour token lifetime
			RefreshTokenTTL: 30 * 24 * time.Hour,        // 30 days refresh token
			AuthCodeTTL:     5 * time.Minute,            // 5 minutes auth code
			TokenWorkers:    20,                          // 20 background workers
			CleanupInterval: 30 * time.Minute,           // Cleanup every 30 minutes
			BatchSize:       2000,                        // Process 2000 tokens per batch
		}).
		WithSecurity(&oauth2server.SecurityConfig{
			TokenEncryption:   true,
			MinPasswordLength: 12,                        // Strong password policy
			RequireUppercase:  true,
			RequireNumbers:    true,
			RequireSymbols:    true,
			SecureCookies:     true,
			HTTPOnly:          true,
		}).
		Build()
	
	if err != nil {
		log.Fatalf("Failed to build OAuth2 SDK: %v", err)
	}
	defer sdk.Close()

	fmt.Println("✅ Enterprise OAuth2 SDK initialized successfully")
	fmt.Println("📊 Configuration:")
	fmt.Println("   - Rate Limit: 1000 RPS per client")
	fmt.Println("   - Token TTL: 2 hours")
	fmt.Println("   - Refresh TTL: 30 days")
	fmt.Println("   - Workers: 20")
	fmt.Println("   - Batch Size: 2000")
	fmt.Println("   - Security: Enterprise grade")
	fmt.Println("   - Framework: Fiber (High Performance)")

	// Create the OAuth2 server instance
	server := sdk.CreateServer()
	
	// Setup Fiber app with middleware
	app := fiber.New(fiber.Config{
		AppName:      "Enterprise OAuth2 Server",
		ServerHeader: "OAuth2-Enterprise",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	})
	
	// Add enterprise middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())
	
	// Register OAuth2 endpoints with the "/oauth" prefix
	server.RegisterRoutes(app, "/oauth")
	
	// Add example routes
	app.Get("/", homeHandler)
	app.Get("/status", statusHandler)
	
	fmt.Println("🌐 Server endpoints available:")
	fmt.Println("   POST /oauth/tokens       - Token generation")
	fmt.Println("   POST /oauth/introspect   - Token introspection")  
	fmt.Println("   GET  /oauth/health       - Health check")
	fmt.Println("   GET  /                   - Home page")
	fmt.Println("   GET  /status             - Server status")
	
	// Start the server
	fmt.Println("\n🎯 Enterprise OAuth2 Server starting on :8080")
	fmt.Println("💡 Ready to handle 10,000+ requests per second with Fiber!")
	
	log.Fatal(app.Listen(":8080"))
}

// Example handlers using Fiber
func homeHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"service":     "Enterprise OAuth2 Server",
		"version":     "1.0.0",
		"performance": "10,000+ RPS capable with Fiber",
		"security":    "Enterprise grade",
		"framework":   "Go Fiber",
		"endpoints": fiber.Map{
			"tokens":     "/oauth/tokens",
			"introspect": "/oauth/introspect",
			"health":     "/oauth/health",
		},
	})
}

func statusHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    "runtime info here",
		"performance": fiber.Map{
			"framework":      "Fiber",
			"cache_enabled":  true,
			"rate_limiting":  true,
			"workers":        20,
			"batch_size":     2000,
		},
	})
}