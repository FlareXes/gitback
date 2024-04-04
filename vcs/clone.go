package vcs

import (
	"fmt"
	"log"

	"os"
	"os/exec"
	"sync"

	gh "github.com/flarexes/gitback/vcs/github"
	"github.com/google/go-github/v59/github"
)

const backupRepoDir = "github-repositories-backup/repo/"
const backupGistDir = "github-repositories-backup/gist/"

var maxConcurrentConnections int

func cloneRepo(url string, logURL string, outputDir string, wg *sync.WaitGroup, limiter chan int) {
	defer wg.Done()
	defer func() { <-limiter }()
	limiter <- 1

	log.Println("Cloning", logURL)
	cmd := exec.Command("git", "clone", url, outputDir)
	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Println("Error cloning repository:", url, err.Error())
		fmt.Println("Command output:", string(output))
	}
}

func cloneRepositories(repos []*github.Repository, noauth bool) {
	var wg sync.WaitGroup
	limiter := make(chan int, maxConcurrentConnections)

	for _, repo := range repos {
		outputDir := backupRepoDir + *repo.Name
		url := *repo.CloneURL

		if !noauth {
			url = gh.GetPatUrl(*repo.FullName)
		}

		wg.Add(1)
		go cloneRepo(url, *repo.CloneURL, outputDir, &wg, limiter)
	}

	wg.Wait()
}

func cloneGists(gists []*github.Gist) {
	var wg sync.WaitGroup
	limiter := make(chan int, maxConcurrentConnections)

	for _, gist := range gists {
		outputDir := backupGistDir + *gist.ID
		url := *gist.HTMLURL

		wg.Add(1)
		go cloneRepo(url, url, outputDir, &wg, limiter)
	}

	wg.Wait()
}

func Run(noauth bool, username string, threads int) {
	var repos []*github.Repository
	var gists []*github.Gist
	var rateInfo *github.Response

	if noauth {
		repos, rateInfo = gh.ListPublicRepos(username)
		gists, rateInfo = gh.ListPublicGists(username)
	} else {
		repos, rateInfo = gh.ListPrivateRepos()
		gists, rateInfo = gh.ListPublicGists(username)
	}

	maxConcurrentConnections = threads
	os.Mkdir(backupRepoDir, os.ModePerm)
	os.Mkdir(backupGistDir, os.ModePerm)

	gh.LogResponse(rateInfo)

	cloneGists(gists)
	cloneRepositories(repos, noauth)

	gh.LogResponse(rateInfo)

}
