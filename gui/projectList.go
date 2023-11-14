package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func MakeCardListsFromProjects(projects *[]string, callback func(string)) fyne.CanvasObject {
	var retval *widget.List
	retval = widget.NewList(
		func() int {
			return len(*projects)
		},
		func() fyne.CanvasObject {
			return MakeCardFromProject("")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*ProjectCard).SetProject((*projects)[id], callback, retval.Refresh)
		})
	return retval
}
