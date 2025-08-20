package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flarexes/gitback/internal/types"
)

// Helper function to check if two configs are equal
func equalConfig(a, b *types.Config) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.NoAuth == b.NoAuth &&
		a.Threads == b.Threads &&
		a.Token == b.Token &&
		a.User == b.User &&
		a.OutputDir == b.OutputDir &&
		a.Timeout == b.Timeout &&
		a.IncludeGists == b.IncludeGists
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	
	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}
	
	if cfg.NoAuth != false {
		t.Errorf("Expected NoAuth to be false, got %v", cfg.NoAuth)
	}
	
	if cfg.Threads != 5 {
		t.Errorf("Expected Threads to be 5, got %d", cfg.Threads)
	}
	
	if cfg.Token != "" {
		t.Errorf("Expected Token to be empty, got %q", cfg.Token)
	}
	
	if cfg.User != "" {
		t.Errorf("Expected User to be empty, got %q", cfg.User)
	}
	
	expectedOutputDir := filepath.Join(os.Getenv("HOME"), "gitbackup")
	if cfg.OutputDir != expectedOutputDir {
		t.Errorf("Expected OutputDir to be %q, got %q", expectedOutputDir, cfg.OutputDir)
	}
	
	if cfg.Timeout != 30 {
		t.Errorf("Expected Timeout to be 30, got %d", cfg.Timeout)
	}
	
	if !cfg.IncludeGists {
		t.Error("Expected IncludeGists to be true, got false")
	}
}

func TestLoadFromEnv(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		want    *types.Config
		wantErr bool
	}{
		{
			name: "all fields",
			env: map[string]string{
				"GITBACK_NOAUTH":        "true",
				"GITBACK_THREADS":       "10",
				"GITHUB_TOKEN":          "test-token",
				"GITBACK_USER":          "testuser",
				"GITBACK_OUTPUT_DIR":    "/custom/path",
				"GITBACK_TIMEOUT":       "60",
				"GITBACK_INCLUDE_GISTS": "false",
			},
			want: &types.Config{
				NoAuth:      true,
				Threads:     10,
				Token:       "test-token",
				User:        "testuser",
				OutputDir:   "/custom/path",
				Timeout:     60,
				IncludeGists: false,
			},
		},
		{
			name: "invalid threads",
			env: map[string]string{
				"GITBACK_THREADS": "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			cfg := DefaultConfig()
			err := LoadFromEnv(cfg)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !equalConfig(cfg, tt.want) {
				t.Errorf("Config mismatch\nGot: %+v\nWant: %+v", cfg, tt.want)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	tests := []struct {
		name    string
		cfg     *types.Config
		wantErr bool
	}{
		{
			name: "valid config with auth",
			cfg: &types.Config{
				NoAuth:    false,
				Token:     "test-token",
				User:      "testuser",
				Threads:   5,
				Timeout:   30,
				OutputDir: tempDir,
			},
			wantErr: false,
		},
		{
			name: "valid config without auth",
			cfg: &types.Config{
				NoAuth:    true,
				Token:     "",
				User:      "testuser",
				Threads:   5,
				Timeout:   30,
				OutputDir: tempDir,
			},
			wantErr: false,
		},
		{
			name: "missing user in noauth mode",
			cfg: &types.Config{
				NoAuth:    true,
				Token:     "",
				User:      "",
				Threads:   5,
				Timeout:   30,
				OutputDir: tempDir,
			},
			wantErr: true,
		},
		{
			name: "invalid thread count",
			cfg: &types.Config{
				NoAuth:    false,
				Token:     "test-token",
				User:      "testuser",
				Threads:   0,
				Timeout:   30,
				OutputDir: tempDir,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.cfg)
			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			} else if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestSanitize(t *testing.T) {
	cfg := &types.Config{
		NoAuth:      true,
		Threads:     5,
		Token:       "sensitive-token",
		User:        "testuser",
		OutputDir:   "/test/dir",
		Timeout:     30,
		IncludeGists: true,
	}

	sanitized := Sanitize(cfg)

	// Original config should not be modified
	if cfg.Token != "sensitive-token" {
		t.Errorf("Expected original token to remain unchanged, got %q", cfg.Token)
	}
	
	// Sanitized config should have empty token
	if sanitized.Token != "" {
		t.Errorf("Expected sanitized token to be empty, got %q", sanitized.Token)
	}
	
	// Other fields should be copied
	if cfg.NoAuth != sanitized.NoAuth {
		t.Errorf("Expected NoAuth %v, got %v", cfg.NoAuth, sanitized.NoAuth)
	}
	if cfg.Threads != sanitized.Threads {
		t.Errorf("Expected Threads %d, got %d", cfg.Threads, sanitized.Threads)
	}
	if cfg.User != sanitized.User {
		t.Errorf("Expected User %q, got %q", cfg.User, sanitized.User)
	}
	if cfg.OutputDir != sanitized.OutputDir {
		t.Errorf("Expected OutputDir %q, got %q", cfg.OutputDir, sanitized.OutputDir)
	}
	if cfg.Timeout != sanitized.Timeout {
		t.Errorf("Expected Timeout %d, got %d", cfg.Timeout, sanitized.Timeout)
	}
	if cfg.IncludeGists != sanitized.IncludeGists {
		t.Errorf("Expected IncludeGists %v, got %v", cfg.IncludeGists, sanitized.IncludeGists)
	}
}
