// internal/state/inventory.go

package state

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/flarexes/gitback/internal/filesystem"
)

func WriteInventory(path string, items []string) error {

	return filesystem.AtomicWriteFile(
		path,
		0600,
		func(w io.Writer) error {

			for _, item := range items {

				if _, err := fmt.Fprintln(w, item); err != nil {

					return fmt.Errorf(
						"write inventory file %s: %w",
						path,
						err,
					)
				}
			}

			return nil
		},
	)
}

func ReadInventory(path string) ([]string, error) {

	file, err := os.Open(path)

	if os.IsNotExist(err) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	defer file.Close()

	var items []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {

		item := strings.TrimSpace(scanner.Text())

		if item == "" {
			continue
		}

		items = append(items, item)
	}

	if err := scanner.Err(); err != nil {

		return nil, fmt.Errorf(
			"scan inventory file %s: %w",
			path,
			err,
		)
	}

	return items, nil
}
