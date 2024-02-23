package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v59/github"
	"github.com/joho/godotenv"
)

func LogResponse(resp *github.Response) {
	if !resp.TokenExpiration.IsZero() {
		fmt.Println("Token Expiration:", resp.TokenExpiration)
	}

	fmt.Println("Rate: ", resp.Rate)
}

func AuthenticateClientWithPAT() *github.Client {
	err := godotenv.Load()

	if err != nil {
		fmt.Println("error:", err)
	}

	token := os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN")
	client := github.NewClient(nil).WithAuthToken(token)

	return client
}

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

func main() {
	repos, resp := ListPublicRepos("flarexes")

	for i, repo := range repos {
		fmt.Println(i, *repo.HTMLURL)
	}

	LogResponse(resp)
}
