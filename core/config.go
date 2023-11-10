package core

import (
	"encoding/json"
	"os"
)

type Config struct {
	ConfigPath  string `json:"-"`
	GitExecPath string `json:"gitPath"`
	RepoPath    string `json:"repoPath"`
}

func LoadConfig(path string) Config {
	defaultConfig := &Config{
		ConfigPath:  path,
		GitExecPath: "git",
		RepoPath:    "./",
	}

	contents, err := os.ReadFile(path)

	if err != nil {
		return *defaultConfig
	}
	json.Unmarshal(contents, defaultConfig)

	return *defaultConfig
}
