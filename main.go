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
	"fyne.io/fyne/v2/storage"
	"github.com/miltoncandelero/ugsg/core"
	"github.com/miltoncandelero/ugsg/gui"
)

var myApp fyne.App
var myWindow fyne.Window
var cards *fyne.Container

func main() {

	myApp = app.New()
	myWindow = myApp.NewWindow("List Data")
	myWindow.Resize(fyne.NewSize(1000, 600))

	myWindow.SetOnDropped(func(_ fyne.Position, uris []fyne.URI) {
		for _, uri := range uris {
			if strings.Contains(uri.String(), ".uproject") {
				uprojectOpened(uri)
				return
			}
		}
	})
	d := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			fmt.Println(err)
			return
		}
		if reader == nil {
			fmt.Println("No file selected")
			return
		}
		defer reader.Close()
		uri := reader.URI()
		if strings.Contains(uri.String(), ".uproject") {
			uprojectOpened(uri)
			return
		}
	}, myWindow)

	d.Resize(fyne.NewSize(600, 600))
	d.SetFilter(storage.NewExtensionFileFilter([]string{".uproject"}))
	d.Show()

	//vertical box for cards
	cards = container.NewVBox()

	center := container.NewCenter(cards)

	config := gui.LoadGUIConfig(gui.GUI_CONFIG_FILE)
	fmt.Printf("config: %v\n", config)
	for _, project := range config.RecentProjects {
		fmt.Printf("project: %v\n", project)
		cards.Add(gui.MakeCardFromProject(project))
	}

	cards.Refresh()

	myWindow.SetContent(center)
	myWindow.ShowAndRun()
}

func uprojectOpened(uprojectPath fyne.URI) {
	fmt.Println("Opening", uprojectPath)
	repoPath := filepath.Dir(uprojectPath.Path())
	if !core.IsPathRepo(repoPath) {
		// This is not a repo! panic
		dialog.ShowError(fmt.Errorf("This is not a git repository"), myWindow)
		return
	}

	config := gui.GetConfig()
	if !slices.Contains(config.RecentProjects, uprojectPath.Path()) {
		config.RecentProjects = append(gui.GetConfig().RecentProjects, uprojectPath.Path())
		gui.SaveConfig()
	}

	cards.Add(gui.MakeCardFromProject(uprojectPath.Path()))
	cards.Refresh()

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
