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

func GetGitHubPAT() string {
	err := godotenv.Load()

	if err != nil {
		fmt.Println("error:", err)
	}

	return os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN")
}

func AuthenticateClientWithPAT() *github.Client {
	pat := GetGitHubPAT()
	client := github.NewClient(nil).WithAuthToken(pat)

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

func CloneRepo() {

}

func GetPatUrl(fullname string) string {
	pat := GetGitHubPAT()
	url := fmt.Sprintf("git clone https://%s@github.com/%s.git", pat, fullname)

	return url
}

func main() {
	repos, resp := ListPublicRepos("flarexes")

	for i, repo := range repos {
		fmt.Println(i, GetPatUrl(*repo.FullName))
	}

	LogResponse(resp)
}
