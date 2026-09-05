// Package example provides embedded example configuration for gitback.

package example

import _ "embed"

//go:embed config.example.toml
var Config string
