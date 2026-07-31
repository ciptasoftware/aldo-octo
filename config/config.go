package config

import (
	"bufio"
	"os"
	"strings"
)

// Environment specifies the runtime environment.
type Environment string

const (
	Development Environment = "development"
	Production  Environment = "production"
)

// Config holds the application configuration.
type Config struct {
	Env  Environment
	Port string
}

// Load reads configuration from environment variables with sensible defaults.
// It will attempt to load a .env file from the current directory if it exists.
func Load() *Config {
	loadEnvFile(".env")

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = string(Development)
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		Env:  Environment(env),
		Port: port,
	}
}

// IsProd returns true if running in production mode.
func (c *Config) IsProd() bool {
	return c.Env == Production
}

// loadEnvFile reads a simple .env file and sets environment variables if they are not already set.
// This is a minimal implementation that doesn't require third-party dependencies.
func loadEnvFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		// File does not exist or cannot be read; ignore
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])

			// Remove surrounding quotes if present
			if len(val) >= 2 && (val[0] == '"' || val[0] == '\'') && val[0] == val[len(val)-1] {
				val = val[1 : len(val)-1]
			}

			// Only set if not already set in the environment
			if _, exists := os.LookupEnv(key); !exists {
				os.Setenv(key, val)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		// Ignore read errors for optional .env files
		return
	}
}
