package vcs

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/v59/github"
)

const PER_PAGE = 100

func getGitHubPAT() string {
	pat := os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN")
	if pat == "" {
		fmt.Println("[-] GITHUB_PERSONAL_ACCESS_TOKEN environment variable does not exist")
		fmt.Println("Tip: use --noauth to disable auth, --username is required on auth disable")
		os.Exit(1)
	}
	return pat
}

func authenticateClientWithPAT() *github.Client {
	pat := getGitHubPAT()
	return github.NewClient(nil).WithAuthToken(pat)
}

func LogResponse(rateInfo *github.Response) {
	if !rateInfo.TokenExpiration.IsZero() {
		fmt.Println("Token Expiration:", rateInfo.TokenExpiration)
	}

	fmt.Println("Rate:", rateInfo.Rate)
}

func GetPatUrl(fullname string) (string, string) {
	pat := getGitHubPAT()
	rUrl := fmt.Sprintf("https://%s@github.com/%s.git", pat, fullname)
	wUrl := fmt.Sprintf("https://%s@github.com/%s.wiki.git", pat, fullname)

	return rUrl, wUrl
}

func ListPublicRepos(username string) ([]*github.Repository, *github.Response) {
	var allRepos []*github.Repository
	var rateInfo *github.Response

	ctx := context.Background()
	client := github.NewClient(nil)
	opt := &github.RepositoryListByUserOptions{
		ListOptions: github.ListOptions{PerPage: PER_PAGE},
	}

	for {
		repos, resp, err := client.Repositories.ListByUser(ctx, username, opt)
		if err != nil {
			log.Fatal("Error listing repositories: ", err)
		}
		LogResponse(resp)

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
	client := authenticateClientWithPAT()
	opt := &github.RepositoryListByAuthenticatedUserOptions{
		Visibility:  "all",
		ListOptions: github.ListOptions{PerPage: PER_PAGE},
	}

	for {
		repos, resp, err := client.Repositories.ListByAuthenticatedUser(ctx, opt)
		if err != nil {
			log.Fatal("Error listing repositories: ", err)
		}
		LogResponse(resp)

		allRepos, rateInfo = append(allRepos, repos...), resp
		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}
	return allRepos, rateInfo
}

func ListPublicGists(username string) ([]*github.Gist, *github.Response) {
	var allGists []*github.Gist
	var rateInfo *github.Response

	ctx := context.Background()
	client := github.NewClient(nil)
	opt := &github.GistListOptions{
		ListOptions: github.ListOptions{PerPage: PER_PAGE},
	}

	for {
		gists, resp, err := client.Gists.List(ctx, username, opt)
		if err != nil {
			log.Fatal("Error listing gists: ", err)
		}
		LogResponse(resp)

		allGists, rateInfo = append(allGists, gists...), resp
		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}
	return allGists, rateInfo
}

func ListPrivateGists() ([]*github.Gist, *github.Response) {
	var allGists []*github.Gist
	var rateInfo *github.Response

	ctx := context.Background()
	client := authenticateClientWithPAT()
	opt := &github.GistListOptions{
		ListOptions: github.ListOptions{PerPage: PER_PAGE},
	}

	for {
		gists, resp, err := client.Gists.List(ctx, "", opt)
		if err != nil {
			log.Fatal("Error listing gists: ", err)
		}
		LogResponse(resp)

		allGists, rateInfo = append(allGists, gists...), resp
		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}
	return allGists, rateInfo
}
