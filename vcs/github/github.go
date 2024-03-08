package github

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/v59/github"
)

func ListPublicRepos(username string) ([]*github.Repository, *github.Response) {
	ctx := context.Background()
	client := github.NewClient(nil)

	repos, resp, err := client.Repositories.ListByUser(ctx, username, nil)

	if err != nil {
		fmt.Printf("Error listing repositories: %v\n", err)
		os.Exit(1)
	}

	return repos, resp
}

func ListPrivateRepos() ([]*github.Repository, *github.Response) {
	ctx := context.Background()
	opt := &github.RepositoryListByAuthenticatedUserOptions{
		Visibility: "all",
	}

	client := AuthenticateClientWithPAT()
	repos, resp, err := client.Repositories.ListByAuthenticatedUser(ctx, opt)

	if err != nil {
		fmt.Printf("Error listing repositories: %v\n", err)
		os.Exit(1)
	}

	return repos, resp
}

func AuthenticateClientWithPAT() *github.Client {
	pat := GetGitHubPAT()
	client := github.NewClient(nil).WithAuthToken(pat)

	return client
}

func GetPatUrl(fullname string) string {
	pat := GetGitHubPAT()
	url := fmt.Sprintf("https://%s@github.com/%s.git", pat, fullname)

	return url
}

func GetGitHubPAT() string {
	pat := os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN")

	if pat == "" {
		log.Fatal("GITHUB_PERSONAL_ACCESS_TOKEN environment variable does not exist")
	}

	return pat
}

func LogResponse(resp *github.Response) {
	if !resp.TokenExpiration.IsZero() {
		fmt.Println("Token Expiration:", resp.TokenExpiration)
	}

	fmt.Println("Rate:", resp.Rate)
}
