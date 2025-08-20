// Package types contains shared types used across the application.
package types

// Config holds the application configuration.
type Config struct {
	// NoAuth indicates whether to run in no-authentication mode
	NoAuth bool
	// Threads specifies the number of concurrent operations
	Threads int
	// Token is the GitHub Personal Access Token
	Token string
	// User is the GitHub username
	User string
	// OutputDir is the directory where backups will be stored
	OutputDir string
	// Timeout is the timeout for API requests in seconds
	Timeout int
	// IncludeGists specifies whether to include gists in the backup
	IncludeGists bool
}
