package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/xanzy/go-gitlab"
	"gopkg.in/alecthomas/kingpin.v2"
)

func scanListAll(getList func(page int) (int, error)) error {

	loop := true
	for page := 1; loop; page++ {
		lastPage, err := getList(page)
		if err != nil {
			return err
		}
		loop = (page < lastPage)
	}

	return nil
}

func listAllOwnedProjects(git *gitlab.Client) ([]*gitlab.Project, error) {

	result := make([]*gitlab.Project, 0, 1000)

	getProjects := func(page int) (int, error) {

		lpOpt := &gitlab.ListProjectsOptions{}
		lpOpt.Page = page
		lpOpt.PerPage = 100

		ps, res, err := git.Projects.ListOwnedProjects(lpOpt)
		if err != nil {
			return 0, err
		}

		if res.StatusCode != 200 {
			return 0, fmt.Errorf("responce %d:%s", res.StatusCode, res.Status)
		}

		result = append(result, ps...)

		// 負荷軽減のため
		time.Sleep(100 * time.Millisecond)

		return res.LastPage, nil
	}

	err := scanListAll(getProjects)

	return result, err
}

func listCommitFromProject(git *gitlab.Client, proj *gitlab.Project, opt *gitlab.ListCommitsOptions) ([]*gitlab.Commit, error) {

	result := make([]*gitlab.Commit, 0, 365*10)

	pid := proj.ID
	getCommits := func(page int) (int, error) {
		opt := &gitlab.ListCommitsOptions{}
		opt.Page = page
		opt.PerPage = 100

		cs, res, err := git.Commits.ListCommits(pid, opt)
		if err != nil {
			return 0, nil
		}

		if res.StatusCode != 200 {
			return 0, fmt.Errorf("pid: %d, responce %d:%s\n", pid, res.StatusCode, res.Status)
		}

		result = append(result, cs...)

		// 負荷軽減のため
		time.Sleep(100 * time.Millisecond)

		return res.LastPage, nil
	}

	err := scanListAll(getCommits)

	return result, err

}

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

	ps, err := listAllOwnedProjects(git)
	if err != nil {
		log.Fatal(err)
	}

	cs := make([]*gitlab.Commit, 0, 365*10)
	for _, p := range ps {
		l, err := listCommitFromProject(git, p, nil)
		if err != nil {
			log.Printf("")
			continue
		}
		cs = append(cs, l...)
	}
}
