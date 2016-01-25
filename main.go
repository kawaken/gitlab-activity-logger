package main

import (
	"log"
	"os"

	"github.com/xanzy/go-gitlab"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	token := os.Getenv("GITLAB_TOKEN")
	log.Printf("token: %s", token)

	var (
		author  = kingpin.Flag("user", "user").Envar("USER").String()
		baseURL = kingpin.Flag("url", "url").Envar("URL").String()
	)

	kingpin.Parse()

	if *author == "" {
		log.Fatal("no author defined")
		return
	}

	git := gitlab.NewClient(nil, token)

	if *baseURL != "" {
		git.SetBaseURL(*baseURL)
	}

	ps, res, err := git.Projects.ListOwnedProjects(nil)
	if err != nil {
		log.Fatal(err)
	}

	if res.StatusCode != 200 {
		log.Fatalf("responce %d:%s", res.StatusCode, res.Status)
	}

	for _, p := range ps {
		log.Print(p)
	}
}
