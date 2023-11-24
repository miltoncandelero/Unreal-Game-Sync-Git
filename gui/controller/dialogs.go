package controller

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func CallFuncShowDialogOnError(f func() error) {
	err := f()
	if err != nil {
		ShowErrorDialog(err)
	}
}

func ShowErrorDialog(err error) {
	if err == nil {
		return
	}
	dialog.ShowError(err, GetApp().Window)
}

func ShowWarningDialog(title string, body string) {
	dialog.ShowInformation(title, body, GetApp().Window)
}

func ShowLoadingDialog(title string) *dialog.CustomDialog {
	bar := widget.NewProgressBarInfinite()
	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(fyne.NewSize(200, 0))
	stack := container.NewStack(rect, bar)
	d := dialog.NewCustomWithoutButtons(title, stack, GetApp().Window)
	d.Show()
	return d
}

func ShowUsernameEmailDialog(provider string, callback func(string, string) error) {
	usernameWidget := widget.NewEntry()
	emailWidget := widget.NewEntry()
	dialog.ShowForm(
		"Enter the username and email for your "+provider+" account",
		"OK",
		"Cancel",
		[]*widget.FormItem{ // we can specify items in the constructor
			{Text: "Username", Widget: usernameWidget},
			{Text: "Email", Widget: emailWidget},
			{Text: "", Widget: layout.NewSpacer()},
		},
		func(ok bool) {
			if !ok {
				return
			} else {
				err := callback(usernameWidget.Text, emailWidget.Text)
				if err != nil {
					ShowErrorDialog(err)
				}
			}
		}, GetApp().Window)
}
