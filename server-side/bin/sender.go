package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/postgres-ci/hooks/git"
)

type Push struct {
	Repository string       `json:"repository"`
	Ref        string       `json:"ref"`
	Before     string       `json:"before"`
	After      string       `json:"after"`
	Commits    []git.Commit `json:"commits"`
}

func main() {

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {

		fields := strings.Fields(scanner.Text())

		if len(fields) != 3 {

			continue
		}

		var (
			pwd  = os.Getenv("PWD")
			push = Push{
				Repository: os.Getenv("REPOSITORY"),
				Ref:        fields[2],
				Before:     fields[0],
				After:      fields[1],
			}
			revisions = []string{push.After}
		)

		if push.Before != git.Z40 {

			if revList, err := git.GitRevList(pwd, push.Before, push.After); err == nil {

				revisions = revList
			}
		}

		for _, revision := range revisions {

			if commit, err := git.GetCommit(pwd, revision); err == nil {

				push.Commits = append(push.Commits, commit)
			}
		}

		send(push)
	}
}

func send(push Push) error {

	data, _ := json.Marshal(push)

	_url := url.URL{
		Scheme: "http",
		Host:   os.Getenv("API_HOST"),
		Path:   "/web-hooks/native/",
	}

	_, err := http.Post(
		_url.String(),
		"application/x-www-form-urlencoded",
		bytes.NewReader(data),
	)

	if err != nil {

		return err
	}

	return nil
}
