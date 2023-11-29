package core

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var LFS_SET_REGEX = regexp.MustCompile(`^([^:]+):\slockable:\sset$`)
var LFS_PUSH_FILE_REGEX = regexp.MustCompile(`^push [a-f0-9]+\s+=>\s+(.+)$`)

type LockDatum struct {
	ID    string `json:"id"`
	Path  string `json:"path"`
	Owner struct {
		Name string `json:"name"`
	} `json:"owner"`
	LockedAt time.Time `json:"locked_at"`

	// These are not part of the JSON
	AssociatedMap string `json:"-"`
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

	for i := range filteredLocks {
		file := &filteredLocks[i]
		if strings.Contains(file.Path, "__ExternalActors__") && strings.Contains(file.Path, ".uasset") {
			mapPossibleName := strings.Replace(file.Path, "__ExternalActors__/", "", 1)
			mapPossibleName = strings.Replace(mapPossibleName, ".uasset", "", 1)
			for {
				fmt.Printf("mapPossibleName: %v\n", mapPossibleName)
				if FileExists(filepath.Join(repoPath, mapPossibleName+".umap")) {
					file.AssociatedMap = mapPossibleName + ".umap"
					break
				}
				if strings.Contains(mapPossibleName, "/") {
					mapPossibleName = mapPossibleName[:strings.LastIndex(mapPossibleName, "/")]
				} else {
					break
				}
			}
		}
	}
	return filteredLocks, nil
}

func PruneLFS(repoPath string) error {
	_, err := ExecuteOneLine(repoPath, GIT, "lfs", "prune", "-fc")
	return err
}

func GetLockableFiles(repoPath string) ([]string, error) {

	allLFSFiles, err := Execute(repoPath, GIT, "lfs", "ls-files", "-n")
	if err != nil {
		return nil, err
	}

	params := []string{"check-attr", "lockable"}
	params = append(params, allLFSFiles[:len(allLFSFiles)-1]...)
	allAttrFiles, err := Execute(repoPath, GIT, params...)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return nil, err
	}

	retval := make([]string, 0)

	for i := 0; i < len(allLFSFiles)-1; i++ {
		if LFS_SET_REGEX.MatchString(allAttrFiles[i]) {
			retval = append(retval, allLFSFiles[i])
		}
	}

	return retval, nil
}

func UnlockLFSFiles(repoPath string, files []LockDatum, force bool) []error {
	retval := make([]error, 0)
	for _, file := range files {
		args := []string{"lfs", "unlock", "-i", file.ID}
		if force {
			args = append(args, "--force")
		}
		_, err := ExecuteOneLine(repoPath, GIT, args...)
		if err != nil {
			retval = append(retval, err)
		}
	}
	if len(retval) > 0 {
		return retval
	} else {
		return nil
	}
}

func ListLFSLockedUnchangedFiles(repoPath string) ([]LockDatum, error) {
	lockedFiles, err := GetLockedFiles(repoPath, "")
	if err != nil {
		return nil, err
	}

	unchangedFiles, err := GetWorkingTreeFiles(repoPath, true)
	if err != nil {
		return nil, err
	}

	pendingPush, err := GetLFSPendingPush(repoPath)
	if err != nil {
		return nil, err
	}

	retval := make([]LockDatum, 0)

	for _, lockedFile := range lockedFiles {
		canAdd := true
		for _, unchangedFile := range unchangedFiles {
			if lockedFile.Path == unchangedFile {
				canAdd = false
			}
		}
		for _, pendingFile := range pendingPush {
			if lockedFile.Path == pendingFile {
				canAdd = false
			}
		}
		if canAdd {
			retval = append(retval, lockedFile)
		}
	}

	fmt.Printf("retval: %v\n", retval)

	return retval, nil
}

func GetLFSPendingPush(repoPath string) ([]string, error) {
	branch, _ := GetCurrentBranchFromRepository(repoPath)

	files, err := Execute(repoPath, GIT, "lfs", "push", "--dry-run", ORIGIN, branch)
	if err != nil {
		return nil, err
	}

	retval := make([]string, 0)
	for _, file := range files {
		regexResult := LFS_PUSH_FILE_REGEX.FindStringSubmatch(file)
		if regexResult != nil && len(regexResult) >= 1 {
			parsedFile := strings.TrimSpace(regexResult[1])
			if parsedFile != "" {
				retval = append(retval, parsedFile)
			}
		}
	}

	return retval, nil
}
