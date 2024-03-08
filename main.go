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

	"github.com/flarexes/gitback/vcs"
)

func main() {
	noauth := flag.Bool("noauth", false, "Disable GitHub Auth, Limited to 60 Request/Hr & Public Data")
	username := flag.String("username", "", "Required When --noauth is Set")
	threads := flag.Int("threads", 10, "Maximum number of concurrent connections")

	flag.Parse()

	if *noauth && *username == "" {
		log.Fatal("error: when --noauth is set, --username is required.")
	}

	vcs.Run(*noauth, *username, *threads)
}
