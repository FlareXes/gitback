/*
GitBack provides functionality to backup GitHub repositories either
using a GitHub Personal Access Token (PAT) or without authentication (public only).

Author: FlareXes
License:  BSD-3-Clause
Original project link: https://github.com/flarexes/gitback/
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/google/go-github/v59/github"
	"github.com/joho/godotenv"
)

const backupDir = "github-repositories-backup/"

func LogResponse(resp *github.Response) {
	if !resp.TokenExpiration.IsZero() {
		fmt.Println("Token Expiration:", resp.TokenExpiration)
	}

	fmt.Println("Rate:", resp.Rate)
}

func GetGitHubPAT() string {
	err := godotenv.Load()

	if err != nil {
		log.Println("error:", err)
	}

	pat := os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN")

	if pat == "" {
		log.Fatal("GITHUB_PERSONAL_ACCESS_TOKEN environment variable does not exist")
	}

	return pat
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

func CloneRepositories(repos []*github.Repository, noauth bool) {
	for _, repo := range repos {
		outputDir := backupDir + *repo.Name
		url := *repo.CloneURL

		if !noauth {
			url = GetPatUrl(*repo.FullName)
		}

		log.Println("Cloning", *repo.CloneURL)
		CloneRepo(url, outputDir)
	}
}

func GetPatUrl(fullname string) string {
	pat := GetGitHubPAT()
	url := fmt.Sprintf("git clone https://%s@github.com/%s.git", pat, fullname)

	return url
}

func main() {
	os.Mkdir(backupDir, os.ModePerm)
	noauth := flag.Bool("noauth", false, "Disable GitHub Auth, Limited to 60 Request/Hr & Public Data")
	username := flag.String("username", "", "Required When --noauth is Set")
	flag.Parse()

	if *noauth && *username == "" {
		log.Fatal("error: when --noauth is set, --username is required.")
	}

	var repos []*github.Repository
	var resp *github.Response

	if *noauth {
		repos, resp = ListPublicRepos(*username)
	} else {
		repos, resp = ListPrivateRepos()
	}

	CloneRepositories(repos, *noauth)
	LogResponse(resp)
}
