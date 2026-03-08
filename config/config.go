// Package config handles environment and configuration loading.
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds the application configuration.
type Config struct {
	AnthropicAPIKey string
	DBPath          string
	LogLevel        string
	Transport       string
	HTTPAddr        string
	HTTPEndpoint    string
}

// Load reads configuration from environment variables and .env file.
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if not found)
	_ = godotenv.Load()

	transport, err := parseTransport(getEnvOrDefault("MCP_TRANSPORT", "stdio"))
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		AnthropicAPIKey: os.Getenv("ANTHROPIC_API_KEY"),
		DBPath:          getEnvOrDefault("DB_PATH", "./data/campaigns"),
		LogLevel:        getEnvOrDefault("LOG_LEVEL", "info"),
		Transport:       transport,
		HTTPAddr:        getEnvOrDefault("HTTP_ADDR", ":8080"),
		HTTPEndpoint:    getEnvOrDefault("MCP_HTTP_ENDPOINT", "/mcp"),
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseTransport(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "stdio":
		return "stdio", nil
	case "http":
		return "http", nil
	case "streamable-http", "streamable_http":
		return "streamable-http", nil
	default:
		return "", fmt.Errorf("invalid MCP_TRANSPORT %q: supported values are stdio, http, streamable-http", value)
	}
}
