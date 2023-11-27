package core

import (
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
	"github.com/go-git/go-git/v5/plumbing/object"
)

type CommitDatum struct {
	Hash          string
	Msg           string
	User          string
	Date          time.Time
	ContentChange bool
	SourceChange  bool
}

const GIT = "git"
const SEP = "§"
const ORIGIN = "origin"

var SOURCE_REGEX = regexp.MustCompile(`.*(\\|\/)?Source((\\|\/).*)?`)
var CONTENT_REGEX = regexp.MustCompile(`.*(\\|\/)?Content((\\|\/).*)?`)

func GetRepoBranchInfo(repoPath string, branchName string) ([]*CommitDatum, error) {
	start := time.Now()

	r, err := git.PlainOpen(repoPath)
	if err != nil {
		log.Fatal(err)
	}

	commitsIter, _ := r.Log(&git.LogOptions{})

	commitData := make([]*CommitDatum, 0)

	commitsIter.ForEach(func(c *object.Commit) error {

		newCommit := &CommitDatum{
			Hash: c.Hash.String(),
			Msg:  strings.Split(c.Message, "\n")[0],
			User: c.Author.Name,
			Date: c.Author.When,
		}

		// git command faster than go-git
		// show vs diff-tree? Show works with initial commit. Diff-tree does not.
		// git --no-pager show <sha> --name-only --format=""
		// git --no-pager diff-tree --no-commit-id --name-only -r <sha> --
		diffLines, _ := Execute(repoPath, GIT, "show", newCommit.Hash, "--name-only", "--format=")
		// diffLines, _ := Execute(repoPath, GIT, "diff-tree", "--no-commit-id", "--name-only", "-r", newCommit.Hash, "--")

		for _, file := range diffLines {
			if CONTENT_REGEX.MatchString(file) {
				newCommit.ContentChange = true
			}
			if SOURCE_REGEX.MatchString(file) {
				newCommit.SourceChange = true
			}

			if newCommit.ContentChange && newCommit.SourceChange {
				break
			}
		}
		commitData = append(commitData, newCommit)

		return nil
	})

	elapsed := time.Since(start)
	log.Printf("Git took %s", elapsed)

	return commitData, nil

}

func GetCurrentBranchFromRepository(repoPath string) (string, error) {

	repository, _ := git.PlainOpen(repoPath)

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
	if !FileExists(filepath.Join(repoPath, ".gitconfig")) {
		return errors.New(".gitconfig file missing!")
	}

	_, err := Execute(repoPath, GIT, "config", "--local", "include.path", ".gitconfig")
	return err
}

func CreateGitConfig(repoPath string) error {
	if FileExists(filepath.Join(repoPath, ".gitconfig")) {
		return errors.New(".gitconfig already exists!")
	}

	f, err := os.Create(filepath.Join(repoPath, ".gitconfig"))

	if err != nil {
		return err
	}

	_, err2 := f.WriteString(GetConfigString())

	if err2 != nil {
		return err2
	}

	f.Close()

	LinkGitConfig(repoPath)

	return nil
}

type GitConfigStatus int

const (
	CONFIG_STATUS_MISSING GitConfigStatus = iota
	CONFIG_STATUS_NOT_LINKED
	CONFIG_STATUS_LINKED
)

func GetGitConfigStatus(repoPath string) GitConfigStatus {
	exists := FileExists(filepath.Join(repoPath, ".gitconfig"))
	if !exists {
		return CONFIG_STATUS_MISSING
	}
	_, err := Execute(repoPath, GIT, "config", "--local", "include.path")
	if err != nil {
		return CONFIG_STATUS_NOT_LINKED
	}

	return CONFIG_STATUS_LINKED
}

func NeedsUsernameFix(repoPath string) bool {
	username, _ := ExecuteOneLine(repoPath, GIT, "config", "--local", "user.name")
	if username == "" {
		return true
	}
	email, _ := ExecuteOneLine(repoPath, GIT, "config", "--local", "user.email")
	if email == "" {
		return true
	}
	return false
}

func GetUsernameFromRepo(repoPath string) string {
	username, _ := ExecuteOneLine(repoPath, GIT, "config", "--local", "user.name")
	return strings.TrimSpace(username)
}

func GetUserEmailFromRepo(repoPath string) string {
	email, _ := ExecuteOneLine(repoPath, GIT, "config", "--local", "user.email")
	return strings.TrimSpace(email)
}

func SetUsernameAndEmail(repoPath string, username string, email string) error {
	_, err := ExecuteOneLine(repoPath, GIT, "config", "--local", "user.name", username)
	if err != nil {
		return err
	}
	_, err2 := ExecuteOneLine(repoPath, GIT, "config", "--local", "user.email", email)
	if err2 != nil {
		return err2
	}
	return nil
}

func GetGitProviderName(repoPath string) string {
	// Probably fails if you have many remotes

	remotes, _ := ExecuteOneLine(repoPath, GIT, "remote", "get-url", ORIGIN)

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
	remotes, _ := ExecuteOneLine(repoPath, GIT, "remote", "get-url", ORIGIN)
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

func GetWorkingTreeFiles(repoPath string, excludeUntracked bool) ([]string, error) {
	var params []string
	if excludeUntracked {
		params = []string{"status", "--porcelain", "--untracked-files=no"}
	} else {
		params = []string{"status", "--porcelain"}
	}
	lines, err := Execute(repoPath, GIT, params...)
	if err != nil {
		return nil, err
	}

	dedupMap := make(map[string]bool)
	retval := make([]string, 0)
	for _, line := range lines {
		if line != "" {
			filename := strings.TrimSpace(line[3:])
			if filename != "" {
				if strings.Contains(filename, "->") {
					// This is a rename, we need to add both files
					split := strings.Split(filename, "->")

					file1 := strings.TrimSpace(split[0])
					if file1 != "" {
						_, exists := dedupMap[file1]
						if !exists {
							dedupMap[file1] = true
							retval = append(retval, file1)
						}
					}

					file2 := strings.TrimSpace(split[1])
					if file2 != "" {
						_, exists := dedupMap[file2]
						if !exists {
							dedupMap[file2] = true
							retval = append(retval, file2)
						}
					}
				} else {
					_, exists := dedupMap[filename]
					if !exists {
						dedupMap[filename] = true
						retval = append(retval, filename)
					}
				}
			}
		}
	}

	return retval, nil
}

func GetWorkingTreeChangeAmount(repoPath string) int {
	lines, _ := Execute(repoPath, GIT, "status", "--porcelain")
	if len(lines) == 0 || lines[0] == "" {
		return 0
	}

	//count non empty lines, is there a better way?
	count := 0
	for _, line := range lines {
		if line != "" {
			count++
		}
	}

	return count
}

func GetAheadBehind(repoPath string) (int, int, error) {
	statusLines, _ := Execute(repoPath, GIT, "status", "--porcelain=2", "--branch")
	for _, line := range statusLines {
		if strings.Contains(line, "# branch.ab ") {
			trimmed := strings.TrimSpace(strings.Replace(strings.TrimSpace(line), "# branch.ab", "", 1))
			ab := strings.Split(trimmed, " ")
			ahead, _ := strconv.Atoi(ab[0])
			behind, _ := strconv.Atoi(ab[1])
			return ahead, -behind, nil
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

func IsPathRepo(repoPath string) bool {
	_, err := ExecuteOneLine(repoPath, GIT, "rev-parse", "--git-dir")
	return err == nil
}

func IsDeatachedHead(repoPath string) bool {
	//git symbolic-ref -q HEAD
	_, err := ExecuteOneLine(repoPath, GIT, "symbolic-ref", "-q", "HEAD")
	return err != nil
}

type GitStatus int

const (
	GIT_STATUS_OK GitStatus = iota
	GIT_STATUS_SHALLOW
	GIT_STATUS_DEATACHED_HEAD
	GIT_STATUS_LAST_COMMIT_MERGE
	GIT_STATUS_REBASE_CONTINUABLE
	GIT_STATUS_REBASE_CONFLICTS
)

func GetGitStatus(repoPath string) GitStatus {

	if IsShallowRepo(repoPath) {
		return GIT_STATUS_SHALLOW
	}

	if IsDeatachedHead(repoPath) {
		return GIT_STATUS_DEATACHED_HEAD
	}

	ahead, _, _ := GetAheadBehind(repoPath)
	if ahead > 0 {
		if IsMergeCommit(repoPath, "") {
			// This shouldn't have happened! :(
			return GIT_STATUS_LAST_COMMIT_MERGE
		}
	}

	// Non porcelain command but there is no way to know if we are mid rebase :(
	currentStatus, _ := ExecuteOneLine(repoPath, GIT, "status", "-uno")
	if strings.Contains(currentStatus, "rebase") {
		if strings.Contains(currentStatus, "git rebase --continue") ||
			strings.Contains(currentStatus, "nothing to commit") ||
			strings.Contains(currentStatus, "all conflicts fixed") {

			// We should be able to continue the rebase
			return GIT_STATUS_REBASE_CONTINUABLE
		} else {
			// We are in a rebase but we have conflicts, this is baaaad
			return GIT_STATUS_REBASE_CONFLICTS
		}
	}

	return GIT_STATUS_OK
}

func IsMergeCommit(repoPath string, hash string) bool {
	if hash == "" {
		hash = "HEAD"
	}
	lines, _ := Execute(repoPath, GIT, "cat-file", "-p", hash)
	countParents := 0
	if len(lines) < 3 {
		return false
	}
	for i := 0; i < 3; i++ {
		if strings.HasPrefix(lines[i], "parent ") {
			countParents++
		}
	}

	return countParents > 1
}

func ReturnToLastBranch(repoPath string) error {
	_, err := ExecuteOneLine(repoPath, GIT, "switch", "-")
	return err
}

func Checkout(repoPath string, hash string) error {
	_, err := ExecuteOneLine(repoPath, GIT, "checkout", hash)
	return err
}

func ResetHard(repoPath string, hash string) error {
	_, err := ExecuteOneLine(repoPath, GIT, "reset", "--hard", hash)
	return err
}
