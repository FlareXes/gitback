// internal/config/config.go

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flarexes/gitback/internal/runtime"
	"github.com/spf13/viper"
)

type Config struct {
	GitHub   GitHubConfig
	Storage  StorageConfig
	Sync     SyncConfig
	Snapshot SnapshotConfig
	Health   HealthConfig
}

type GitHubConfig struct {
	BackupGists bool `mapstructure:"backup_gists"`
}

type StorageConfig struct {
	MirrorRoot string `mapstructure:"mirror_root"`
}

type SnapshotConfig struct {
	OutputDirectory string `mapstructure:"output_directory"`
	Retention       int    `mapstructure:"retention"`
}

type SyncConfig struct {
	Workers       int `mapstructure:"workers"`
	RetryAttempts int `mapstructure:"retry_attempts"`
}

type HealthConfig struct {
	MinimumFreeDiskPercent uint8 `mapstructure:"minimum_free_disk_percent"`
}

// RepositoryMirrorRoot, GistMirrorRoot, and QuarantineDir are DERIVED from
// the user-configured MirrorRoot.
func (c Config) RepositoryMirrorRoot() string {
	return filepath.Join(c.Storage.MirrorRoot, "repositories")
}

func (c Config) GistMirrorRoot() string {
	return filepath.Join(c.Storage.MirrorRoot, "gists")
}

func (c Config) QuarantineDir() string {
	return filepath.Join(filepath.Dir(c.Storage.MirrorRoot), "quarantine")
}

// Default returns the default configuration, given a resolved Layout so
// that default mirror/snapshot paths sit alongside GitBack's other data.
func Default(layout runtime.Layout) Config {
	return Config{
		GitHub: GitHubConfig{BackupGists: true},
		Storage: StorageConfig{
			MirrorRoot: filepath.Join(layout.DataDir, "mirrors"),
		},
		Snapshot: SnapshotConfig{
			OutputDirectory: filepath.Join(layout.DataDir, "snapshots"),
			Retention:       0,
		},
		Sync: SyncConfig{
			Workers:       3,
			RetryAttempts: 3,
		},
		Health: HealthConfig{
			MinimumFreeDiskPercent: 20,
		},
	}
}

// Write writes the GitBack configuration file.
func Write(path string, cfg Config) error {
	content := fmt.Sprintf(`# GitBack configuration

[github]
backup_gists = %t

[storage]
mirror_root = %q

[snapshot]
output_directory = %q
retention = %d

[sync]
workers = %d
retry_attempts = %d

[health]
minimum_free_disk_percent = %d
`,
		cfg.GitHub.BackupGists,
		cfg.Storage.MirrorRoot,
		cfg.Snapshot.OutputDirectory,
		cfg.Snapshot.Retention,
		cfg.Sync.Workers,
		cfg.Sync.RetryAttempts,
		cfg.Health.MinimumFreeDiskPercent,
	)

	return os.WriteFile(path, []byte(content), 0600)
}

// Load reads and validates configuration using the given Layout to locate
// config.toml. It never falls back to defaults — a missing config file is
// a hard error, since GitBack shouldn't silently run on unconfigured
// defaults. Run `gitback init` to create one.
func Load(layout runtime.Layout) (*Config, error) {

	if _, err := os.Stat(layout.ConfigFile); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf(
				"config file not found at %s; run `gitback init`",
				layout.ConfigFile,
			)
		}
		return nil, err
	}

	cfg := Default(layout)

	if err := ReadConfig(layout, &cfg); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func ReadConfig(layout runtime.Layout, cfg *Config) error {
	v := viper.New()
	v.SetConfigFile(layout.ConfigFile)
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return err
	}

	return v.Unmarshal(cfg)
}

// ReadToken reads the GitHub token from env or the layout's token file.
func ReadToken(layout runtime.Layout) (string, error) {
	if token := strings.TrimSpace(os.Getenv("GITBACK_TOKEN")); token != "" {
		return token, nil
	}

	data, err := os.ReadFile(layout.TokenFile)
	if err != nil {
		return "", err
	}

	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", fmt.Errorf(
			"github token not configured; either:\n" +
				"  • set GITBACK_TOKEN\n" +
				"  • run: gitback init",
		)
	}

	return token, nil
}
