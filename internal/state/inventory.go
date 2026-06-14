package state

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func WriteInventory(path string, items []string) error {

	f, err := os.OpenFile(
		path,
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0600,
	)

	if err != nil {
		return fmt.Errorf("open inventory file %s: %w",
			path,
			err,
		)
	}

	defer f.Close()

	for _, item := range items {

		if _, err := fmt.Fprintln(f, item); err != nil {

			return fmt.Errorf("write inventory file %s: %w",
				path,
				err,
			)
		}
	}

	return nil
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
