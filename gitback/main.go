package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/google/go-github/v59/github"
	"github.com/joho/godotenv"
)

func LogResponse(resp *github.Response) {
	if !resp.TokenExpiration.IsZero() {
		fmt.Println("Token Expiration:", resp.TokenExpiration)
	}

	fmt.Println("Rate:", resp.Rate)
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

func CloneRepo(url string, outputDir string) {
	cmd := exec.Command("git", "clone", url, outputDir)
	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Println("Error cloning repository:", err)
		fmt.Println(output)
	}
}

func GetPatUrl(fullname string) string {
	pat := GetGitHubPAT()
	url := fmt.Sprintf("git clone https://%s@github.com/%s.git", pat, fullname)

	return url
}

func main() {
	os.Mkdir("github-repositories-backup", os.ModePerm)
	repos, resp := ListPublicRepos("flarexes")

	for _, repo := range repos {
		log.Println("Cloning", *repo.CloneURL)
		outputDir := "github-repositories-backup/" + *repo.Name

		// CloneRepo(GetPatUrl(*repo.FullName), outputDir)
		CloneRepo(*repo.CloneURL, outputDir)
	}

	LogResponse(resp)
}
