// internal/filesystem/atomic.go

package filesystem

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// AtomicWriteFile safely replaces a file using a temporary file and an
// atomic rename.
//
// The temporary file is created in the destination directory to ensure
// the rename operation is atomic on the same filesystem.
//
// On success, the destination file is either completely replaced or left
// untouched if an error occurs.
func AtomicWriteFile(
	path string,
	perm os.FileMode,
	write func(io.Writer) error,
) error {

	dir := filepath.Dir(path)

	tmp, err := os.CreateTemp(
		dir,
		".gitback-*",
	)

	if err != nil {
		return fmt.Errorf(
			"create temporary file: %w",
			err,
		)
	}

	tmpPath := tmp.Name()

	// Remove the temporary file unless the rename succeeds.
	defer os.Remove(tmpPath)

	if err := tmp.Chmod(perm); err != nil {

		tmp.Close()

		return fmt.Errorf(
			"set permissions on temporary file: %w",
			err,
		)
	}

	if err := write(tmp); err != nil {

		tmp.Close()

		return err
	}

	// Flush file contents to disk before replacing the destination.
	if err := tmp.Sync(); err != nil {

		tmp.Close()

		return fmt.Errorf(
			"sync temporary file: %w",
			err,
		)
	}

	if err := tmp.Close(); err != nil {

		return fmt.Errorf(
			"close temporary file: %w",
			err,
		)
	}

	if err := os.Rename(
		tmpPath,
		path,
	); err != nil {

		return fmt.Errorf(
			"replace destination file: %w",
			err,
		)
	}

	// Flush the directory metadata so the rename itself becomes durable.
	dirHandle, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf(
			"open parent directory: %w",
			err,
		)
	}

	defer dirHandle.Close()

	if err := dirHandle.Sync(); err != nil {

		return fmt.Errorf(
			"sync parent directory: %w",
			err,
		)
	}

	return nil
}
