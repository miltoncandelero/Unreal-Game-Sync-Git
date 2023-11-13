package gui

import (
	"encoding/json"
	"os"
)

const GUI_CONFIG_FILE = "ugsg_gui.json"

type GUIConfig struct {
	ConfigPath     string   `json:"-"`
	RecentProjects []string `json:"recentProjects"`
}

var config *GUIConfig

func LoadGUIConfig(configPath string) *GUIConfig {

	config = &GUIConfig{
		ConfigPath:     configPath,
		RecentProjects: []string{},
	}

	contents, err := os.ReadFile(configPath)

	if err != nil {
		return config
	}
	json.Unmarshal(contents, config)

	return config
}

func GetConfig() *GUIConfig {
	return config
}

func SaveConfig() error {
	contents, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(config.ConfigPath, contents, 0644)
}
