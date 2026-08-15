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
		e.layout.TempDir,
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

// gitEnv builds the environment for a git subprocess, injecting the
// GitBack-managed token and disabling interactive prompts.
//
// Every key we set here is first stripped from the inherited
// environment before we append our own value.
func (e *Engine) gitEnv(askPass string) []string {

	token, _ := config.ReadToken(e.layout)

	env := os.Environ()
	env = filterEnv(env, "GITBACK_TOKEN", "GIT_ASKPASS", "GIT_TERMINAL_PROMPT")

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

// filterEnv returns env with any entries for the given keys removed.
// Keys are compared exactly as they appear before "=", matching how
// os.Environ() formats entries.
func filterEnv(env []string, keys ...string) []string {

	filtered := env[:0]

	for _, kv := range env {

		skip := false

		for _, key := range keys {

			prefix := key + "="

			if len(kv) >= len(prefix) && kv[:len(prefix)] == prefix {
				skip = true
				break
			}
		}

		if !skip {
			filtered = append(filtered, kv)
		}
	}

	return filtered
}
