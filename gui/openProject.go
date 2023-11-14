package gui

import (
	"fmt"
	"path/filepath"
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/miltoncandelero/ugsg/core"
	"github.com/ncruces/zenity"
)

func MakeOpenProjectScreen() fyne.CanvasObject {
	config := GetConfig()
	cardList := MakeCardListsFromProjects(&config.RecentProjects, UProjectOpened)

	subtitle := canvas.NewText("Recent Projects", theme.ForegroundColor())
	subtitle.TextStyle.Bold = true
	subtitle.TextSize = theme.TextSubHeadingSize()

	// add an open button
	button := widget.NewButton("Open project", openFilePickerUproject)

	border := container.NewBorder(subtitle, button, layout.NewSpacer(), layout.NewSpacer(), cardList)

	return border
}

func UProjectOpened(uprojectPath string) {
	fmt.Println("Opening", uprojectPath)
	repoPath := filepath.Dir(uprojectPath)
	if !core.IsPathRepo(repoPath) {
		// This is not a repo! panic
		dialog.ShowError(fmt.Errorf("This is not a git repository"), GetApp().Window)
		return
	}

	config := GetConfig()

	foundIdx := slices.Index(config.RecentProjects, uprojectPath)
	if foundIdx != -1 {
		// Remove it from the list
		config.RecentProjects = slices.Delete(config.RecentProjects, foundIdx, foundIdx+1)
	}

	config.RecentProjects = append([]string{uprojectPath}, config.RecentProjects...)
	SaveConfig()
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
		UProjectOpened(file)
	}
}
