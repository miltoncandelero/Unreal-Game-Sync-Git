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

	refreshProjectStatus(projectStatus, repoPath)
	refreshUserData(projectStatus, repoPath)
	refreshConfigStatus(projectStatus, repoPath)

	appendProjectToMainWindow(projectStatus, uprojectPath)
}

func refreshUserData(projectStatus *view.ProjectStatus, repoPath string) {
	if core.NeedsUsernameFix(repoPath) {
		projectStatus.RepoUser.Text.Text = "Username missing!"
		projectStatus.RepoUser.SetIcon(theme.ErrorIcon())
		projectStatus.RepoUser.SetColor(theme.ColorNameError)
		projectStatus.FixUserLink.Text = "Fix"
	} else {
		projectStatus.RepoUser.Text.Text = core.GetUsernameFromRepo(repoPath) + " (" + core.GetUserEmailFromRepo(repoPath) + ")"
		projectStatus.RepoUser.SetIcon(theme.AccountIcon())
		projectStatus.RepoUser.SetColor(theme.ColorNameForeground)
		projectStatus.FixUserLink.Text = "Change"
	}
	projectStatus.FixUserLinkCallback = func() {
		ShowUsernameEmailDialog(core.GetGitProviderName(repoPath),
			func(username string, email string) error {
				err := core.SetUsernameAndEmail(repoPath, username, email)
				if err != nil {
					return err
				}
				refreshUserData(projectStatus, repoPath)
				return nil
			})
	}
}

func refreshConfigStatus(projectStatus *view.ProjectStatus, repoPath string) {
	switch core.GetGitConfigStatus(repoPath) {
	case core.CONFIG_STATUS_MISSING:
		projectStatus.ConfigStatus.Text.Text = ".gitconfig missing"
		projectStatus.ConfigStatus.SetColor(theme.ColorNameWarning)
		projectStatus.ConfigStatus.SetIcon(theme.QuestionIcon())
		projectStatus.FixConfigLink.Text = "Create"
		projectStatus.FixConfigLink.Show()
		projectStatus.FixConfigLinkCallback = func() {
			err := core.CreateGitConfig(repoPath)
			if err != nil {
				ShowErrorDialog(err)
			}
			refreshConfigStatus(projectStatus, repoPath)
		}
	case core.CONFIG_STATUS_NOT_LINKED:
		projectStatus.ConfigStatus.Text.Text = ".gitconfig found but not installed!"
		projectStatus.ConfigStatus.SetColor(theme.ColorNameError)
		projectStatus.ConfigStatus.SetIcon(theme.ErrorIcon())
		projectStatus.FixConfigLink.Text = "Fix"
		projectStatus.FixConfigLink.Show()
		projectStatus.FixConfigLinkCallback = func() {
			err := core.LinkGitConfig(repoPath)
			if err != nil {
				ShowErrorDialog(err)
			}
			refreshConfigStatus(projectStatus, repoPath)
		}
	case core.CONFIG_STATUS_LINKED:
		projectStatus.ConfigStatus.Text.Text = ".gitconfig linked"
		projectStatus.ConfigStatus.SetColor(theme.ColorNameSuccess)
		projectStatus.ConfigStatus.SetIcon(theme.ConfirmIcon())
		projectStatus.FixConfigLink.Hide()
	}
}

func refreshProjectStatus(projectStatus *view.ProjectStatus, repoPath string) {
	status := core.GetGitStatus(repoPath)
	switch status {
	case core.GIT_STATUS_OK:
		projectStatus.RepoStatus.Text.Text = "Repo ok"
		projectStatus.RepoStatus.SetColor(theme.ColorNameSuccess)
		projectStatus.RepoStatus.SetIcon(theme.ConfirmIcon())
		projectStatus.FixRepoStatusLink.Hide()
	case core.GIT_STATUS_SHALLOW:
		projectStatus.RepoStatus.Text.Text = "Repo is shallow!"
		projectStatus.RepoStatus.SetColor(theme.ColorNameWarning)
		projectStatus.RepoStatus.SetIcon(theme.WarningIcon())
		projectStatus.FixRepoStatusLink.Text = "unshallow"
		projectStatus.FixRepoStatusLink.Show()
		projectStatus.FixRepoStatusCallback = func() {
			err := core.UnshallowRepo(repoPath)
			if err != nil {
				ShowErrorDialog(err)
			}
			refreshProjectStatus(projectStatus, repoPath)
		}
	case core.GIT_STATUS_REBASE_CONTINUABLE:
		projectStatus.RepoStatus.Text.Text = "Rebase underway, ready to continue"
		projectStatus.RepoStatus.SetColor(theme.ColorNameForeground)
		projectStatus.RepoStatus.SetIcon(theme.WarningIcon())
		projectStatus.FixRepoStatusLink.Text = "continue"
		projectStatus.FixRepoStatusLink.Show()
		projectStatus.FixRepoStatusCallback = func() {
			err := core.FinishRebase(repoPath)
			if err != nil {
				ShowErrorDialog(err)
			}
			refreshProjectStatus(projectStatus, repoPath)
		}
	case core.GIT_STATUS_REBASE_CONFLICTS:
		projectStatus.RepoStatus.Text.Text = "Rebase underway, conflicts detected!"
		projectStatus.RepoStatus.SetColor(theme.ColorNameError)
		projectStatus.RepoStatus.SetIcon(theme.ErrorIcon())
		projectStatus.FixRepoStatusLink.Text = "continue"
		projectStatus.FixRepoStatusLink.Show()
		projectStatus.FixRepoStatusCallback = func() {
			err := core.FinishRebase(repoPath)
			if err != nil {
				ShowErrorDialog(err)
			}
			refreshProjectStatus(projectStatus, repoPath)
		}
	case core.GIT_STATUS_LAST_COMMIT_MERGE:
		projectStatus.RepoStatus.Text.Text = "Merge commit detected! This shouldn't have happened!"
		projectStatus.RepoStatus.SetColor(theme.ColorNameWarning)
		projectStatus.RepoStatus.SetIcon(theme.WarningIcon())
		projectStatus.FixRepoStatusLink.Hide()
	case core.GIT_STATUS_DEATACHED_HEAD:
		projectStatus.RepoStatus.Text.Text = "Deatached head. Currently time travelling!"
		projectStatus.RepoStatus.SetColor(theme.ColorNameWarning)
		projectStatus.RepoStatus.SetIcon(theme.WarningIcon())
		projectStatus.FixRepoStatusLink.Hide()
	}

	ahead, behind, _ := core.GetAheadBehind(repoPath)
	projectStatus.RepoAhead.Text.Text = strconv.Itoa(ahead)
	projectStatus.RepoBehind.Text.Text = strconv.Itoa(behind)

	changes := core.GetWorkingTreeChangeAmount(repoPath)
	if changes == 0 {
		projectStatus.RepoWorkingTree.Hide()
	} else {
		projectStatus.RepoWorkingTree.Show()
		projectStatus.RepoWorkingTree.Text.Text = strconv.Itoa(changes)
	}
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
