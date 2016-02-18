package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

const zeroRevision = "0000000000000000000000000000000000000000"

type Committer struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
type Commit struct {
	ID        string    `json:"id"`
	Author    Committer `json:"author"`
	Committer Committer `json:"committer"`
	Message   string    `json:"message"`
}

type Push struct {
	Repository string   `json:"repository"`
	Ref        string   `json:"ref"`
	Before     string   `json:"before"`
	After      string   `json:"after"`
	Commits    []Commit `json:"commits"`
}

func main() {

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {

		fields := strings.Fields(scanner.Text())

		if len(fields) != 3 {

			continue
		}

		push := Push{
			Repository: os.Getenv("REPOSITORY"),
			Ref:        fields[2],
			Before:     fields[0],
			After:      fields[1],
		}

		var revisions []string

		if push.Before != zeroRevision {

			cmd := exec.Command("git", "rev-list", fmt.Sprintf("%s..%s", push.Before, push.After))
			cmd.Stdout = &bytes.Buffer{}
			cmd.Run()

			for {

				if revision, err := cmd.Stdout.(*bytes.Buffer).ReadString('\n'); err == nil {

					revisions = append(revisions, strings.TrimSpace(revision))

				} else {

					break
				}
			}

		} else {

			revisions = append(revisions, push.After)
		}

		for _, revision := range revisions {

			cmd := exec.Command("git", "cat-file", "commit", revision)
			cmd.Stdout = &bytes.Buffer{}
			cmd.Run()

			commit := Commit{
				ID: revision,
			}

			for {

				line, err := cmd.Stdout.(*bytes.Buffer).ReadString(byte('\n'))

				if err != nil {

					break
				}

				fields := strings.Fields(line)

				switch true {
				case len(fields) == 0:
					commit.Message = cmd.Stdout.(*bytes.Buffer).String()
					break

				case len(fields) >= 4:

					switch fields[0] {
					case "author", "tagger":
						commit.Author = committer(fields[1:])
					case "committer":
						commit.Committer = committer(fields[1:])
					}
				}
			}

			push.Commits = append(push.Commits, commit)
		}

		json.NewEncoder(os.Stdout).Encode(push)

		send(push)
	}
}

func send(push Push) error {

	data, _ := json.Marshal(push)

	_url := url.URL{
		Scheme: "http",
		Host:   os.Getenv("API_HOST"),
		Path:   "/web-hook/native/",
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

func committer(in []string) Committer {

	return Committer{
		Name: strings.Join(in[:len(in)-3], " "),
		Email: strings.TrimFunc(in[len(in)-3], func(r rune) bool {
			return r == '<' || r == '>'
		}),
	}
}
