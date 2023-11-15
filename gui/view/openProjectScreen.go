package view

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/miltoncandelero/ugsg/gui/model"
)

func MakeOpenProjectScreen(UProjectOpened func(string), openFilePickerUproject func(), forgetProject func(string), config *model.GUIConfig) fyne.CanvasObject {
	cardList := MakeCardListsFromProjects(&config.RecentProjects, UProjectOpened, forgetProject)

	subtitle := canvas.NewText("Recent Projects", theme.ForegroundColor())
	subtitle.TextStyle.Bold = true
	subtitle.TextSize = theme.TextSubHeadingSize()

	// add an open button
	button := widget.NewButton("Open project", openFilePickerUproject)

	border := container.NewBorder(subtitle, button, layout.NewSpacer(), layout.NewSpacer(), cardList)

	return border
}
