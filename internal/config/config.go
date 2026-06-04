// internal/config/config.go

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	GitHubOwner string `mapstructure:"github_owner"`

	// `GitHubToken` loaded from dedicated token file located at `TokenFile`
	GitHubToken string
	TokenFile   string `mapstructure:"token_file"`

	BaseDir     string `mapstructure:"base_dir"`
	ConfigDir   string `mapstructure:"config_dir"`
	MirrorDir   string `mapstructure:"mirror_dir"`
	SnapshotDir string `mapstructure:"snapshot_dir"`
	StateDir    string `mapstructure:"state_dir"`
	LogDir      string `mapstructure:"log_dir"`
	LogFile     string `mapstructure:"log_file"`
	HealthDir   string `mapstructure:"health_dir"`
	TempDir     string `mapstructure:"temp_dir"`

	RepoInventory string `mapstructure:"repo_inventory"`
	LockFile      string `mapstructure:"lock_file"`

	CooldownMinSeconds int `mapstructure:"cooldown_min_seconds"`
	CooldownMaxSeconds int `mapstructure:"cooldown_max_seconds"`

	MinimumFreeDiskPercent int `mapstructure:"minimum_free_disk_percent"`
}

func Default() Config {
	base := "/tmp/lib/gitback"

	return Config{
		GitHubOwner: "CHANGE_ME",
		TokenFile:   filepath.Join(base, "state", "github.token"),

		BaseDir:     base,
		ConfigDir:   "/tmp/gitback",
		MirrorDir:   filepath.Join(base, "mirrors"),
		SnapshotDir: filepath.Join(base, "snapshots"),
		StateDir:    filepath.Join(base, "state"),
		LogDir:      "/tmp/log/gitback",
		LogFile:     "/tmp/log/gitback/gitback.log",
		HealthDir:   "/tmp/log/gitback/health",
		TempDir:     filepath.Join(base, "tmp"),

		RepoInventory: filepath.Join(base, "state", "repositories.txt"),
		LockFile:      filepath.Join(base, "state", "gitback.lock"),

		CooldownMinSeconds: 60,
		CooldownMaxSeconds: 120,

		MinimumFreeDiskPercent: 15,
	}
}

func Load() (*Config, error) {
	cfg := Default()

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(cfg.ConfigDir)

	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if data, err := os.ReadFile(cfg.TokenFile); err == nil {
		cfg.GitHubToken = strings.TrimSpace(string(data))
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func Write(path string, cfg Config) error {

	content := fmt.Sprintf(`github_owner: %s
token_file: %s

base_dir: %s
config_dir: %s
mirror_dir: %s
snapshot_dir: %s
state_dir: %s
log_dir: %s
health_dir: %s
temp_dir: %s

repo_inventory: %s
log_file: %s
lock_file: %s

cooldown_min_seconds: %d
cooldown_max_seconds: %d

minimum_free_disk_percent: %d
`,
		cfg.GitHubOwner,
		cfg.TokenFile,

		cfg.BaseDir,
		cfg.ConfigDir,
		cfg.MirrorDir,
		cfg.SnapshotDir,
		cfg.StateDir,
		cfg.LogDir,
		cfg.HealthDir,
		cfg.TempDir,

		cfg.RepoInventory,
		cfg.LogFile,
		cfg.LockFile,

		cfg.CooldownMinSeconds,
		cfg.CooldownMaxSeconds,

		cfg.MinimumFreeDiskPercent,
	)

	return os.WriteFile(
		path,
		[]byte(content),
		0600,
	)
}
