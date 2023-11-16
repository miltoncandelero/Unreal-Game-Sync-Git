package view

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type IconText struct {
	// extends widget
	widget.BaseWidget

	Container *fyne.Container

	Text         *canvas.Text
	Icon         *widget.Icon
	iconResource *theme.ThemedResource
	colorName    fyne.ThemeColorName
}

func (this *IconText) CreateRenderer() fyne.WidgetRenderer {
	this.ExtendBaseWidget(this)
	return widget.NewSimpleRenderer(this.Container)
}

func (this *IconText) SetColor(colorName fyne.ThemeColorName) {
	this.ExtendBaseWidget(this)
	this.iconResource.ColorName = colorName
	this.colorName = colorName
	this.Text.Color = fyne.CurrentApp().Settings().Theme().Color(colorName, fyne.CurrentApp().Settings().ThemeVariant())
	this.Container.Refresh()
}

func (this *IconText) SetText(text string) {
	this.ExtendBaseWidget(this)
	this.Text.Text = text
	this.Container.Refresh()
}

func (this *IconText) SetIcon(icon fyne.Resource) {
	this.ExtendBaseWidget(this)
	this.iconResource = theme.NewThemedResource(icon)
	this.iconResource.ColorName = this.colorName
	this.Icon.Resource = this.iconResource
	this.Container.Refresh()
}

func MakeIconText(text string, icon fyne.Resource) *IconText {
	retval := &IconText{}
	retval.Container = container.NewHBox()
	retval.Text = canvas.NewText(text, theme.ForegroundColor())
	retval.iconResource = theme.NewThemedResource(icon)
	retval.Icon = widget.NewIcon(retval.iconResource)
	retval.ExtendBaseWidget(retval)
	retval.Container.Add(retval.Icon)
	retval.Container.Add(retval.Text)
	return retval
}
