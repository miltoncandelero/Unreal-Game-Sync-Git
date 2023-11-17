package controller

import (
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"github.com/miltoncandelero/ugsg/core"
	"github.com/miltoncandelero/ugsg/gui/assets"
	"github.com/miltoncandelero/ugsg/gui/view"
	"github.com/ncruces/zenity"
)

func UProjectOpened(uprojectPath string) {
	fmt.Println("Opening", uprojectPath)
	repoPath := filepath.Dir(uprojectPath)
	if !core.IsPathRepo(repoPath) {
		// This is not a repo! panic

		dialog.ShowError(fmt.Errorf("This is not a git repository"), GetApp().Window)
		return
	}

	config := GetConfig()

	foundIdx := slices.Index(config.RecentProjects, uprojectPath)
	if foundIdx != -1 {
		// Remove it from the list
		config.RecentProjects = slices.Delete(config.RecentProjects, foundIdx, foundIdx+1)
	}

	config.RecentProjects = append([]string{uprojectPath}, config.RecentProjects...)
	SaveConfig()

	projectStatus := view.MakeProjectStatus(uprojectPath)
	projectStatus.ProjectTitle.Text = strings.ReplaceAll(filepath.Base(uprojectPath), ".uproject", "")
	projectStatus.Subtitle.Text = uprojectPath
	projectStatus.RepoOrigin.Text.Text = core.GetRepoOrigin(repoPath)
	switch core.GetGitProviderName(repoPath) {
	case "GitHub":
		projectStatus.RepoOrigin.SetIcon(assets.ResGithubSvg)
	case "GitLab":
		projectStatus.RepoOrigin.SetIcon(assets.ResGitlabSvg)
	case "Gitea":
		projectStatus.RepoOrigin.SetIcon(assets.ResGiteaSvg)
	default:
		projectStatus.RepoOrigin.SetIcon(assets.ResGitSvg)
	}
	projectStatus.RepoUser.Text.Text = core.GetUsernameFromRepo(repoPath) + " (" + core.GetUserEmailFromRepo(repoPath) + ")"
	ahead, behind, _ := core.GetAheadBehind(repoPath)
	projectStatus.RepoAhead.Text.Text = strconv.Itoa(ahead)
	projectStatus.RepoBehind.Text.Text = strconv.Itoa(behind)
	switch core.GetGitConfigStatus(repoPath) {
	case core.FILE_MISSING:
		projectStatus.ConfigStatus.Text.Text = "Missing"
		projectStatus.ConfigStatus.SetColor(theme.ColorNameWarning)
		projectStatus.ConfigStatus.SetIcon(theme.QuestionIcon())
	case core.FILE_EXIST_BUT_NOT_LINKED:
		projectStatus.ConfigStatus.Text.Text = "Not linked"
		projectStatus.ConfigStatus.SetColor(theme.ColorNameError)
		projectStatus.ConfigStatus.SetIcon(theme.ErrorIcon())
	case core.FILE_LINKED:
		projectStatus.ConfigStatus.Text.Text = "Linked"
		projectStatus.ConfigStatus.SetColor(theme.ColorNameSuccess)
		projectStatus.ConfigStatus.SetIcon(theme.ConfirmIcon())
	}

	appendProjectToMainWindow(projectStatus, uprojectPath)
}

func openFilePickerUproject() {
	file, err := zenity.SelectFile(
		zenity.Filename("./"),
		zenity.Title("Select Unreal Engine Project file"),
		zenity.Modal(),
		zenity.FileFilters{
			{
				Name:     "Unreal Engine Project files",
				Patterns: []string{"*.uproject"},
				CaseFold: false,
			},
		},
	)
	if err == nil {
		UProjectOpened(file)
	}
}

func ForgetProject(projectFile string) {
	config := GetConfig()
	foundIdx := slices.Index(config.RecentProjects, projectFile)
	if foundIdx != -1 {
		// Remove it from the list
		config.RecentProjects = slices.Delete(config.RecentProjects, foundIdx, foundIdx+1)
	}
	SaveConfig()
}
