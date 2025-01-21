package vcs

import (
	"log"
	"net/http"
	"os/exec"
	"sync"
	"time"

	"github.com/google/go-github/v59/github"
)

var maxConcurrentConnections int

var dateTime string = time.Now().Format("2006-01-02_15-04-05")
var BACKUP_REPOS_DIR string = "gitback-backup_" + dateTime + "/repos/"
var BACKUP_GISTS_DIR string = "gitback-backup_" + dateTime + "/gists/"

type Repo struct {
	URL       string
	OutputDir string
	HasWiki   bool
}

func gitCloneExec(repo Repo, wg *sync.WaitGroup, limiter chan int) {
	defer wg.Done()
	defer func() { <-limiter }()
	limiter <- 1

	// check if wiki truly exist
	if repo.HasWiki {
		resp, err := http.Head(repo.URL)
		if err != nil || resp.StatusCode == http.StatusNotFound {
			return
		}
	}

	log.Println("Cloning: ", repo.URL)
	cmd := exec.Command("git", "clone", repo.URL, repo.OutputDir)
	output, err := cmd.CombinedOutput() // TODO: Print output

	if err != nil {
		log.Println("Error cloning repository: ", repo.URL, err.Error())
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
			repoParams := Repo{
				URL:       wikiUrl,
				OutputDir: outputWikiDir,
				HasWiki:   true,
			}
			wg.Add(1)
			go gitCloneExec(repoParams, &wg, limiter)
		}

		repoParams := Repo{
			URL:       repoUrl,
			OutputDir: outputRepoDir,
			HasWiki:   false,
		}
		wg.Add(1)
		go gitCloneExec(repoParams, &wg, limiter)
	}
	wg.Wait()
}

func cloneGists(gists []*github.Gist) {
	var wg sync.WaitGroup
	limiter := make(chan int, maxConcurrentConnections)

	for _, gist := range gists {
		outputDir := BACKUP_GISTS_DIR + *gist.ID
		url := *gist.GitPullURL
		gistParams := Repo{
			URL:       url,
			OutputDir: outputDir,
			HasWiki:   false,
		}
		wg.Add(1)
		go gitCloneExec(gistParams, &wg, limiter)
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
