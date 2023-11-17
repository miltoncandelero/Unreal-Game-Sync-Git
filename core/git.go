package core

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type LockDatum struct {
	ID    string `json:"id"`
	Path  string `json:"path"`
	Owner struct {
		Name string `json:"name"`
	} `json:"owner"`
	LockedAt time.Time `json:"locked_at"`
}

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

func LinkGitConfig(repoPath string) error {
	_, err := Execute(repoPath, GIT, "config", "--local", "include.path", ".gitconfig")
	return err
}

type GitConfigStatus int

const (
	FILE_MISSING GitConfigStatus = iota
	FILE_EXIST_BUT_NOT_LINKED
	FILE_LINKED
)

func GetGitConfigStatus(repoPath string) GitConfigStatus {
	_, err := os.Stat(filepath.Join(repoPath, ".gitconfig"))
	if err != nil {
		return FILE_MISSING
	}
	_, err = Execute(repoPath, GIT, "config", "--local", "include.path")
	if err != nil {
		return FILE_EXIST_BUT_NOT_LINKED
	}

	return FILE_LINKED
}

func NeedsCredentials(repoPath string) bool {
	username, _ := ExecuteOneLine(repoPath, GIT, "config", "user.name")
	if username == "" {
		return true
	}
	email, _ := ExecuteOneLine(repoPath, GIT, "config", "user.email")
	if email == "" {
		return true
	}
	return false
}

func GetUsernameFromRepo(repoPath string) string {
	username, _ := ExecuteOneLine(repoPath, GIT, "config", "user.name")
	return strings.TrimSpace(username)
}

func GetUserEmailFromRepo(repoPath string) string {
	email, _ := ExecuteOneLine(repoPath, GIT, "config", "user.email")
	return strings.TrimSpace(email)
}

func GetGitProviderName(repoPath string) string {
	// Probably fails if you have many remotes

	remotes, _ := ExecuteOneLine(repoPath, GIT, "remote", "get-url", "origin")

	if strings.Contains(remotes, "github") {
		return "GitHub"
	}
	if strings.Contains(remotes, "gitlab") {
		return "GitLab"
	}
	if strings.Contains(remotes, "gitea") {
		return "Gitea" // probably not correct, but what can I do?
	}

	return "Unknown"
}

func GetRepoOrigin(repoPath string) string {
	remotes, _ := ExecuteOneLine(repoPath, GIT, "remote", "get-url", "origin")
	return strings.TrimSpace(remotes)
}

func FinishRebase(repoPath string) error {
	// Non porcelain command but there is no way to know if we are mid rebase :(
	currentStatus, _ := ExecuteOneLine(repoPath, GIT, "status", "-uno")
	if !strings.Contains(currentStatus, "rebase") {
		return errors.New("Not in a rebase")
	}

	if IsUnrealRunning() {
		return errors.New("Unreal is running, cannot finish rebase")
	}

	if strings.Contains(currentStatus, "git rebase --continue") ||
		strings.Contains(currentStatus, "nothing to commit") ||
		strings.Contains(currentStatus, "all conflicts fixed") {

		// We should be able to continue the rebase
		_, err := ExecuteOneLine(repoPath, GIT, "rebase", "--continue")
		return err
	} else {
		// We are in a rebase but we have conflicts, this is baaaad
		return errors.New("You are in the middle of a rebase. Changes on one of your commits will be overridden by incoming changes. Please request help to resolve conflicts.")
	}
}

func PruneLFS(repoPath string) error {
	_, err := ExecuteOneLine(repoPath, GIT, "lfs", "prune", "-fc")
	return err
}

func IsShallowRepo(repoPath string) bool {
	isShallow, _ := ExecuteOneLine(repoPath, GIT, "rev-parse", "--is-shallow-repository")
	if strings.Contains(isShallow, "true") {
		return true
	}
	return false
}

func UnshallowRepo(repoPath string) error {
	if IsShallowRepo(repoPath) {
		_, err := ExecuteOneLine(repoPath, GIT, "fetch", "--unshallow")
		return err
	}
	return nil
}

func IsWorkingTreeClean(repoPath string) bool {
	status, _ := ExecuteOneLine(repoPath, GIT, "status", "--porcelain")
	return status == ""
}

func GetAheadBehind(repoPath string) (int, int, error) {
	statusLines, _ := Execute(repoPath, GIT, "status", "--porcelain=2", "--branch")
	for _, line := range statusLines {
		if strings.Contains(line, "# branch.ab ") {
			strings.Replace(line, "# branch.ab ", "", 1)
			ab := strings.Split(line, " ")
			ahead, _ := strconv.Atoi(ab[0])
			behind, _ := strconv.Atoi(ab[1])
			return ahead, behind, nil
		}
	}
	return 0, 0, errors.New("Could not find ahead/behind")
}

func GitSmartPull(repoPath string) error {
	ahead, behind, err := GetAheadBehind(repoPath)
	if err != nil {
		return err
	}
	if behind == 0 {
		// nothing to pull :)
		return nil
	}
	if ahead == 0 {
		// we can fast forward, we have no changes!
		_, err := ExecuteOneLine(repoPath, GIT, "pull", "--ff-only")
		return err
	} else {
		// we have changes, we need to rebase
		_, err := ExecuteOneLine(repoPath, GIT, "pull", "--rebase", "--autostash")
		return err
	}
}

func GitPush(repoPath string) error {
	ahead, behind, err := GetAheadBehind(repoPath)

	if err != nil {
		return err
	}
	if ahead == 0 {
		//nothing to push :)
		return nil
	}

	if behind != 0 {
		// we are behind, we need to pull first
		return errors.New("Cannot push, you are behind")
	}

	_, err = ExecuteOneLine(repoPath, GIT, "push")
	return err
}

func GetLockedFiles(repoPath string, fromUser string) ([]LockDatum, error) {
	jsonLocks, _ := ExecuteOneLine(repoPath, GIT, "lfs", "locks", "--json")
	locks := make([]LockDatum, 0)
	err := json.Unmarshal([]byte(jsonLocks), &locks)
	if err != nil {
		return nil, err
	}
	if fromUser == "" {
		return locks, nil
	}
	filteredLocks := make([]LockDatum, 0)
	for _, lock := range locks {
		if lock.Owner.Name == fromUser {
			filteredLocks = append(filteredLocks, lock)
		}
	}
	return filteredLocks, nil
}

func IsPathRepo(repoPath string) bool {
	_, err := ExecuteOneLine(repoPath, GIT, "rev-parse", "--git-dir")
	return err == nil
}

func IsDeatachedHead(repoPath string) bool {
	//git symbolic-ref -q HEAD
	_, err := ExecuteOneLine(repoPath, GIT, "symbolic-ref", "-q", "HEAD")
	return err != nil
}
