// internal/filesystem/directory.go

package filesystem

import (
	"errors"
	"fmt"
	"os"
)

const (
	// Default permissions used for directories
	DefaultDirectoryMode os.FileMode = 0700
)

// EnsureDirectory makes sure a directory exists.
//
// If the directory does not exist, it is created using the default
// GitBack directory permissions.
//
// Returns:
//
//	created == true
//
// when the directory had to be created.
//
// The returned error always includes the directory path for easier
// troubleshooting.
func CreateDirectory(path string) (created bool, err error) {

	info, err := os.Stat(path)

	switch {

	// Directory already exists.
	case err == nil:

		if !info.IsDir() {

			return false, fmt.Errorf(
				"%q exists but is not a directory",
				path,
			)
		}

		return false, nil

	// Directory missing.
	case errors.Is(err, os.ErrNotExist):

		if err := os.MkdirAll(path, DefaultDirectoryMode); err != nil {

			if errors.Is(err, os.ErrPermission) {

				return false, fmt.Errorf(
					"cannot create directory %q: permission denied (check parent directory permissions and ownership)",
					path,
				)
			}

			return false, fmt.Errorf(
				"create directory %q: %w",
				path,
				err,
			)
		}

		return true, nil

	default:

		return false, fmt.Errorf(
			"access directory %q: %w",
			path,
			err,
		)
	}
}
