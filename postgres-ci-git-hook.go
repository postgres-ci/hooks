package main

import (
	"github.com/postgres-ci/hooks/git"

	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const usage = `
Postgres-CI Git hook

Environment variables:

    HOST  - Postgres-CI app-server host
    TOKEN - Repository token
    EVENT - One of two possible values: post-commit or post-receive
    
Example:

    Server-side hook:

        Add to post-receive hook file (.git/hooks/post-receive)

            #!/bin/bash

            export HOST=https://postgres-ci.com
            export TOKEN=587e6d7b-a023-4cab-b982-8169557e9e0c
            export EVENT=post-receive

            /usr/bin/postgres-ci-git-hook

        and make it executable (chmod +x)

    Local hook:

        Add to post-commit hook file (.git/hooks/post-commit)
        
            #!/bin/bash

            export HOST=http://postgres-ci.local
            export TOKEN=587e6d7b-a023-4cab-b982-8169557e9e0c
            export EVENT=post-commit

            /usr/bin/postgres-ci-git-hook

        and make it executable (chmod +x)
`

var (
	address *url.URL
	client  = http.Client{
		Timeout: time.Second * 2,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
)

func main() {

	if isH() {

		fmt.Print(usage)

		os.Exit(0)
	}

	var err error

	address, err = url.Parse(os.Getenv("HOST"))

	if err != nil {

		log.Fatalf("Invalid host: %v", err)
	}

	if os.Getenv("EVENT") == "post-commit" {

		postCommit()

		return
	}

	postReceive()
}

func postCommit() {

	ref, err := git.GetCurrentRef(os.Getenv("PWD"))

	if err != nil {

		return
	}

	if commit, err := git.GetLastCommit(os.Getenv("PWD")); err == nil {

		err := send("commit", struct {
			Ref string `json:"ref"`
			git.Commit
		}{
			Ref:    ref,
			Commit: commit,
		})

		if err != nil {

			fmt.Println(err)

			return
		}

		fmt.Printf("\nCommit %s was sent to Postgres-CI server (%s://%s)\n\n", commit.ID, address.Scheme, address.Host)
	}
}

func postReceive() {

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

		if err := send("push", push); err != nil {

			fmt.Println(err)

			return
		}

		fmt.Println("\nCommits:")

		for _, commit := range push.Commits {

			fmt.Printf("  %s %s %s\n", push.Ref, commit.ID, commit.Author.Name)
		}

		fmt.Printf("\nwas sent to Postgres-CI server (%s://%s)\n\n", address.Scheme, address.Host)
	}
}

func send(event string, message interface{}) error {

	buffer := bytes.Buffer{}

	if err := json.NewEncoder(&buffer).Encode(message); err != nil {

		return err
	}

	url := url.URL{
		Scheme: address.Scheme,
		Host:   address.Host,
		Path:   "/webhooks/native/",
	}

	req, _ := http.NewRequest("POST", url.String(), &buffer)

	req.Header.Set("X-Event", event)
	req.Header.Set("X-Token", os.Getenv("TOKEN"))
	req.Header.Set("Content-Type", "application/json")

	response, err := client.Do(req)

	if err != nil {

		return fmt.Errorf("Error when sending a '%s' event to Postgres-CI: %v", event, err)
	}

	if response.StatusCode != http.StatusOK {

		if response.StatusCode == http.StatusBadRequest {

			var message struct {
				Success bool   `json:"success"`
				Code    int    `json:"code"`
				Error   string `json:"error"`
			}

			if err := json.NewDecoder(response.Body).Decode(&message); err == nil {

				return fmt.Errorf("Error when sending a '%s' to Postgres-CI: %s", event, message.Error)
			}
		}

		return fmt.Errorf("Error when sending a '%s' event to Postgres-CI: %s", event, response.Status)
	}

	return nil
}

func isH() bool {

	for _, env := range []string{"HOST", "EVENT", "TOKEN"} {

		if len(os.Getenv(env)) == 0 {

			return true
		}
	}

	if event := os.Getenv("EVENT"); !(event == "post-commit" || event == "post-receive") {

		return true
	}

	for _, flag := range os.Args {

		if flag == "-h" || flag == "--help" {

			return true
		}
	}

	return false
}
