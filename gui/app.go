package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

type MainApp struct {
	App    fyne.App
	Window fyne.Window
}

var mainAppRef *MainApp

// Singleton of sorts
func GetApp() *MainApp {
	if mainAppRef == nil {
		myApp := app.New()
		myWindow := myApp.NewWindow("List Data")
		myWindow.Resize(fyne.NewSize(1600, 600))

		mainAppRef = &MainApp{
			App:    myApp,
			Window: myWindow,
		}
	}

	return mainAppRef
}
