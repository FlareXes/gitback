# GitBack

GitBack is a production-ready tool designed to backup GitHub repositories and gists with support for both authenticated and unauthenticated access. It's built with Go and follows best practices for configuration management, error handling, and code organization.

## Features

- **Flexible Authentication**: Backup with or without GitHub authentication
- **Comprehensive Backup**: Supports repositories and gists (public and private with token)
- **Concurrent Operations**: Configurable number of concurrent operations
- **Production Ready**: Proper error handling, logging, and configuration management
- **Configurable**: Multiple ways to configure the tool (flags, environment variables)
- **Rate Limiting**: Built-in rate limit handling for GitHub API
- **Docker Support**: Easy deployment using Docker
- **CI/CD Ready**: GitHub Actions workflow for automated testing and deployment

## Installation

### Prerequisites

- Go 1.21 or higher (for building from source)
- Git (for repository cloning)
- Docker (for containerized deployment)

### Building from Source

1. Clone the repository:
   ```bash
   git clone https://github.com/flarexes/gitback.git
   cd gitback
   ```

2. Build the binary:
   ```bash
   make build
   ```

3. (Optional) Install to your GOPATH:
   ```bash
   go install ./cmd/gitback
   ```

### Using Docker

```bash
# Build the Docker image
docker build -t gitback .

# Run with environment variables
docker run --rm -e GITHUB_TOKEN=your_token -e GITBACK_USER=your_username -v $(pwd)/backups:/data gitback
```

### Using Docker Compose

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```
2. Edit `.env` with your configuration
3. Run with Docker Compose:
   ```bash
   docker-compose up -d
   ```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GITHUB_TOKEN` | GitHub Personal Access Token | (none) |
| `GITBACK_NOAUTH` | Run without authentication | `false` |
| `GITBACK_THREADS` | Number of concurrent operations | `5` |
| `GITBACK_USER` | GitHub username (required in no-auth mode) | (none) |
| `GITBACK_OUTPUT_DIR` | Directory to store backups | `/data` (container) / `./data` (local) |
| `GITBACK_TIMEOUT` | API request timeout in seconds | `30` |
| `GITBACK_INCLUDE_GISTS` | Whether to include gists in backup | `true` |
| `GITBACK_LOG_LEVEL` | Log level (debug, info, warn, error) | `info` |
| `GITBACK_LOG_FORMAT` | Log format (text, json) | `text` |

### Command Line Flags

```
Usage of gitback:
  -no-gists
        Skip backing up gists
  -noauth
        Disable GitHub Auth (limited to 60 requests/hour for public data only)
  -output-dir string
        Directory to store backups (default "~/gitbackup")
  -thread int
        Maximum number of concurrent connections (default 5)
  -timeout int
        Timeout in seconds for API requests (default 30)
  -token string
        GitHub API token (can also be set via GITHUB_TOKEN)
  -username string
        GitHub username (required when --noauth is set)
  -version
        Show version information and exit
```

## Usage Examples

### Backup Public Repositories (No Authentication)

```bash
./gitback -noauth -username flarexes
```

### Backup Private and Public Repositories (With Authentication)

```bash
# Using environment variable
export GITHUB_TOKEN=your_github_token
./gitback

# Or using command line flag
./gitback -token your_github_token
```

### Custom Output Directory and Concurrency

```bash
./gitback -output-dir /path/to/backup -thread 10
```

### Skip Gists

```bash
./gitback -no-gists
```

## Project Structure

```
gitback/
├── cmd/
│   └── gitback/         # Main application entry point
├── internal/
│   ├── types/          # Shared type definitions
│   └── vcs/            # Version control system interfaces and implementations
│       └── github/     # GitHub-specific implementation
└── pkg/
    └── config/         # Configuration management
```

## Development

### Building and Testing

```bash
# Run tests
go test ./...

# Build with debug information
go build -ldflags="-w -s" -o gitback ./cmd/gitback
```

### Code Style

This project follows the standard Go code style. Please run `gofmt` and `golint` before submitting changes.

## Contributing

Contributions are welcome! Please feel free to submit a pull request, I want to take this project futher.

## Issues

If you encounter any issues or have suggestions for improvements, please open an issue on the [GitHub repository](https://github.com/flarexes/gitback/issues).


## License

This project is licensed under the BSD-3-Clause license. For more information, please see the [LICENSE](LICENSE) file.
