package mirror

import (
	"os"

	"github.com/flarexes/gitback/internal/config"
)

func (e *Engine) createAskPassScript() (string, error) {

	content := `#!/bin/sh

case "$1" in
*Username*)
	echo "oauth2"
	;;
*Password*)
	echo "$GITBACK_TOKEN"
	;;
esac
`

	file, err := os.CreateTemp(
		config.TempDir(),
		"gitback-askpass-*",
	)
	if err != nil {
		return "", err
	}

	if _, err := file.WriteString(content); err != nil {
		file.Close()
		return "", err
	}

	file.Close()

	if err := os.Chmod(file.Name(), 0700); err != nil {
		return "", err
	}

	return file.Name(), nil
}

func (e *Engine) gitEnv(askPass string) []string {

	token, _ := config.ReadToken()

	env := os.Environ()

	env = append(
		env,
		"GIT_ASKPASS="+askPass,
	)

	env = append(
		env,
		"GITBACK_TOKEN="+token,
	)

	env = append(
		env,
		"GIT_TERMINAL_PROMPT=0",
	)

	return env
}
