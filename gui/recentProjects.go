package gui

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/skratchdot/open-golang/open"
)

func MakeCardFromProject(projectFile string) *fyne.Container {
	fmt.Print("wat?")
	projectName := strings.Replace(path.Base(projectFile), ".uproject", "", -1)

	button := widget.NewButton("", func() {
		fmt.Println("Opening project: ", projectFile)
	})

	card := widget.NewCard(projectName, projectFile, nil)

	projectPath := path.Dir(projectFile)

	openInExplorer := widget.NewHyperlink("Open in Explorer", &url.URL{})

	openInExplorer.OnTapped = func() {
		fmt.Println("Opening in explorer: ", projectPath)
		open.Start(projectPath)
	}

	horizontalStack := container.NewHBox(openInExplorer)

	card.Content = horizontalStack

	overlay := container.NewStack(button, card)

	return overlay
}
