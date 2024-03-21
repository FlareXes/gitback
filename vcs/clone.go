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

const backupDir = "github-repositories-backup/"

var maxConcurrentConnections int

func CloneRepo(url string, logURL string, outputDir string, wg *sync.WaitGroup, limiter chan int) {
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

func CloneRepositories(repos []*github.Repository, noauth bool) {
	var wg sync.WaitGroup
	limiter := make(chan int, maxConcurrentConnections)

	for _, repo := range repos {
		outputDir := backupDir + *repo.Name
		url := *repo.CloneURL

		if !noauth {
			url = gh.GetPatUrl(*repo.FullName)
		}

		wg.Add(1)
		go CloneRepo(url, *repo.CloneURL, outputDir, &wg, limiter)
	}

	wg.Wait()
}

func Run(noauth bool, username string, threads int) {
	var repos []*github.Repository
	// var resp *github.Response

	if noauth {
		repos = gh.ListPublicRepos(username)
	} else {
		repos = gh.ListPrivateRepos()
	}

	maxConcurrentConnections = threads
	os.Mkdir(backupDir, os.ModePerm)

	// gh.LogResponse(resp)

	CloneRepositories(repos, noauth)

	// gh.LogResponse(resp)
}
