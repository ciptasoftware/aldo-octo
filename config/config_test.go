package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	// Clear env to test defaults
	os.Unsetenv("APP_ENV")
	os.Unsetenv("APP_PORT")

	cfg := Load()

	if cfg.Env != Development {
		t.Errorf("expected Env=%q, got %q", Development, cfg.Env)
	}
	if cfg.Port != "8080" {
		t.Errorf("expected Port=%q, got %q", "8080", cfg.Port)
	}
}

func TestLoadFromEnv(t *testing.T) {
	os.Setenv("APP_ENV", "production")
	os.Setenv("APP_PORT", "3000")
	defer os.Unsetenv("APP_ENV")
	defer os.Unsetenv("APP_PORT")

	cfg := Load()

	if cfg.Env != Production {
		t.Errorf("expected Env=%q, got %q", Production, cfg.Env)
	}
	if cfg.Port != "3000" {
		t.Errorf("expected Port=%q, got %q", "3000", cfg.Port)
	}
}

func TestIsProd(t *testing.T) {
	tests := []struct {
		env      Environment
		expected bool
	}{
		{Production, true},
		{Development, false},
		{"staging", false},
	}

	for _, tt := range tests {
		cfg := &Config{Env: tt.env}
		if got := cfg.IsProd(); got != tt.expected {
			t.Errorf("IsProd() for env=%q: expected %v, got %v", tt.env, tt.expected, got)
		}
	}
}

func TestLoadEnvFile(t *testing.T) {
	// Create a temporary .env file
	content := []byte("TEST_LOAD_ENV_KEY=\"hello world\"\n# comment line\nTEST_LOAD_ENV_KEY2=no_quotes\n")
	tmpFile := ".env.test"
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile)

	// Ensure keys don't exist
	os.Unsetenv("TEST_LOAD_ENV_KEY")
	os.Unsetenv("TEST_LOAD_ENV_KEY2")

	loadEnvFile(tmpFile)

	val := os.Getenv("TEST_LOAD_ENV_KEY")
	if val != "hello world" {
		t.Errorf("expected %q, got %q", "hello world", val)
	}

	val2 := os.Getenv("TEST_LOAD_ENV_KEY2")
	if val2 != "no_quotes" {
		t.Errorf("expected %q, got %q", "no_quotes", val2)
	}

	// Cleanup
	os.Unsetenv("TEST_LOAD_ENV_KEY")
	os.Unsetenv("TEST_LOAD_ENV_KEY2")
}

func TestLoadEnvFileDoesNotOverride(t *testing.T) {
	// Pre-set a value
	os.Setenv("TEST_OVERRIDE_KEY", "original")
	defer os.Unsetenv("TEST_OVERRIDE_KEY")

	content := []byte("TEST_OVERRIDE_KEY=overridden\n")
	tmpFile := ".env.override.test"
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile)

	loadEnvFile(tmpFile)

	val := os.Getenv("TEST_OVERRIDE_KEY")
	if val != "original" {
		t.Errorf("expected original value %q, got %q", "original", val)
	}
}

func TestLoadEnvFileMissing(t *testing.T) {
	// Should not panic or error on missing file
	loadEnvFile("nonexistent.env.file")
}
