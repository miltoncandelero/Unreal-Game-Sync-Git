package controller

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"github.com/miltoncandelero/ugsg/core"
	"github.com/miltoncandelero/ugsg/gui/model"
	"github.com/miltoncandelero/ugsg/gui/view"
)

type MainApp struct {
	App         fyne.App
	Window      fyne.Window
	MainTabs    *container.DocTabs
	ProjectTabs map[string]*container.TabItem
}

var mainAppRef *MainApp

// Singleton of sorts
func GetApp() *MainApp {
	if mainAppRef == nil {
		myApp := app.NewWithID("com.killabunnies.ugsg")
		myWindow := myApp.NewWindow("List Data")
		myWindow.Resize(fyne.NewSize(1600, 600))

		mainAppRef = &MainApp{
			App:         myApp,
			Window:      myWindow,
			ProjectTabs: make(map[string]*container.TabItem),
		}
	}

	return mainAppRef
}

func InitializeApplication() *MainApp {

	mainApp := GetApp()
	LoadGUIConfig()

	if !isRunningFromInsideProject() {
		mainTabs := container.NewDocTabs()
		mainApp.Window.SetContent(mainTabs)
		openProjectScreen := view.MakeOpenProjectScreen(UProjectOpened, openFilePickerUproject, ForgetProject, GetConfig())
		openFileTab := container.NewTabItem("Open Project", openProjectScreen)
		mainTabs.CloseIntercept = func(tab *container.TabItem) {
			if tab == openFileTab {
				return
			}
			mainTabs.Remove(tab)
			if f := mainTabs.OnClosed; f != nil {
				f(tab)
			}
			for key, openedTab := range mainApp.ProjectTabs {
				if openedTab == tab {
					delete(mainApp.ProjectTabs, key)
				}
			}
		}
		mainTabs.Append(openFileTab)
		mainApp.MainTabs = mainTabs
	} else {
		println("Running from inside project")
		println("TODO: implement this")
	}
	return mainApp
}

func appendProjectToMainWindow(content fyne.CanvasObject, projectPath string) *container.TabItem {
	mainApp := GetApp()
	if !isRunningFromInsideProject() {
		if mainApp.ProjectTabs[projectPath] != nil {
			mainApp.MainTabs.Select(mainApp.ProjectTabs[projectPath])
			return mainApp.ProjectTabs[projectPath]
		} else {
			newTab := container.NewTabItem(filepath.Base(projectPath), content)
			mainApp.ProjectTabs[projectPath] = newTab
			mainApp.MainTabs.Append(newTab)
			mainApp.MainTabs.Select(newTab)
			return newTab
		}
	} else {
		mainApp.Window.SetContent(content)
		return nil
	}
}

func isRunningFromInsideProject() bool {
	entries, err := os.ReadDir("./")
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.Name() == core.CONFIG_FILE {
			return true
		}
	}
	return false
}

var config *model.GUIConfig

func LoadGUIConfig() *model.GUIConfig {

	config = &model.GUIConfig{
		RecentProjects: GetApp().App.Preferences().StringListWithFallback("recentProjects", []string{}),
	}

	return config
}

func GetConfig() *model.GUIConfig {
	return config
}

func SaveConfig() {
	GetApp().App.Preferences().SetStringList("recentProjects", config.RecentProjects)
}
