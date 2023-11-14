package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/miltoncandelero/ugsg/core"
	"github.com/miltoncandelero/ugsg/gui"
	"github.com/ncruces/zenity"
)

var myApp fyne.App
var myWindow fyne.Window

func main() {

	myApp = app.New()
	myWindow = myApp.NewWindow("List Data")
	myWindow.Resize(fyne.NewSize(1000, 600))

	myWindow.SetOnDropped(func(_ fyne.Position, uris []fyne.URI) {
		for _, uri := range uris {
			if strings.Contains(uri.String(), ".uproject") {
				uprojectOpened(uri.Path())
				return
			}
		}
	})

	//vertical box for menu

	config := gui.LoadGUIConfig(gui.GUI_CONFIG_FILE)
	cardList := gui.MakeCardListsFromProjects(&config.RecentProjects)

	// add an open button
	button := widget.NewButton("Open project", openFilePickerUproject)

	center := container.NewBorder(layout.NewSpacer(), button, layout.NewSpacer(), layout.NewSpacer(), cardList)

	myWindow.SetContent(container.NewAppTabs(container.NewTabItem("Open Project", center)))
	myWindow.ShowAndRun()
}

func uprojectOpened(uprojectPath string) {
	fmt.Println("Opening", uprojectPath)
	repoPath := filepath.Dir(uprojectPath)
	if !core.IsPathRepo(repoPath) {
		// This is not a repo! panic
		dialog.ShowError(fmt.Errorf("This is not a git repository"), myWindow)
		return
	}

	config := gui.GetConfig()
	if !slices.Contains(config.RecentProjects, uprojectPath) {
		config.RecentProjects = append(gui.GetConfig().RecentProjects, uprojectPath)
		gui.SaveConfig()
	}

	// cards.Add(gui.MakeCardFromProject(uprojectPath))
	// cards.Refresh()

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

func openFilePickerUproject() {
	file, err := zenity.SelectFile(
		zenity.Filename("./"),
		zenity.Title("Select Unreal Engine Project file"),
		zenity.Modal(),
		zenity.FileFilters{
			{
				Name:     "Unreal Engine Project files",
				Patterns: []string{"*.uproject"},
				CaseFold: false,
			},
		},
	)
	if err == nil {
		uprojectOpened(file)
	}
}
