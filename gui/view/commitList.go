package view

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/miltoncandelero/ugsg/core"

	"fyne.io/x/fyne/layout"
)

type CommitList struct {
	// extends widget
	fyneWidget *widget.Tree

	SelectedCommit *core.CommitDatum
}

func MakePopupMenu(parentTree *CommitList) *widget.PopUpMenu {
	checkoutItem := fyne.NewMenuItem("Flashback to here (Checkout)", func() {
		fmt.Printf("parentTree.SelectedCommit: %v\n", parentTree.SelectedCommit)
	})

	resetItem := fyne.NewMenuItem("Time travel to here (reset --hard)", func() {
		fmt.Printf("parentTree.SelectedCommit: %v\n", parentTree.SelectedCommit)
	})
	menu := fyne.NewMenu("Time travel menu", checkoutItem, resetItem)

	// TODO: Fix the ugly all windows thingy
	popUpMenu := widget.NewPopUpMenu(menu, fyne.CurrentApp().Driver().AllWindows()[0].Canvas())
	return popUpMenu
}

func MakeCommitList(commits []*core.CommitDatum) fyne.CanvasObject {
	datesMap := make(map[string][]string)
	datesArr := make([]string, 0)
	hashMap := make(map[string]*core.CommitDatum, len(commits))

	for _, commit := range commits {
		_, ok := datesMap[commit.Date.Local().Format("Mon Jan _2 2006")]
		if !ok {
			datesMap[commit.Date.Local().Format("Mon Jan _2 2006")] = make([]string, 0)
			datesArr = append(datesArr, commit.Date.Local().Format("Mon Jan _2 2006"))
		}
		datesMap[commit.Date.Local().Format("Mon Jan _2 2006")] = append(datesMap[commit.Date.Local().Format("Mon Jan _2 2006")], commit.Hash)
		hashMap[commit.Hash] = commit
	}

	tree := &CommitList{}
	popupMenu := MakePopupMenu(tree)

	tree.fyneWidget = widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			if id == "" {
				return datesArr
			}

			return datesMap[id]
		},
		func(id widget.TreeNodeID) bool {
			if id == "" {
				return true
			}
			_, ok := datesMap[id]
			return ok
		},
		func(branch bool) fyne.CanvasObject {
			if branch {
				return widget.NewLabel("Branch template")
			}
			return MakeCommitWidget(popupMenu, tree)
		},
		func(id widget.TreeNodeID, branch bool, o fyne.CanvasObject) {
			if branch {
				o.(*widget.Label).SetText(id)
			} else {
				UpdateCommitWidget(hashMap[id], o)
			}
		})
	tree.fyneWidget.OpenAllBranches()
	return tree.fyneWidget
}

type CommitItem struct {
	// extends widget
	widget.BaseWidget

	Container *fyne.Container

	Commit *core.CommitDatum

	Hash       *widget.Label
	Date       *widget.Label
	User       *widget.Label
	Msg        *widget.Label
	Menu       *widget.PopUpMenu
	ParentTree *CommitList
}

func (citem *CommitItem) CreateRenderer() fyne.WidgetRenderer {
	citem.ExtendBaseWidget(citem)
	return widget.NewSimpleRenderer(citem.Container)
}

func (t *CommitItem) TappedSecondary(e *fyne.PointEvent) {
	fmt.Println("Right clicked me :D")
	t.ParentTree.fyneWidget.UnselectAll()
	t.ParentTree.fyneWidget.Select(t.Commit.Hash)
	t.ParentTree.SelectedCommit = t.Commit
	t.Menu.ShowAtPosition(e.AbsolutePosition)
}

func (t *CommitItem) Tapped(e *fyne.PointEvent) {
	t.ParentTree.fyneWidget.UnselectAll()
	t.ParentTree.fyneWidget.Select(t.Commit.Hash)
	t.ParentTree.SelectedCommit = t.Commit
}

func MakeCommitWidget(menu *widget.PopUpMenu, parentTree *CommitList) fyne.CanvasObject {
	layout := layout.NewHPortion([]float64{1, 1, 1, 10})
	hbox := container.New(layout,
		(widget.NewLabel("")),
		(widget.NewLabel("")),
		(widget.NewLabel("")),
		(widget.NewLabel("")),
	)

	retval := &CommitItem{
		Container:  hbox,
		Hash:       hbox.Objects[0].(*widget.Label),
		Date:       hbox.Objects[1].(*widget.Label),
		User:       hbox.Objects[2].(*widget.Label),
		Msg:        hbox.Objects[3].(*widget.Label),
		Menu:       menu,
		ParentTree: parentTree,
	}
	retval.ExtendBaseWidget(retval)
	retval.Hash.Alignment = fyne.TextAlignLeading
	retval.Hash.Truncation = fyne.TextTruncateEllipsis
	retval.Date.Alignment = fyne.TextAlignLeading
	retval.Date.Truncation = fyne.TextTruncateEllipsis
	retval.User.Alignment = fyne.TextAlignLeading
	retval.User.Truncation = fyne.TextTruncateEllipsis
	retval.Msg.Alignment = fyne.TextAlignLeading
	retval.Msg.Truncation = fyne.TextTruncateEllipsis
	return retval
}

func UpdateCommitWidget(commit *core.CommitDatum, o fyne.CanvasObject) {
	cItem := o.(*CommitItem)
	cItem.Hash.SetText(commit.Hash)
	cItem.Date.SetText(commit.Date.Local().Format(time.Kitchen))
	cItem.User.SetText(commit.User)
	cItem.Msg.SetText(commit.Msg)
	cItem.Commit = commit
	o.Refresh()
}
