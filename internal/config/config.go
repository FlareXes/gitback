// internal/config/config.go

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flarexes/gitback/internal/logging"
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

	Retention int `mapstructure:"retention"`
}

type SyncConfig struct {
	Workers int `mapstructure:"workers"`

	RetryAttempts int `mapstructure:"retry_attempts"`
}

type HealthConfig struct {
	MinimumFreeDiskPercent int `mapstructure:"minimum_free_disk_percent"`
}

// Default returns the default configuration.
func Default() Config {
	home, _ := os.UserHomeDir()

	return Config{
		GitHub: GitHubConfig{
			BackupGists: true,
		},

		Storage: StorageConfig{
			MirrorRoot: filepath.Join(
				home,
				".local",
				"share",
				"gitback",
				"mirrors",
			),
		},

		Snapshot: SnapshotConfig{
			OutputDirectory: filepath.Join(
				home,
				".local",
				"share",
				"gitback",
				"snapshots",
			),

			Retention: 0,
		},

		Sync: SyncConfig{
			Workers: 3,

			RetryAttempts: 3,
		},

		Health: HealthConfig{
			MinimumFreeDiskPercent: 20,
		},
	}
}

// ----------------------------------
// Private helper functions.
// ----------------------------------

// configDirectory returns the configuration directory.
// Because configuration location is not configurable, it is hardcoded.
func configDirectory() string {

	home, err := os.UserHomeDir()

	if err != nil {
		return "."
	}

	return filepath.Join(
		home,
		".config",
		"gitback",
	)
}

// StateDir returns the state directory.
// Because state location is not configurable, it is hardcoded.
func StateDir() string {

	home, err := os.UserHomeDir()

	if err != nil {
		home = "."
	}

	return filepath.Join(
		home,
		".local",
		"share",
		"gitback",
		"state",
	)
}

// logDir returns the log directory.
// Because log location is not configurable, it is hardcoded.
func logDir() string {

	home, err := os.UserHomeDir()

	if err != nil {
		home = "."
	}

	return filepath.Join(
		home,
		".local",
		"state",
		"gitback",
	)
}

// ----------------------------------
// Public methods.
// ----------------------------------

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

	return os.WriteFile(
		path,
		[]byte(content),
		0600,
	)
}

func Load() (*Config, error) {

	cfg := Default()

	if err := ReadConfig(&cfg); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func ReadConfig(cfg *Config) error {

	v := viper.New()

	v.SetConfigFile(ConfigFile())

	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return err
	}

	return v.Unmarshal(cfg)
}

func ReadToken() (string, error) {

	data, err := os.ReadFile(TokenFile())

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

func ConfigFile() string {

	return filepath.Join(
		configDirectory(),
		"config.toml",
	)
}

func LogFile() string {

	return filepath.Join(
		logDir(),
		"gitback.log",
	)
}

func TokenFile() string {

	return filepath.Join(
		StateDir(),
		"github.token",
	)
}

func MirrorsStateFile() string {

	return filepath.Join(
		StateDir(),
		"mirrors.json",
	)
}

func RepositoryInventoryFile() string {

	return filepath.Join(
		StateDir(),
		"repositories.txt",
	)
}

func GistInventoryFile() string {

	return filepath.Join(
		StateDir(),
		"gists.txt",
	)
}

func LockFile() string {

	return filepath.Join(
		os.TempDir(),
		"gitback.lock",
	)
}

func TempDir() string {

	return filepath.Join(
		StateDir(),
		"tmp",
	)
}

// --------------------------------------
// Helper functions.
// --------------------------------------

// EnsureRuntimeDirectories creates all runtime directories required by GitBack.
//
// Missing directories are treated as a recoverable condition.
// This allows commands such as:
//
//	gitback sync
//
// to work even if:
//
//	~/.local/share/gitback/mirrors
//	~/.local/share/gitback/state
//	~/.local/share/gitback/tmp
//	~/.local/share/gitback/snapshots
//
// were accidentally removed.
func (cfg *Config) EnsureRuntimeDirectories() error {

	dirs := []string{
		cfg.Storage.MirrorRoot,

		cfg.Snapshot.OutputDirectory,

		TempDir(),

		StateDir(),

		logDir(),
	}

	logger, err := logging.New(LogFile())
	if err != nil {
		return err
	}
	defer logger.Close()

	for _, dir := range dirs {

		if _, err := os.Stat(dir); os.IsNotExist(err) {

			fmt.Println("[WARN] Recreated missing directory: ", dir)

			logger.Warn(
				logging.Events.Filesystem.DirectoryRecreated,
				"",
				fmt.Sprintf("Recreated missing directory: %s", dir),
			)
		}

		if err := os.MkdirAll(dir, 0700); err != nil {

			return fmt.Errorf(
				"create directory %s: %w",
				dir,
				err,
			)
		}
	}

	return nil
}
