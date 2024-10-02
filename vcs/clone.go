package vcs

import (
	"fmt"
	"log"

	"os"
	"os/exec"
	"sync"

	"github.com/google/go-github/v59/github"
)

var maxConcurrentConnections int
const backupRepoDir = "github-repositories-backup/repo/"

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
			url = GetPatUrl(*repo.FullName)
		}

		wg.Add(1)
		go cloneRepo(url, *repo.CloneURL, outputDir, &wg, limiter)
	}

	wg.Wait()
}

func Run(noauth bool, username string, threads int) {
	var repos []*github.Repository
	var rateInfo *github.Response

	if noauth {
		repos, rateInfo = ListPublicRepos(username)
	} else {
		repos, rateInfo = ListPrivateRepos()
	}

	maxConcurrentConnections = threads
	os.Mkdir(backupRepoDir, os.ModePerm)

	LogResponse(rateInfo)

	cloneRepositories(repos, noauth)

	LogResponse(rateInfo)
}
