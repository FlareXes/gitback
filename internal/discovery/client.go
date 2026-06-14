// internal/discovery/client.go

package discovery

import (
	"fmt"

	"github.com/flarexes/gitback/internal/config"
	"github.com/flarexes/gitback/internal/logging"
	"github.com/google/go-github/v88/github"
)

type Client struct {
	cfg    *config.Config
	logger *logging.Logger
	api    *github.Client
}

func New(cfg *config.Config, logger *logging.Logger) (*Client, error) {

	if cfg.GitHubToken == "" {
		return nil, fmt.Errorf("github token not configured; run: gitback init")
	}

	api, err := github.NewClient(
		github.WithAuthToken(
			cfg.GitHubToken,
		),
	)

	if err != nil {
		return nil, err
	}

	return &Client{
		cfg:    cfg,
		logger: logger,
		api:    api,
	}, nil
}
