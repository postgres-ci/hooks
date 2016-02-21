package git

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const Z40 = "0000000000000000000000000000000000000000"

func RevList(dir, old, new string) ([]string, error) {

	cmd := exec.Command("git", "rev-list", fmt.Sprintf("%s..%s", old, new))
	cmd.Dir = dir
	cmd.Stdout = &bytes.Buffer{}

	if err := cmd.Run(); err != nil {

		return nil, err
	}

	var revisions []string

	for {

		if revision, err := cmd.Stdout.(*bytes.Buffer).ReadString('\n'); err == nil {

			revisions = append(revisions, strings.TrimSpace(revision))

		} else {

			if err == io.EOF {

				break
			}

			return nil, err
		}
	}

	return revisions, nil
}

func GetCommit(dir, revision string) (Commit, error) {

	cmd := exec.Command("git", "cat-file", "commit", revision)
	cmd.Dir = dir
	cmd.Stdout = &bytes.Buffer{}
	cmd.Stderr = &bytes.Buffer{}

	if err := cmd.Run(); err != nil {

		return Commit{}, fmt.Errorf(cmd.Stderr.(*bytes.Buffer).String())
	}

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

		default:

			switch fields[0] {
			case "author", "tagger":

				commit.Author = committer(fields[1:])

				if timestamp, err := strconv.ParseInt(fields[len(fields)-2], 10, 64); err == nil {

					commit.CommittedAt = time.Unix(timestamp, 0)
				}

			case "committer":
				commit.Committer = committer(fields[1:])
			}
		}
	}

	return commit, nil
}

func GetLastCommit(dir string) (Commit, error) {

	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	cmd.Stdout = &bytes.Buffer{}
	cmd.Stderr = &bytes.Buffer{}

	if err := cmd.Run(); err != nil {

		return Commit{}, fmt.Errorf(cmd.Stderr.(*bytes.Buffer).String())
	}

	revision := strings.TrimSpace(cmd.Stdout.(*bytes.Buffer).String())

	return GetCommit(dir, revision)
}

func GetCurrentRef(dir string) (string, error) {

	cmd := exec.Command("git", "symbolic-ref", "HEAD")
	cmd.Dir = dir
	cmd.Stdout = &bytes.Buffer{}

	if err := cmd.Run(); err != nil {

		return "", err
	}

	return strings.TrimSpace(cmd.Stdout.(*bytes.Buffer).String()), nil
}

func committer(in []string) Committer {

	return Committer{
		Name: strings.Join(in[:len(in)-3], " "),
		Email: strings.TrimFunc(in[len(in)-3], func(r rune) bool {
			return r == '<' || r == '>'
		}),
	}
}
