// internal/config/config.go

package config

import (
	"errors"
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

	LockFile string `mapstructure:"lock_file"`

	MinimumFreeDiskPercent int `mapstructure:"minimum_free_disk_percent"`

	GitRetryAttempts int `mapstructure:"git_retry_attempts"`
	SyncWorkers      int `mapstructure:"sync_workers"`

	SnapshotRetention int `mapstructure:"snapshot_retention"`

	BackupGists bool `mapstructure:"backup_gists"`
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

		LockFile: "/tmp/gitback.lock",

		MinimumFreeDiskPercent: 20,

		GitRetryAttempts: 3,
		SyncWorkers:      3,

		SnapshotRetention: 0, // if <= 0: disabled

		BackupGists: true,
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

		var notFound viper.ConfigFileNotFoundError

		if errors.As(err, &notFound) {

			return nil, fmt.Errorf("gitback is not initialized\n\nRun: gitback init")
		}

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

lock_file: %s

minimum_free_disk_percent: %d

git_retry_attempts: %d
sync_workers: %d

snapshot_retention: %d

backup_gists: %t
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

		cfg.LockFile,

		cfg.MinimumFreeDiskPercent,

		cfg.GitRetryAttempts,
		cfg.SyncWorkers,

		cfg.SnapshotRetention,

		cfg.BackupGists,
	)

	return os.WriteFile(
		path,
		[]byte(content),
		0600,
	)
}

func (c *Config) RepositoryMirrorDir() string {
	return filepath.Join(
		c.MirrorDir,
		"repositories",
	)
}

func (c *Config) GistMirrorDir() string {
	return filepath.Join(
		c.MirrorDir,
		"gists",
	)
}

func (c *Config) RepositoryInventoryFile() string {
	return filepath.Join(
		c.StateDir,
		"repositories.txt",
	)
}

func (c *Config) GistInventoryFile() string {
	return filepath.Join(
		c.StateDir,
		"gists.txt",
	)
}
