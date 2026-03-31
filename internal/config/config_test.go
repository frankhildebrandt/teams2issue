package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/frankhildebrandt/teams2issue/internal/config"
)

func TestLoadConfigFileAndEnvOverride(t *testing.T) {
	t.Setenv("TEAMS2ISSUE_LOGGING_LEVEL", "debug")

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	configYAML := []byte(`app:
  name: bootstrap-test
http:
  address: 127.0.0.1:9090
metrics:
  path: /metrics
`)

	if err := os.WriteFile(configPath, configYAML, 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.App.Name != "bootstrap-test" {
		t.Fatalf("expected app name from file, got %q", cfg.App.Name)
	}
	if cfg.HTTP.Address != "127.0.0.1:9090" {
		t.Fatalf("expected http address override, got %q", cfg.HTTP.Address)
	}
	if cfg.Logging.Level != "debug" {
		t.Fatalf("expected env override for logging level, got %q", cfg.Logging.Level)
	}
	if cfg.App.StartupTimeout != 10*time.Second {
		t.Fatalf("expected default startup timeout, got %s", cfg.App.StartupTimeout)
	}
}
