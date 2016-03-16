package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/postgres-ci/hooks/git"
)

func main() {

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {

		fields := strings.Fields(scanner.Text())

		if len(fields) != 3 {

			continue
		}

		var (
			pwd  = os.Getenv("PWD")
			push = git.Push{
				Ref: fields[2],
				Old: fields[0],
				New: fields[1],
			}
			revisions = []string{push.New}
		)

		if push.Old != git.Z40 {

			if revList, err := git.RevList(pwd, push.Old, push.New); err == nil {

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

var client http.Client

func send(push git.Push) error {

	data, _ := json.Marshal(push)

	_url := url.URL{
		Scheme: os.Getenv("SCHEME"),
		Host:   os.Getenv("HOST"),
		Path:   "/web-hooks/native/",
	}

	req, _ := http.NewRequest("POST", _url.String(), bytes.NewReader(data))

	req.Header.Set("X-Event", "push")
	req.Header.Set("X-Token", os.Getenv("TOKEN"))
	req.Header.Set("Content-Type", "application/json")

	if _, err := client.Do(req); err != nil {

		fmt.Printf("Error when sending a commit to Postgres-CI (err: %v)\n", err)

		return nil
	}

	return nil
}
