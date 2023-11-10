package core

import (
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type CommitDatum struct {
	Hash             string
	Msg              string
	User             string
	Emoji            map[string]bool
	DistanceFromHead int // Not fully working, now just says 1,0,-1
}

const GIT = "git"
const SEP = "§"

var SOURCE_REGEX = regexp.MustCompile(`.*(\\|\/)?Source((\\|\/).*)?`)
var CONTENT_REGEX = regexp.MustCompile(`.*(\\|\/)?Content((\\|\/).*)?`)

const SOURCE_EMOJI = "⚙️"
const CONTENT_EMOJI = "📦"

func GetRepoBranchInfo(repoPath string, branchName string) ([]*CommitDatum, error) {
	start := time.Now()

	headLines, err := Execute(repoPath, GIT, "rev-parse", "HEAD")
	if err != nil {
		return nil, err
	}
	headHash := headLines[0]

	commitInfo := []string{"%H", "%s", "%an"}

	//	cmd := exec.Command(`git.exe`, `-P`, `log`, `-10000`, `--pretty=%H||%s||%an`)
	logArgs := []string{"-P", "log", "--pretty=" + strings.Join(commitInfo, SEP)}

	if branchName != "" {
		logArgs = append(logArgs, branchName)
	} else {

	}

	logLines, err := Execute(repoPath, GIT, logArgs...)

	if err != nil {
		return nil, err
	}

	commits := make([]*CommitDatum, 0)

	beforeHead := true

	for _, commit := range logLines {
		splitCommit := strings.Split(commit, SEP)

		if len(splitCommit) != len(commitInfo) {
			continue
		}

		newCommit := &CommitDatum{
			Hash:  splitCommit[0],
			Msg:   splitCommit[1],
			User:  splitCommit[2],
			Emoji: make(map[string]bool),
		}

		if headHash == newCommit.Hash {
			beforeHead = false
			newCommit.DistanceFromHead = 0
		} else {
			if beforeHead {
				newCommit.DistanceFromHead = 1
			} else {
				newCommit.DistanceFromHead = -1
			}
		}

		// diffCmd := exec.Command("git.exe", "diff-tree", "--no-commit-id", "--name-only", "-r", commit[0], "--")
		diffLines, err := Execute(repoPath, GIT, "diff-tree", "--no-commit-id", "--name-only", "-r", newCommit.Hash, "--")
		if err != nil {
			return nil, err
		}

		for _, file := range diffLines {
			if CONTENT_REGEX.MatchString(file) {
				newCommit.Emoji[CONTENT_EMOJI] = true
			}
			if SOURCE_REGEX.MatchString(file) {
				newCommit.Emoji[SOURCE_EMOJI] = true
			}
			// faster early exit
			if len(newCommit.Emoji) == 2 {
				break
			}
		}

		commits = append(commits, newCommit)
	}

	elapsed := time.Since(start)
	log.Printf("Git took %s", elapsed)

	return commits, nil

}

func GetCurrentBranchFromRepository(repository *git.Repository) (string, error) {
	branchRefs, err := repository.Branches()
	if err != nil {
		return "", err
	}

	headRef, err := repository.Head()
	if err != nil {
		return "", err
	}

	var currentBranchName string
	err = branchRefs.ForEach(func(branchRef *plumbing.Reference) error {
		if branchRef.Hash() == headRef.Hash() {
			currentBranchName = branchRef.Name().String()

			return nil
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return currentBranchName, nil
}
