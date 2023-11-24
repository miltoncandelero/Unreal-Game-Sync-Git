package view

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/miltoncandelero/ugsg/core"
	"github.com/miltoncandelero/ugsg/gui/assets"

	"fyne.io/x/fyne/layout"
)

type CommitList struct {
	fyneWidget *widget.Tree

	Container *fyne.Container

	SelectedCommit *core.CommitDatum

	datesMap map[string][]string
	datesArr []string
	hashMap  map[string]*core.CommitDatum
}

func (this *CommitList) UpdateCommits(commits []*core.CommitDatum) {
	this.datesMap = make(map[string][]string)
	this.datesArr = make([]string, 0)
	this.hashMap = make(map[string]*core.CommitDatum, len(commits))

	for _, commit := range commits {
		_, ok := this.datesMap[commit.Date.Local().Format("Mon Jan _2 2006")]
		if !ok {
			this.datesMap[commit.Date.Local().Format("Mon Jan _2 2006")] = make([]string, 0)
			this.datesArr = append(this.datesArr, commit.Date.Local().Format("Mon Jan _2 2006"))
		}
		this.datesMap[commit.Date.Local().Format("Mon Jan _2 2006")] = append(this.datesMap[commit.Date.Local().Format("Mon Jan _2 2006")], commit.Hash)
		this.hashMap[commit.Hash] = commit
	}
	this.fyneWidget.OpenAllBranches()
	this.fyneWidget.Refresh()
}

func MakePopupMenu(parentTree *CommitList, checkoutCallback func(string), resetCallback func(string), windowCanvas fyne.Canvas) *widget.PopUpMenu {
	checkoutItem := fyne.NewMenuItem("Flashback to here (Checkout)", func() {
		fmt.Printf("parentTree.SelectedCommit: %v\n", parentTree.SelectedCommit)
		checkoutCallback(parentTree.SelectedCommit.Hash)
	})

	resetItem := fyne.NewMenuItem("Time travel to here (reset --hard)", func() {
		fmt.Printf("parentTree.SelectedCommit: %v\n", parentTree.SelectedCommit)
		resetCallback(parentTree.SelectedCommit.Hash)
	})
	menu := fyne.NewMenu("Time travel menu", checkoutItem, resetItem)

	popUpMenu := widget.NewPopUpMenu(menu, windowCanvas)
	return popUpMenu
}

func MakeCommitList(checkoutCallback func(string), resetCallback func(string), windowCanvas fyne.Canvas) *CommitList {
	tree := &CommitList{}

	popupMenu := MakePopupMenu(tree, checkoutCallback, resetCallback, windowCanvas)

	tree.fyneWidget = widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			if id == "" {
				return tree.datesArr
			}

			return tree.datesMap[id]
		},
		func(id widget.TreeNodeID) bool {
			if id == "" {
				return true
			}
			_, ok := tree.datesMap[id]
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
				UpdateCommitWidget(tree.hashMap[id], o)
			}
		})
	tree.Container = container.NewBorder(MakeHeaderWidget(), nil, nil, nil, tree.fyneWidget)
	return tree
}

type CommitItem struct {
	// extends widget
	widget.BaseWidget

	Container *fyne.Container

	Commit *core.CommitDatum

	Hash         *widget.Label
	Icon         *widget.Icon
	iconResource *theme.ThemedResource
	Date         *widget.Label
	User         *widget.Label
	Msg          *widget.Label
	Menu         *widget.PopUpMenu
	ParentTree   *CommitList
}

func (this *CommitItem) SetIcon(icon fyne.Resource) {
	this.ExtendBaseWidget(this)
	this.iconResource = theme.NewThemedResource(icon)
	this.iconResource.ColorName = theme.ColorNameForeground
	this.Icon.Resource = this.iconResource
	this.Container.Refresh()
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

func MakeHeaderWidget() *fyne.Container {
	layout := layout.NewHPortion([]float64{1, 1, 1, 2, 20})
	hbox := container.New(layout,
		(widget.NewLabel("Hash")),
		(widget.NewLabel("Type")),
		(widget.NewLabel("Date")),
		(widget.NewLabel("User")),
		(widget.NewLabel("Msg")),
	)

	for _, o := range hbox.Objects {
		o.(*widget.Label).Alignment = fyne.TextAlignLeading
		o.(*widget.Label).Truncation = fyne.TextTruncateEllipsis
	}

	padding := canvas.NewRectangle(color.Transparent)
	padding.SetMinSize(fyne.NewSize(GetChildPadding(), 0))
	return container.NewBorder(nil, nil, padding, nil, hbox)
}

func MakeCommitWidget(menu *widget.PopUpMenu, parentTree *CommitList) fyne.CanvasObject {
	layout := layout.NewHPortion([]float64{1, 1, 1, 2, 20})
	hbox := container.New(layout,
		(widget.NewLabel("")),
		(widget.NewIcon(theme.QuestionIcon())),
		(widget.NewLabel("")),
		(widget.NewLabel("")),
		(widget.NewLabel("")),
	)

	retval := &CommitItem{
		Container:  hbox,
		Hash:       hbox.Objects[0].(*widget.Label),
		Icon:       hbox.Objects[1].(*widget.Icon),
		Date:       hbox.Objects[2].(*widget.Label),
		User:       hbox.Objects[3].(*widget.Label),
		Msg:        hbox.Objects[4].(*widget.Label),
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
	cItem.Hash.SetText(commit.Hash[:6])
	cItem.Date.SetText(commit.Date.Local().Format(time.Kitchen))
	cItem.User.SetText(commit.User)
	cItem.Msg.SetText(commit.Msg)
	cItem.Commit = commit

	if commit.ContentChange == commit.SourceChange {
		if commit.ContentChange {
			cItem.SetIcon(assets.ResMixedSvg)
		} else {
			cItem.SetIcon(theme.QuestionIcon())
		}
	} else {
		if commit.ContentChange {
			cItem.SetIcon(assets.ResContentSvg)
		} else {
			cItem.SetIcon(assets.ResCodeSvg)
		}
	}

	o.Refresh()
}

func GetChildPadding() float32 {
	return 2*(theme.IconInlineSize()+theme.Padding()) + theme.Padding()
}
