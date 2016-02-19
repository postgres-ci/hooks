package git

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

const Z40 = "0000000000000000000000000000000000000000"

func GitRevList(dir, old, new string) ([]string, error) {

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

	if err := cmd.Run(); err != nil {

		return Commit{}, err
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
			case "committer":
				commit.Committer = committer(fields[1:])
			}
		}
	}

	return commit, nil
}

func committer(in []string) Committer {

	return Committer{
		Name: strings.Join(in[:len(in)-3], " "),
		Email: strings.TrimFunc(in[len(in)-3], func(r rune) bool {
			return r == '<' || r == '>'
		}),
	}
}
