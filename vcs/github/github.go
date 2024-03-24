package github

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/v59/github"
)

func ListPublicRepos(username string) ([]*github.Repository, *github.Response) {
	var allRepos []*github.Repository
	var rateInfo *github.Response

	ctx := context.Background()
	client := github.NewClient(nil)
	opt := &github.RepositoryListByUserOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		repos, resp, err := client.Repositories.ListByUser(ctx, username, opt)

		if err != nil {
			fmt.Printf("Error listing repositories: %v\n", err)
			os.Exit(1)
		}

		allRepos, rateInfo = append(allRepos, repos...), resp

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return allRepos, rateInfo
}

func ListPrivateRepos() ([]*github.Repository, *github.Response) {
	var allRepos []*github.Repository
	var rateInfo *github.Response

	ctx := context.Background()
	client := AuthenticateClientWithPAT()
	opt := &github.RepositoryListByAuthenticatedUserOptions{
		Visibility:  "all",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		repos, resp, err := client.Repositories.ListByAuthenticatedUser(ctx, opt)

		if err != nil {
			fmt.Printf("Error listing repositories: %v\n", err)
			os.Exit(1)
		}

		allRepos, rateInfo = append(allRepos, repos...), resp

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return allRepos, rateInfo
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

func LogResponse(rateInfo *github.Response) {
	if !rateInfo.TokenExpiration.IsZero() {
		fmt.Println("Token Expiration:", rateInfo.TokenExpiration)
	}

	fmt.Println("Rate:", rateInfo.Rate)
}
