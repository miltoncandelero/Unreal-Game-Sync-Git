package gui

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/skratchdot/open-golang/open"
)

type ProjectCard struct {
	// extends widget
	widget.BaseWidget

	Renderer *fyne.WidgetRenderer

	Container   *fyne.Container
	Title       *canvas.Text
	Subtitle    *canvas.Text
	Button      *widget.Button
	ExploreLink *widget.Hyperlink
	ForgetLink  *widget.Hyperlink

	ProjectFile string
	ProjectName string
	ProjectPath string
}

func (card *ProjectCard) CreateRenderer() fyne.WidgetRenderer {
	card.ExtendBaseWidget(card)
	return widget.NewSimpleRenderer(card.Container)
}

func (card *ProjectCard) MinSize() fyne.Size {
	card.ExtendBaseWidget(card)
	return card.Container.MinSize()
}

func (card *ProjectCard) SetProject(projectFile string, callback func(string), refresh func()) {
	projectPath := filepath.Dir(projectFile)
	projectName := strings.Replace(filepath.Base(projectFile), ".uproject", "", -1)

	card.Title.Text = projectName
	card.Subtitle.Text = projectPath
	card.ProjectFile = projectFile
	card.ProjectName = projectName
	card.ProjectPath = projectPath
	card.Button.OnTapped = func() {
		println("Opening project: ", card.ProjectFile)
		callback(card.ProjectFile)
		refresh()
	}

	//check if file exists
	_, err := os.Stat(card.ProjectFile)
	if err != nil {
		card.Subtitle.Text = card.Subtitle.Text + " (File not found!)"
		card.Subtitle.Color = theme.ErrorColor()
		card.Button.Disable()
	} else {
		card.Subtitle.Color = theme.ForegroundColor()
		card.Button.Enable()
	}

	card.ExploreLink.OnTapped = func() {
		fmt.Println("Opening in explorer: ", card.ProjectPath)
		open.Start(card.ProjectPath)
	}

	card.ForgetLink.OnTapped = func() {
		fmt.Println("Forgetting project: ", card.ProjectFile)
		config := GetConfig()
		foundIdx := slices.Index(config.RecentProjects, card.ProjectFile)
		if foundIdx != -1 {
			// Remove it from the list
			config.RecentProjects = slices.Delete(config.RecentProjects, foundIdx, foundIdx+1)
		}
		SaveConfig()
		refresh()
	}

	card.Refresh()
	card.Container.Refresh()
}

func MakeCardFromProject(projectFile string) *ProjectCard {

	projectCard := &ProjectCard{
		Container:   container.NewStack(),
		Button:      widget.NewButton("", func() {}),
		Title:       canvas.NewText("Title", theme.ForegroundColor()),
		Subtitle:    canvas.NewText("Subtitle", theme.ForegroundColor()),
		ExploreLink: widget.NewHyperlink("Open in Explorer", &url.URL{}),
		ForgetLink:  widget.NewHyperlink("Forget", &url.URL{}),
	}

	linkLabelStack := container.NewHBox()
	verticalStack := container.NewVBox()
	projectCard.Title.TextSize = theme.TextHeadingSize()
	projectCard.Subtitle.TextSize = theme.TextSubHeadingSize()

	verticalStack.Add(projectCard.Title)
	verticalStack.Add(projectCard.Subtitle)
	verticalStack.Add(linkLabelStack)
	verticalStack.Refresh()

	linkLabelStack.Add(projectCard.ExploreLink)
	linkLabelStack.Add(projectCard.ForgetLink)
	projectCard.Container.Add(projectCard.Button)
	// card for nice padding
	projectCard.Container.Add(widget.NewCard("", "", verticalStack))
	// projectCard.Container.Add(verticalStack)

	//projectCard.SetProject(projectFile, func(string) {}, func() {})
	return projectCard
}
