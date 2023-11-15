package view

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ProjectStatus struct {
	// extends widget
	widget.BaseWidget

	Container *fyne.Container

	ProjectTitle *canvas.Text
	Subtitle     *canvas.Text

	// Git buttons
	RepoStatus              *canvas.Text
	RepoAheadBehind         *canvas.Text
	RepoBranch              *canvas.Text
	FixConfigButton         *widget.Button
	FixConfigButtonCallback func()
	PullButton              *widget.Button
	PullButtonCallback      func()
	PushButton              *widget.Button
	PushButtonCallback      func()
	CommitButton            *widget.Button
	CommitButtonCallback    func()

	// Build manager buttons
	BuildStatus                 *canvas.Text
	DownloadBuildButton         *widget.Button
	DownloadBuildButtonCallback func()
	UploadBuildButton           *widget.Button
	UploadBuildButtonCallback   func()

	// Unreal tool buttons
	GenerateSolutionButton         *widget.Button
	GenerateSolutionButtonCallback func()
	BuildButton                    *widget.Button
	BuildButtonCallback            func()
}

func (pstatus *ProjectStatus) CreateRenderer() fyne.WidgetRenderer {
	pstatus.ExtendBaseWidget(pstatus)
	return widget.NewSimpleRenderer(pstatus.Container)
}

func MakeProjectStatus(projectFile string) *ProjectStatus {
	pstatus := &ProjectStatus{}

	pstatus.ProjectTitle = canvas.NewText("Project Name", theme.ForegroundColor())
	pstatus.ProjectTitle.TextStyle.Bold = true
	pstatus.ProjectTitle.TextSize = theme.TextSize() * 1.5

	pstatus.Subtitle = canvas.NewText("Project Path", theme.ForegroundColor())
	pstatus.Subtitle.TextSize = theme.TextSize()

	// Git buttons
	pstatus.RepoStatus = canvas.NewText("Repo Status", theme.ForegroundColor())
	pstatus.RepoStatus.TextSize = theme.TextSize()
	pstatus.RepoAheadBehind = canvas.NewText("Ahead/Behind", theme.ForegroundColor())
	pstatus.RepoAheadBehind.TextSize = theme.TextSize()
	pstatus.RepoBranch = canvas.NewText("Branch", theme.ForegroundColor())
	pstatus.RepoBranch.TextSize = theme.TextSize()
	pstatus.FixConfigButton = widget.NewButton("Fix Config", func() { pstatus.FixConfigButtonCallback() })
	pstatus.PullButton = widget.NewButton("Pull", func() { pstatus.PullButtonCallback() })
	pstatus.PushButton = widget.NewButton("Push", func() { pstatus.PushButtonCallback() })
	pstatus.CommitButton = widget.NewButton("Commit", func() { pstatus.CommitButtonCallback() })

	// Build manager buttons
	pstatus.BuildStatus = canvas.NewText("Build Status", theme.ForegroundColor())
	pstatus.BuildStatus.TextSize = theme.TextSize()
	pstatus.DownloadBuildButton = widget.NewButton("Download Build", func() { pstatus.DownloadBuildButtonCallback() })
	pstatus.UploadBuildButton = widget.NewButton("Upload Build", func() { pstatus.UploadBuildButtonCallback() })

	// Unreal tool buttons
	pstatus.GenerateSolutionButton = widget.NewButton("Generate Solution", func() { pstatus.GenerateSolutionButtonCallback() })
	pstatus.BuildButton = widget.NewButton("Build", func() { pstatus.BuildButtonCallback() })

	pstatus.Container = container.NewVBox(
		pstatus.ProjectTitle,
		pstatus.Subtitle,
		pstatus.RepoStatus,
		pstatus.RepoAheadBehind,
		pstatus.RepoBranch,
		pstatus.FixConfigButton,
		pstatus.PullButton,
		pstatus.PushButton,
		pstatus.CommitButton,
		pstatus.BuildStatus,
		pstatus.DownloadBuildButton,
		pstatus.UploadBuildButton,
		pstatus.GenerateSolutionButton,
		pstatus.BuildButton,
	)

	return pstatus
}
