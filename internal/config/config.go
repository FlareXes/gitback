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

	// Loaded from TokenFile during Load().
	GitHubToken string

	TokenFile string `mapstructure:"token_file"`

	ConfigDir string `mapstructure:"config_dir"`

	DataDir string `mapstructure:"data_dir"`

	StateDir         string `mapstructure:"state_dir"`
	MirrorsStateFile string `mapstructure:"mirrors_state_file"`

	MirrorDir   string `mapstructure:"mirror_dir"`
	SnapshotDir string `mapstructure:"snapshot_dir"`

	LogDir  string `mapstructure:"log_dir"`
	LogFile string `mapstructure:"log_file"`

	TempDir string `mapstructure:"temp_dir"`

	RepoInventory string `mapstructure:"repo_inventory"`
	LockFile      string `mapstructure:"lock_file"`

	CooldownMinSeconds int `mapstructure:"cooldown_min_seconds"`
	CooldownMaxSeconds int `mapstructure:"cooldown_max_seconds"`

	MinimumFreeDiskPercent int `mapstructure:"minimum_free_disk_percent"`

	GitRetryAttempts int `mapstructure:"retry_attempts"`

	SnapshotRetention int `mapstructure:"snapshot_retention"`
}

func Default() Config {

	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	configDir := filepath.Join(
		home,
		".config",
		"gitback",
	)

	dataDir := filepath.Join(
		home,
		".local",
		"share",
		"gitback",
	)

	logDir := filepath.Join(
		home,
		".local",
		"state",
		"gitback",
	)

	return Config{

		ConfigDir: configDir,

		DataDir: dataDir,

		LogDir: logDir,

		LogFile: filepath.Join(
			logDir,
			"gitback.log",
		),

		StateDir: filepath.Join(
			dataDir,
			"state",
		),

		MirrorDir: filepath.Join(
			dataDir,
			"mirrors",
		),

		SnapshotDir: filepath.Join(
			dataDir,
			"snapshots",
		),

		MirrorsStateFile: filepath.Join(
			dataDir,
			"state",
			"mirrors.json",
		),

		TokenFile: filepath.Join(
			dataDir,
			"state",
			"github.token",
		),

		TempDir: filepath.Join(
			dataDir,
			"tmp",
		),

		RepoInventory: filepath.Join(
			dataDir,
			"state",
			"repositories.txt",
		),

		LockFile: "/tmp/gitback.lock",

		CooldownMinSeconds: 0,
		CooldownMaxSeconds: 5,

		MinimumFreeDiskPercent: 20,

		GitRetryAttempts: 3,

		SnapshotRetention: 0, // if <= 0: disabled
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

		cfg.GitHubToken = strings.TrimSpace(
			string(data),
		)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func Write(path string, cfg Config) error {

	content := fmt.Sprintf(`token_file: %s

config_dir: %s

data_dir: %s

state_dir: %s
mirrors_state_file: %s

mirror_dir: %s
snapshot_dir: %s

log_dir: %s
log_file: %s

temp_dir: %s

repo_inventory: %s
lock_file: %s

cooldown_min_seconds: %d
cooldown_max_seconds: %d

minimum_free_disk_percent: %d

git_retry_attempts: %d

snapshot_retention: %d
`,
		cfg.TokenFile,

		cfg.ConfigDir,

		cfg.DataDir,

		cfg.StateDir,
		cfg.MirrorsStateFile,

		cfg.MirrorDir,
		cfg.SnapshotDir,

		cfg.LogDir,
		cfg.LogFile,

		cfg.TempDir,

		cfg.RepoInventory,
		cfg.LockFile,

		cfg.CooldownMinSeconds,
		cfg.CooldownMaxSeconds,

		cfg.MinimumFreeDiskPercent,

		cfg.GitRetryAttempts,

		cfg.SnapshotRetention,
	)

	return os.WriteFile(
		path,
		[]byte(content),
		0600,
	)
}
