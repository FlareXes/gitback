// internal/discovery/client.go

package discovery

import (
	"fmt"

	"github.com/flarexes/gitback/internal/config"
	"github.com/flarexes/gitback/internal/logging"
	"github.com/flarexes/gitback/internal/runtime"
	"github.com/google/go-github/v88/github"
)

type Client struct {
	cfg    *config.Config
	layout runtime.Layout
	logger *logging.Logger
	api    *github.Client
}

func New(cfg *config.Config, layout runtime.Layout, logger *logging.Logger) (*Client, error) {
	token, err := config.ReadToken(layout)
	if err != nil {
		return nil, fmt.Errorf("read github token: %w", err)
	}

	api, err := github.NewClient(github.WithAuthToken(token))
	if err != nil {
		return nil, err
	}

	return &Client{cfg: cfg, layout: layout, logger: logger, api: api}, nil
}
