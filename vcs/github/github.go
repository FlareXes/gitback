package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/v59/github"
)

type Repo struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
	CloneURL string `json:"clone_url"`
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

func ListPrivateRepos() []Repo {
	url := "https://api.github.com/user/repos"
	pat := GetGitHubPAT()

	var mergedRepos []Repo

	for {
		// Create a new HTTP request and set required headers
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("Authorization", "Bearer "+pat)
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

		// Send the HTTP request and receive the response
		client := &http.Client{}
		resp, err := client.Do(req)

		if err != nil {
			fmt.Println("Error making request:", err)
		}

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
		}

		var repo []Repo
		json.Unmarshal([]byte(body), &repo)
		mergedRepos = append(mergedRepos, repo...)

		nextPage := getRelNextURL(resp.Header.Get("Link"))
		if nextPage == "" {
			break
		} else {
			url = nextPage
		}
	}

	return mergedRepos
}

func getRelNextURL(linkHeader string) string {
	links := strings.Split(linkHeader, ",")

	for _, link := range links {
		parts := strings.Split(strings.TrimSpace(link), ";")

		if strings.TrimSpace(parts[1]) == `rel="next"` {
			url := strings.Trim(parts[0], "<>")
			return url
		}
	}

	return ""
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
