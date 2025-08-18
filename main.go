/*
Gitback provides functionality to backup GitHub repositories either
using a GitHub Personal Access Token (PAT) or without authentication (public only).

Author: FlareXes
License:  BSD-3-Clause
Project: https://github.com/flarexes/gitback/
*/

package main

import (
	"flag"
	"log"
	"os"

	"github.com/flarexes/gitback/vcs"
)

func main() {
	noauth := flag.Bool("noauth", false, "Disable GitHub Auth, Limited to 60 Request/Hr & Public Data")
	username := flag.String("username", "", "Required When --noauth is Set")
	thread := flag.Int("thread", 10, "Maximum number of concurrent connections")
	token := flag.String("token", "", "GitHub API Token")

	flag.Parse()

	if *noauth && *username == "" {
		log.Println("Error: when --noauth is set, --username is required")
		log.Fatal("Usage: gitback --noauth --username flarexes")
	}

	if *thread <= 1 {
		log.Fatal("Error: --thread must be a positive integer")
	}

	if *token != "" {
		if err := os.Setenv("GITHUB_PERSONAL_ACCESS_TOKEN", *token); err != nil {
			log.Fatal("Error setting environment variable:", err)
		}
	}

	vcs.Run(*noauth, *username, *thread)
}
