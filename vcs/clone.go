package vcs

import (
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/google/go-github/v59/github"
)

var MaxConcurrentConnections int

const BACKUP_REPOS_DIR = "github-repositories-backup/repos/"
const BACKUP_GISTS_DIR = "github-repositories-backup/gists/"

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
	limiter := make(chan int, MaxConcurrentConnections)

	for _, repo := range repos {
		outputRepoDir := BACKUP_REPOS_DIR + *repo.Name
		outputWikiDir := outputRepoDir + ".wiki"
		repoUrl := *repo.CloneURL
		wikiUrl := "https://github.com/" + *repo.FullName + ".wiki.git"

		if !noauth {
			repoUrl, wikiUrl = GetPatUrl(*repo.FullName)
		}

		if *repo.HasWiki && *repo.HasPages {
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
	limiter := make(chan int, MaxConcurrentConnections)

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

	if noauth {
		repos, _ = ListPublicRepos(username)
		gists, rateInfo = ListPublicGists(username)
	} else {
		repos, _ = ListPrivateRepos()
		gists, rateInfo = ListPrivateGists()
	}

	MaxConcurrentConnections = threads
	os.Mkdir(BACKUP_REPOS_DIR, os.ModePerm)

	if len(repos) != 0 {
		cloneRepos(repos, noauth)
	}

	if len(gists) != 0 {
		cloneGists(gists)
	}
	LogResponse(rateInfo)
}
