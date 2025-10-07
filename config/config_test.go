package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Test with default values
	cfg := Load()

	if cfg.Database.Host != "127.0.0.1" {
		t.Errorf("Expected default DB host 127.0.0.1, got %s", cfg.Database.Host)
	}

	if cfg.Database.Port != "5434" {
		t.Errorf("Expected default DB port 5434, got %s", cfg.Database.Port)
	}

	if cfg.Server.Port != "8080" {
		t.Errorf("Expected default server port 8080, got %s", cfg.Server.Port)
	}

	if cfg.Cache.MaxSize != 1000 {
		t.Errorf("Expected default cache max size 1000, got %d", cfg.Cache.MaxSize)
	}

	if cfg.Cache.TTL != 3600 {
		t.Errorf("Expected default cache TTL 3600, got %d", cfg.Cache.TTL)
	}
}

func TestLoadWithEnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("DB_HOST", "testhost")
	os.Setenv("DB_PORT", "5555")
	os.Setenv("SERVER_PORT", "9000")
	os.Setenv("CACHE_MAX_SIZE", "500")
	os.Setenv("CACHE_TTL", "1800")

	defer func() {
		// Clean up environment variables
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("CACHE_MAX_SIZE")
		os.Unsetenv("CACHE_TTL")
	}()

	cfg := Load()

	if cfg.Database.Host != "testhost" {
		t.Errorf("Expected DB host testhost, got %s", cfg.Database.Host)
	}

	if cfg.Database.Port != "5555" {
		t.Errorf("Expected DB port 5555, got %s", cfg.Database.Port)
	}

	if cfg.Server.Port != "9000" {
		t.Errorf("Expected server port 9000, got %s", cfg.Server.Port)
	}

	if cfg.Cache.MaxSize != 500 {
		t.Errorf("Expected cache max size 500, got %d", cfg.Cache.MaxSize)
	}

	if cfg.Cache.TTL != 1800 {
		t.Errorf("Expected cache TTL 1800, got %d", cfg.Cache.TTL)
	}
}

func TestDatabaseDSN(t *testing.T) {
	cfg := &Config{
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "testuser",
			Password: "testpass",
			DBName:   "testdb",
			SSLMode:  "disable",
		},
	}

	expected := "host=localhost user=testuser password=testpass dbname=testdb port=5432 sslmode=disable"
	if dsn := cfg.DatabaseDSN(); dsn != expected {
		t.Errorf("Expected DSN %s, got %s", expected, dsn)
	}
}

func TestGetEnvAsInt(t *testing.T) {
	tests := []struct {
		envValue    string
		defaultVal  int
		expected    int
		description string
	}{
		{"", 100, 100, "empty environment variable should return default"},
		{"200", 100, 200, "valid integer should be parsed"},
		{"invalid", 100, 100, "invalid integer should return default"},
		{"0", 100, 0, "zero should be parsed correctly"},
		{"-50", 100, -50, "negative integer should be parsed"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			os.Setenv("TEST_VAR", tt.envValue)
			defer os.Unsetenv("TEST_VAR")

			result := getEnvAsInt("TEST_VAR", tt.defaultVal)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d for env value '%s'", tt.expected, result, tt.envValue)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		envValue    string
		defaultVal  string
		expected    string
		description string
	}{
		{"", "default", "default", "empty environment variable should return default"},
		{"custom", "default", "custom", "set environment variable should return custom value"},
		{" spaces ", "default", " spaces ", "spaces should be preserved"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			if tt.envValue == "" {
				os.Unsetenv("TEST_VAR")
			} else {
				os.Setenv("TEST_VAR", tt.envValue)
			}
			defer os.Unsetenv("TEST_VAR")

			result := getEnv("TEST_VAR", tt.defaultVal)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s' for env value '%s'", tt.expected, result, tt.envValue)
			}
		})
	}
}