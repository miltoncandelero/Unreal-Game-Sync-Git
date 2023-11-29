package view

import (
	"image/color"
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/miltoncandelero/ugsg/core"
)

type LockedFileItem struct {
	// extends widget
	widget.BaseWidget

	Container *fyne.Container

	AssociatedFile *core.LockDatum

	parent *LockedFilesList

	Checkbox  *widget.Check
	FileLabel *widget.Label
	UserLabel *widget.Label
}

func (this *LockedFileItem) CreateRenderer() fyne.WidgetRenderer {
	this.ExtendBaseWidget(this)
	return widget.NewSimpleRenderer(this.Container)
}
func (this *LockedFileItem) Tapped(_ *fyne.PointEvent) {
	this.Checkbox.Checked = !this.Checkbox.Checked
	this.Refresh()
}
func (this *LockedFileItem) Recycle(newDatum *core.LockDatum) {
	this.AssociatedFile = newDatum
	this.Checkbox.Checked = this.parent.Selected[newDatum.ID]
	if newDatum.AssociatedMap != "" {
		this.FileLabel.SetText(newDatum.AssociatedMap + " (External Actor)")
	} else {
		this.FileLabel.SetText(newDatum.Path)
	}
	this.UserLabel.SetText(newDatum.Owner.Name)
	this.Refresh()
}

func MakeLockedItem(listRef *LockedFilesList) *LockedFileItem {
	retval := &LockedFileItem{}
	retval.parent = listRef
	retval.Checkbox = widget.NewCheck("", func(b bool) {
		if b {
			retval.parent.Selected[retval.AssociatedFile.ID] = true
		} else {
			delete(retval.parent.Selected, retval.AssociatedFile.ID)
		}
	})
	retval.FileLabel = widget.NewLabel("")
	retval.FileLabel.Truncation = fyne.TextTruncateEllipsis
	retval.UserLabel = widget.NewLabel("")
	retval.UserLabel.Importance = widget.LowImportance
	retval.Container = container.NewBorder(nil, nil, retval.Checkbox, retval.UserLabel, retval.FileLabel)
	retval.ExtendBaseWidget(retval)
	retval.Refresh()
	return retval
}

type LockedFilesList struct {
	fyneWidget *widget.List

	Container *fyne.Container

	LockedFiles     []core.LockDatum
	UnchangedLocked []core.LockDatum
	Selected        map[string]bool
}

func MakeLockedFilesList() *LockedFilesList {
	retval := &LockedFilesList{}

	retval.LockedFiles = make([]core.LockDatum, 0)
	retval.Selected = make(map[string]bool, 0)

	retval.fyneWidget = widget.NewList(
		func() int {
			return len(retval.LockedFiles)
		},
		func() fyne.CanvasObject {
			return MakeLockedItem(retval)
		},
		func(id widget.ListItemID, o fyne.CanvasObject) {

			c := o.(*LockedFileItem)
			c.Recycle(&retval.LockedFiles[id])
		})

	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(fyne.NewSize(800, 600))
	retval.Container = container.NewStack(rect, retval.fyneWidget)

	return retval
}

type LockedDialog struct {
	*dialog.CustomDialog
	lockedList         *LockedFilesList
	selectByUserList   *widget.Select
	UnlockFilesCalback func([]core.LockDatum, bool)
	RefreshCallback    func()
}

func (this *LockedDialog) UpdateData(lockedFiles []core.LockDatum, unchangedLocked []core.LockDatum) {
	this.lockedList.LockedFiles = lockedFiles
	this.lockedList.UnchangedLocked = unchangedLocked
	this.lockedList.Selected = make(map[string]bool, len(lockedFiles))
	this.lockedList.fyneWidget.Refresh()

	users := make([]string, 0)
	users = append(users, "(none)")
	for _, file := range lockedFiles {
		if !slices.Contains(users, file.Owner.Name) {
			users = append(users, file.Owner.Name)
		}
	}
	this.selectByUserList.Options = users
	this.selectByUserList.Refresh()
}

func (this *LockedDialog) GetSelected() []core.LockDatum {
	retval := make([]core.LockDatum, 0, len(this.lockedList.Selected))
	for k, v := range this.lockedList.Selected {
		if v {
			for _, file := range this.lockedList.LockedFiles {
				if file.ID == k {
					retval = append(retval, file)
					break
				}
			}
		}
	}
	return retval
}

func (this *LockedDialog) SelectAll() {
	this.selectByUserList.SetSelectedIndex(0)
	for _, file := range this.lockedList.LockedFiles {
		this.lockedList.Selected[file.ID] = true
	}
	this.lockedList.fyneWidget.Refresh()
}

func (this *LockedDialog) SelectNone() {
	this.selectByUserList.SetSelectedIndex(0)
	for _, file := range this.lockedList.LockedFiles {
		delete(this.lockedList.Selected, file.ID)
	}
	this.lockedList.fyneWidget.Refresh()
}

func (this *LockedDialog) SelectUnchanged() {
	this.selectByUserList.SetSelectedIndex(0)
	this.SelectNone()
	for _, file := range this.lockedList.LockedFiles {
		for _, unchanged := range this.lockedList.UnchangedLocked {
			if file.ID == unchanged.ID {
				this.lockedList.Selected[file.ID] = true
				break
			}
		}
	}
	this.lockedList.fyneWidget.Refresh()
}

func (this *LockedDialog) SelectByUser(user string) {

	for _, file := range this.lockedList.LockedFiles {
		delete(this.lockedList.Selected, file.ID)
	}

	for _, file := range this.lockedList.LockedFiles {
		if user == file.Owner.Name {
			this.lockedList.Selected[file.ID] = true
		} else {
			delete(this.lockedList.Selected, file.ID)
		}
	}
	this.lockedList.fyneWidget.Refresh()
}

func MakeLockedDialog(window fyne.Window) *LockedDialog {

	retval := &LockedDialog{}

	closeBtn := widget.NewButton("Close", nil)
	forceCheckbox := widget.NewCheck("Force unlock?", nil)
	unlockSelected := widget.NewButton("Unlock selected", func() {
		retval.UnlockFilesCalback(retval.GetSelected(), forceCheckbox.Checked)
	})
	actionsContainer := container.NewVBox(unlockSelected, forceCheckbox)
	bottomContainer := container.NewBorder(actionsContainer, nil, nil, closeBtn, nil)

	selectAllBtn := widget.NewButton("Select all", retval.SelectAll)
	selectNoneBtn := widget.NewButton("Select none", retval.SelectNone)
	selectUnchangedBtn := widget.NewButton("Select unchanged", retval.SelectUnchanged)
	selectByUserLabel := widget.NewLabel("Select by user:")
	selectByUserList := widget.NewSelect([]string{"All"}, retval.SelectByUser)
	selectAllNoneContainer := container.NewHBox(selectAllBtn, selectNoneBtn, selectUnchangedBtn, selectByUserLabel, selectByUserList)
	topContainer := container.NewBorder(nil, nil, selectAllNoneContainer, nil, nil)

	lockedList := MakeLockedFilesList()
	border := container.NewBorder(topContainer, bottomContainer, nil, nil, lockedList.Container)

	dialog := dialog.NewCustomWithoutButtons("Locked Files", border, window)
	closeBtn.OnTapped = func() {
		dialog.Hide()
		retval.RefreshCallback()
	}

	retval.CustomDialog = dialog
	retval.lockedList = lockedList
	retval.selectByUserList = selectByUserList

	// retval.UpdateData(lockedFiles, unchangedLocked)

	return retval
}
