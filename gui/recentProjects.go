package gui

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/skratchdot/open-golang/open"
)

type ProjectCard struct {
	// extends widget
	widget.BaseWidget

	Container       *fyne.Container
	Card            *widget.Card
	Button          *widget.Button
	ExploreLink     *widget.Hyperlink
	HorizontalStack *fyne.Container

	ProjectFile string
	ProjectName string
	ProjectPath string
}

type ProjectList struct {
	// extends widget
	widget.List

	Projects *[]string
}

func (card *ProjectCard) CreateRenderer() fyne.WidgetRenderer {
	card.ExtendBaseWidget(card)
	return widget.NewSimpleRenderer(card.Container)
}

func (card *ProjectCard) MinSize() fyne.Size {
	card.ExtendBaseWidget(card)
	return card.Container.MinSize()
}

func (card *ProjectCard) Refresh() {
	card.ExtendBaseWidget(card)
	card.Container.Refresh()
}

func (card *ProjectCard) Size() fyne.Size {
	card.ExtendBaseWidget(card)
	return card.Container.Size()
}

func (card *ProjectCard) Layout(size fyne.Size) {
	card.ExtendBaseWidget(card)
	card.Container.Layout.Layout(card.Container.Objects, size)
}

func (card *ProjectCard) SetProject(projectFile string) {
	projectPath := filepath.Dir(projectFile)
	projectName := strings.Replace(filepath.Base(projectFile), ".uproject", "", -1)

	card.Card.Title = projectName
	card.Card.Subtitle = projectPath
	card.ProjectFile = projectFile
	card.ProjectName = projectName
	card.ProjectPath = projectPath
	card.Button.OnTapped = func() {
		println("Opening project: ", card.ProjectFile)
	}

	card.ExploreLink.OnTapped = func() {
		fmt.Println("Opening in explorer: ", card.ProjectPath)
		open.Start(card.ProjectPath)
	}

	card.Refresh()
	card.Container.Refresh()
}

func MakeCardFromProject(projectFile string) *ProjectCard {

	projectCard := &ProjectCard{
		Container:       container.NewStack(),
		Button:          widget.NewButton("", func() {}),
		Card:            widget.NewCard("", "", nil),
		ExploreLink:     widget.NewHyperlink("Open in Explorer", &url.URL{}),
		HorizontalStack: container.NewHBox(),
	}

	projectCard.HorizontalStack.Add(projectCard.ExploreLink)
	projectCard.Card.Content = projectCard.HorizontalStack
	projectCard.Container.Add(projectCard.Button)
	projectCard.Container.Add(projectCard.Card)

	projectCard.SetProject(projectFile)
	return projectCard
}

func MakeCardListsFromProjects(projects *[]string) *ProjectList {
	list := &ProjectList{
		List: *widget.NewList(
			func() int {
				return len(*projects)
			},
			func() fyne.CanvasObject {
				return MakeCardFromProject("")
			},
			func(id widget.ListItemID, item fyne.CanvasObject) {
				item.(*ProjectCard).SetProject((*projects)[id])
			}),
		Projects: projects,
	}
	list.ExtendBaseWidget(list)
	return list
}
