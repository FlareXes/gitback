// internal/runtime/layout.go

package runtime

import (
	"os"
	"path/filepath"
)

// Layout is the complete, fixed filesystem geography for GitBack.
// None of these paths are user-configurable — they exist regardless of
// what a user puts in config.toml. Anything a user CAN configure (mirror
// root, snapshot output dir, workers, etc.) lives in config.Config instead.
type Layout struct {
	ConfigDir  string
	ConfigFile string

	DataDir  string
	StateDir string
	LogDir   string
	LogFile  string

	TokenFile string
	LockFile  string
	TempDir   string

	MirrorsStateFile        string
	RepositoryInventoryFile string
	GistInventoryFile       string
}

// New resolves Layout from the OS home directory (XDG-style conventions).
func New() (Layout, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Layout{}, err
	}
	return newFromRoot(home, os.TempDir()), nil
}

// NewWithRoot builds a Layout rooted under a custom directory instead of
// the real $HOME. Intended for tests: pass t.TempDir().
func NewWithRoot(root string) Layout {
	return newFromRoot(root, root)
}

func newFromRoot(home, tmp string) Layout {
	configDir := filepath.Join(home, ".config", "gitback")
	dataDir := filepath.Join(home, ".local", "share", "gitback")
	stateDir := filepath.Join(dataDir, "state")
	logDir := filepath.Join(home, ".local", "state", "gitback")

	return Layout{
		ConfigDir:  configDir,
		ConfigFile: filepath.Join(configDir, "config.toml"),

		DataDir:  dataDir,
		StateDir: stateDir,
		LogDir:   logDir,
		LogFile:  filepath.Join(logDir, "gitback.log"),

		TokenFile: filepath.Join(stateDir, "github.token"),
		LockFile:  filepath.Join(tmp, "gitback.lock"),
		TempDir:   filepath.Join(stateDir, "tmp"),

		MirrorsStateFile:        filepath.Join(stateDir, "mirrors.json"),
		RepositoryInventoryFile: filepath.Join(stateDir, "repositories.txt"),
		GistInventoryFile:       filepath.Join(stateDir, "gists.txt"),
	}
}

// EnsureDirs creates every directory this Layout needs. Missing directories
// are treated as recoverable — commands like `gitback sync` should work
// again even if state/gitback dirs were deleted out from under GitBack.
func (l Layout) EnsureDirs() error {
	dirs := []string{
		l.ConfigDir,
		l.DataDir,
		l.StateDir,
		l.LogDir,
		l.TempDir,
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
	}
	return nil
}
