package view

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/miltoncandelero/ugsg/gui/assets"
)

type ProjectStatus struct {
	// extends widget
	widget.BaseWidget

	Container *fyne.Container

	ProjectTitle       *canvas.Text
	Subtitle           *canvas.Text
	EngineVersion      *canvas.Text
	SwapEngineButton   *widget.Button
	SwapEngineCallback func()

	// Git buttons
	RepoOrigin            *IconText
	RepoUser              *IconText
	FixUserLink           *widget.Hyperlink
	FixUserLinkCallback   func()
	RepoStatus            *IconText
	FixRepoStatusLink     *widget.Hyperlink
	FixRepoStatusCallback func()
	RepoAhead             *IconText
	RepoBehind            *IconText
	RepoWorkingTree       *IconText
	RepoBranch            *widget.Select
	SwapBranchCallback    func(string) string
	ConfigStatus          *IconText
	FixConfigLink         *widget.Hyperlink
	FixConfigLinkCallback func()
	PullButton            *widget.Button
	PullButtonCallback    func()
	SyncButton            *widget.Button
	SyncButtonCallback    func()
	CommitButton          *widget.Button
	CommitButtonCallback  func()

	// Build manager buttons
	BuildStatus                 *IconText
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

	pstatus.ProjectTitle = canvas.NewText("My Cool Project", theme.ForegroundColor())
	pstatus.ProjectTitle.TextSize = theme.TextHeadingSize()
	pstatus.ProjectTitle.Alignment = fyne.TextAlignCenter
	pstatus.Subtitle = canvas.NewText("C:/Somewhere/over/the/rainbow", theme.ForegroundColor())
	pstatus.Subtitle.TextSize = theme.TextSubHeadingSize()
	pstatus.Subtitle.Alignment = fyne.TextAlignCenter
	pstatus.EngineVersion = canvas.NewText("Engine: 5.0.1", theme.ForegroundColor())
	pstatus.SwapEngineButton = widget.NewButtonWithIcon("Swap Engine", theme.SearchReplaceIcon(), nil)
	pstatus.SwapEngineCallback = func() {}

	// Git buttons
	repositoryTitleLabel := canvas.NewText("REPOSITORY", theme.ForegroundColor())
	repositoryTitleLabel.Alignment = fyne.TextAlignCenter
	repositoryTitleLabel.TextSize = theme.TextSubHeadingSize()

	pstatus.RepoOrigin = MakeIconText("Origin", assets.ResGitlabSvg)
	pstatus.RepoUser = MakeIconText("User", theme.AccountIcon())
	pstatus.FixUserLink = widget.NewHyperlink("Fix User", nil)
	pstatus.FixUserLink.OnTapped = func() { pstatus.FixUserLinkCallback() }
	// pstatus.FixUserLinkCallback = func() {  }

	pstatus.RepoStatus = MakeIconText("Status", theme.QuestionIcon())
	pstatus.FixRepoStatusLink = widget.NewHyperlink("Fix Status", nil)
	pstatus.FixRepoStatusLink.OnTapped = func() { pstatus.FixRepoStatusCallback() }
	pstatus.RepoAhead = MakeIconText("32", theme.MenuDropUpIcon())
	pstatus.RepoAhead.SetColor(theme.ColorNameSuccess)
	pstatus.RepoBehind = MakeIconText("12", theme.MenuDropDownIcon())
	pstatus.RepoBehind.SetColor(theme.ColorNameError)
	pstatus.RepoWorkingTree = MakeIconText("13", theme.DocumentSaveIcon())
	pstatus.RepoWorkingTree.SetColor(theme.ColorNameWarning)
	pstatus.RepoBranch = widget.NewSelect([]string{"Branch"}, func(string) {})
	pstatus.SwapBranchCallback = func(string) string { return "" }
	pstatus.ConfigStatus = MakeIconText("Config", theme.QuestionIcon())
	pstatus.FixConfigLink = widget.NewHyperlink("Fix Config", nil)
	pstatus.FixConfigLink.OnTapped = func() { pstatus.FixConfigLinkCallback() }
	// pstatus.FixConfigLinkCallback = func() {}
	pstatus.PullButton = widget.NewButtonWithIcon("Pull", theme.MenuDropDownIcon(), nil)
	pstatus.PullButtonCallback = func() {}
	pstatus.SyncButton = widget.NewButtonWithIcon("Sync", theme.ViewRefreshIcon(), nil)
	pstatus.SyncButtonCallback = func() {}
	pstatus.CommitButton = widget.NewButtonWithIcon("Commit", theme.DocumentSaveIcon(), nil)
	pstatus.CommitButtonCallback = func() {}

	// Build manager buttons
	buildTitleLabel := canvas.NewText("BUILD", theme.ForegroundColor())
	buildTitleLabel.Alignment = fyne.TextAlignCenter
	buildTitleLabel.TextSize = theme.TextSubHeadingSize()

	pstatus.BuildStatus = MakeIconText("Build", theme.QuestionIcon())
	pstatus.DownloadBuildButton = widget.NewButtonWithIcon("Download Build", theme.DownloadIcon(), nil)
	pstatus.DownloadBuildButtonCallback = func() {}
	pstatus.UploadBuildButton = widget.NewButtonWithIcon("Upload Build", theme.UploadIcon(), nil)
	pstatus.UploadBuildButtonCallback = func() {}

	// Unreal tool buttons
	unrealTitleLabel := canvas.NewText("UNREAL TOOLS", theme.ForegroundColor())
	unrealTitleLabel.Alignment = fyne.TextAlignCenter
	unrealTitleLabel.TextSize = theme.TextSubHeadingSize()

	pstatus.GenerateSolutionButton = widget.NewButtonWithIcon("Generate Solution", theme.ViewRefreshIcon(), nil)
	pstatus.GenerateSolutionButtonCallback = func() {}
	pstatus.BuildButton = widget.NewButtonWithIcon("Build", theme.SettingsIcon(), nil)
	pstatus.BuildButtonCallback = func() {}

	pstatus.Container = container.NewStack(container.NewVBox(
		pstatus.ProjectTitle,
		pstatus.Subtitle,
		widget.NewSeparator(),
		container.NewHBox(
			&layout.Spacer{},
			container.NewVBox(
				&layout.Spacer{FixVertical: true},
				repositoryTitleLabel,
				widget.NewSeparator(),
				pstatus.RepoOrigin,
				pstatus.RepoBranch,
				container.NewHBox(pstatus.RepoStatus, pstatus.FixRepoStatusLink, widget.NewSeparator(), pstatus.RepoAhead, pstatus.RepoBehind, widget.NewSeparator(), pstatus.RepoWorkingTree),
				container.NewHBox(pstatus.RepoUser, pstatus.FixUserLink),
				container.NewHBox(pstatus.ConfigStatus, pstatus.FixConfigLink),
				widget.NewSeparator(),
				canvas.NewText("Actions", theme.ForegroundColor()),
				pstatus.CommitButton,
				pstatus.SyncButton,
				pstatus.PullButton,
			),
			&layout.Spacer{},
			container.NewVBox(
				&layout.Spacer{FixVertical: true},
				buildTitleLabel,
				widget.NewSeparator(),
				pstatus.BuildStatus,
				widget.NewSeparator(),
				canvas.NewText("Actions", theme.ForegroundColor()),
				pstatus.DownloadBuildButton,
				pstatus.UploadBuildButton,
			),
			&layout.Spacer{},
			container.NewVBox(
				&layout.Spacer{FixVertical: true},
				unrealTitleLabel,
				widget.NewSeparator(),
				pstatus.EngineVersion,
				pstatus.SwapEngineButton,
				pstatus.GenerateSolutionButton,
				pstatus.BuildButton,
			),
			&layout.Spacer{},
		),
	))

	return pstatus
}
