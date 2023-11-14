package main

import (
	"log"
	"os"

	"fyne.io/fyne/v2/container"
	"github.com/miltoncandelero/ugsg/core"
	"github.com/miltoncandelero/ugsg/gui"
)

func main() {

	mainApp := gui.GetApp()

	if !isRunningFromInsideProject() {
		gui.LoadGUIConfig(gui.GUI_CONFIG_FILE)
		mainTabs := container.NewAppTabs()
		mainApp.Window.SetContent(mainTabs)
		openFileTab := container.NewTabItem("Open Project", gui.MakeOpenProjectScreen())
		mainTabs.Append(openFileTab)
	} else {
		println("Running from inside project")
		println("TODO: implement this")
		return
	}

	mainApp.Window.ShowAndRun()
}

func isRunningFromInsideProject() bool {
	entries, err := os.ReadDir("./")
	if err != nil {
		log.Println(err)
		return false
	}
	for _, entry := range entries {
		if entry.Name() == core.CONFIG_FILE {
			return true
		}
	}
	return false
}
