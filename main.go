package main

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/miltoncandelero/ugsg/core"
)

const REPO = "../../Unreal/UzKartNew/"

type rightClickTable struct {
	widget.Table
	menu     *widget.PopUpMenu
	selected widget.TableCellID
}

func (t *rightClickTable) TappedSecondary(e *fyne.PointEvent) {
	t.Table.Tapped(e)
	t.menu.ShowAtPosition(e.AbsolutePosition)
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("List Data")
	myWindow.Resize(fyne.NewSize(1000, 600))

	commitData, _ := core.GetRepoBranchInfo(REPO, "origin/master")

	// list := widget.NewTableWithHeaders(
	list := &rightClickTable{
		Table: *widget.NewTableWithHeaders(
			func() (int, int) {
				return len(commitData), 4
			},
			func() fyne.CanvasObject {
				x := widget.NewLabel("wide content")
				x.Truncation = fyne.TextTruncateEllipsis
				return x
			},
			func(i widget.TableCellID, o fyne.CanvasObject) {
				lbl := o.(*widget.Label)
				switch commitData[i.Row].DistanceFromHead {
				case 1:
					lbl.Importance = widget.LowImportance
				case 0:
					lbl.Importance = widget.HighImportance
				case -1:
					lbl.Importance = widget.MediumImportance
				}

				switch i.Col {
				case 0:
					lbl.SetText(commitData[i.Row].Hash)
				case 1:
					lbl.SetText(commitData[i.Row].Msg)
				case 2:
					lbl.SetText(commitData[i.Row].User)
				case 3:

					lbl.SetText("")

					for emoji := range commitData[i.Row].Emoji {
						lbl.SetText(lbl.Text + emoji)
					}

				}
			},
		),
	}
	menuItem1 := fyne.NewMenuItem("Checkout", func() {
		fmt.Printf("list.selected: %v\n", list.selected)
	})
	menu := fyne.NewMenu("File", menuItem1)

	popUpMenu := widget.NewPopUpMenu(menu, myWindow.Canvas())

	list.menu = popUpMenu

	list.ExtendBaseWidget(list)
	list.ShowHeaderRow = true
	list.ShowHeaderColumn = false

	list.UpdateHeader = func(id widget.TableCellID, template fyne.CanvasObject) {
		switch id.Col {
		case 0:
			template.(*widget.Label).SetText("Sha")
		case 1:
			template.(*widget.Label).SetText("Message")
		case 2:
			template.(*widget.Label).SetText("User")
		case 3:
			template.(*widget.Label).SetText("Commit type")
		}
	}

	list.OnSelected = func(id widget.TableCellID) {
		log.Println("selected", id)
		list.selected = id
	}

	list.SetColumnWidth(0, 500)
	list.SetColumnWidth(1, 500)
	list.SetColumnWidth(2, 500)
	list.SetColumnWidth(3, 500)

	myWindow.SetContent(container.NewStack(list))
	myWindow.ShowAndRun()
}
