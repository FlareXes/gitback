package vcs

import (
	"log"
	"os/exec"
	"sync"
	"time"

	"github.com/google/go-github/v59/github"
)

var maxConcurrentConnections int

var dateTime string = time.Now().Format("2006-01-02_15-04-05")
var BACKUP_REPOS_DIR string = "gitback-backup_" + dateTime + "/repos/"
var BACKUP_GISTS_DIR string = "gitback-backup_" + dateTime + "/gists/"

func gitCloneExec(url string, logURL string, outputDir string, wg *sync.WaitGroup, limiter chan int) {
	defer wg.Done()
	defer func() { <-limiter }()
	limiter <- 1

	log.Println("Cloning", logURL)
	cmd := exec.Command("git", "clone", url, outputDir)
	output, err := cmd.CombinedOutput() // TODO: Print output

	if err != nil {
		log.Println("Error cloning repository: ", url, err.Error())
		log.Println("Command output: ", string(output))
	}
}

func cloneRepos(repos []*github.Repository, noauth bool) {
	var wg sync.WaitGroup
	limiter := make(chan int, maxConcurrentConnections)

	for _, repo := range repos {
		outputRepoDir := BACKUP_REPOS_DIR + *repo.Name
		outputWikiDir := outputRepoDir + ".wiki"
		repoUrl := *repo.CloneURL
		wikiUrl := "https://github.com/" + *repo.FullName + ".wiki.git"

		if !noauth {
			repoUrl, wikiUrl = GetPatUrl(*repo.FullName)
		}

		if *repo.HasWiki {
			wg.Add(1)
			go gitCloneExec(wikiUrl, wikiUrl, outputWikiDir, &wg, limiter)
		}

		wg.Add(1)
		go gitCloneExec(repoUrl, repoUrl, outputRepoDir, &wg, limiter)
	}
	wg.Wait()
}

func cloneGists(gists []*github.Gist) {
	var wg sync.WaitGroup
	limiter := make(chan int, maxConcurrentConnections)

	for _, gist := range gists {
		outputDir := BACKUP_GISTS_DIR + *gist.ID
		url := *gist.GitPullURL
		wg.Add(1)
		go gitCloneExec(url, url, outputDir, &wg, limiter)
	}
	wg.Wait()
}

func Run(noauth bool, username string, threads int) {
	var repos []*github.Repository
	var gists []*github.Gist
	var rateInfo *github.Response
	maxConcurrentConnections = threads

	if noauth {
		repos, _ = ListPublicRepos(username)
		gists, rateInfo = ListPublicGists(username)
	} else {
		repos, _ = ListPrivateRepos()
		gists, rateInfo = ListPrivateGists()
	}

	if len(repos) != 0 {
		cloneRepos(repos, noauth)
	}

	if len(gists) != 0 {
		cloneGists(gists)
	}
	LogResponse(rateInfo)
}
