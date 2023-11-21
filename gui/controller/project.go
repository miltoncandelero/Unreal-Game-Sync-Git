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
	// stuff that won't change goes here
	projectStatus.ProjectTitle.Text = strings.ReplaceAll(filepath.Base(uprojectPath), ".uproject", "")
	projectStatus.Subtitle.Text = uprojectPath
	projectStatus.RefreshButtonCallback = func() {
		refreshProject(projectStatus, repoPath)
	}

	projectStatus.PullButtonCallback = func() {
		defer refreshProject(projectStatus, repoPath)
		if core.GetGitStatus(repoPath) != core.GIT_STATUS_OK {
			ShowErrorDialog(fmt.Errorf("Repo not ok. Can't pull"))
			return
		}
		err := core.GitSmartPull(repoPath)
		if err != nil {
			ShowErrorDialog(err)
		}
	}
	projectStatus.SyncButtonCallback = func() {
		defer refreshProject(projectStatus, repoPath)
		if core.GetGitStatus(repoPath) != core.GIT_STATUS_OK {
			ShowErrorDialog(fmt.Errorf("Repo not ok. Can't sync"))
			return
		}
		err := core.GitSmartPull(repoPath)
		if err != nil {
			ShowErrorDialog(err)
			return
		}
		err = core.GitPush(repoPath)
		if err != nil {
			ShowErrorDialog(err)
			return
		}
	}

	projectStatus.CommitButtonCallback = func() {
		defer refreshProject(projectStatus, repoPath)
		if core.GetGitStatus(repoPath) != core.GIT_STATUS_OK {
			ShowErrorDialog(fmt.Errorf("Repo not ok. Can't commit"))
			return
		}
		dialog.ShowInformation("Not implemented", "Not implemented yet :P", GetApp().Window)
	}

	projectStatus.RepoOrigin.SetText(core.GetRepoOrigin(repoPath))
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

	appendProjectToMainWindow(projectStatus, uprojectPath)

	refreshProject(projectStatus, repoPath)
}

func refreshProject(projectStatus *view.ProjectStatus, repoPath string) {
	projectStatus.RefreshButton.Disable()
	defer projectStatus.RefreshButton.Enable()

	refreshRepo(projectStatus, repoPath)
}

func refreshRepo(projectStatus *view.ProjectStatus, repoPath string) {
	refreshRepoStatus(projectStatus, repoPath)
	refreshRepoUserData(projectStatus, repoPath)
	refreshRepoConfigStatus(projectStatus, repoPath)
	refreshRepoActions(projectStatus, repoPath)
}

func refreshRepoUserData(projectStatus *view.ProjectStatus, repoPath string) {
	if core.NeedsUsernameFix(repoPath) {
		projectStatus.RepoUser.SetText("Username missing!")
		projectStatus.RepoUser.SetIcon(theme.ErrorIcon())
		projectStatus.RepoUser.SetColor(theme.ColorNameError)
		projectStatus.FixUserLink.Text = "Fix"
	} else {
		projectStatus.RepoUser.SetText(core.GetUsernameFromRepo(repoPath) + " (" + core.GetUserEmailFromRepo(repoPath) + ")")
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
				refreshRepoUserData(projectStatus, repoPath)
				return nil
			})
	}
}

func refreshRepoConfigStatus(projectStatus *view.ProjectStatus, repoPath string) {
	switch core.GetGitConfigStatus(repoPath) {
	case core.CONFIG_STATUS_MISSING:
		projectStatus.ConfigStatus.SetText(".gitconfig missing")
		projectStatus.ConfigStatus.SetColor(theme.ColorNameWarning)
		projectStatus.ConfigStatus.SetIcon(theme.QuestionIcon())
		projectStatus.FixConfigLink.Text = "Create"
		projectStatus.FixConfigLink.Show()
		projectStatus.FixConfigLinkCallback = func() {
			err := core.CreateGitConfig(repoPath)
			if err != nil {
				ShowErrorDialog(err)
			}
			refreshRepoConfigStatus(projectStatus, repoPath)
		}
	case core.CONFIG_STATUS_NOT_LINKED:
		projectStatus.ConfigStatus.SetText(".gitconfig found but not installed!")
		projectStatus.ConfigStatus.SetColor(theme.ColorNameError)
		projectStatus.ConfigStatus.SetIcon(theme.ErrorIcon())
		projectStatus.FixConfigLink.Text = "Fix"
		projectStatus.FixConfigLink.Show()
		projectStatus.FixConfigLinkCallback = func() {
			err := core.LinkGitConfig(repoPath)
			if err != nil {
				ShowErrorDialog(err)
			}
			refreshRepoConfigStatus(projectStatus, repoPath)
		}
	case core.CONFIG_STATUS_LINKED:
		projectStatus.ConfigStatus.SetText(".gitconfig linked")
		projectStatus.ConfigStatus.SetColor(theme.ColorNameSuccess)
		projectStatus.ConfigStatus.SetIcon(theme.ConfirmIcon())
		projectStatus.FixConfigLink.Hide()
	}
}

func refreshRepoStatus(projectStatus *view.ProjectStatus, repoPath string) {
	status := core.GetGitStatus(repoPath)
	switch status {
	case core.GIT_STATUS_OK:
		projectStatus.RepoStatus.SetText("Repo ok")
		projectStatus.RepoStatus.SetColor(theme.ColorNameSuccess)
		projectStatus.RepoStatus.SetIcon(theme.ConfirmIcon())
		projectStatus.FixRepoStatusLink.Hide()
	case core.GIT_STATUS_SHALLOW:
		projectStatus.RepoStatus.SetText("Repo is shallow!")
		projectStatus.RepoStatus.SetColor(theme.ColorNameWarning)
		projectStatus.RepoStatus.SetIcon(theme.WarningIcon())
		projectStatus.FixRepoStatusLink.Text = "unshallow"
		projectStatus.FixRepoStatusLink.Show()
		projectStatus.FixRepoStatusCallback = func() {
			err := core.UnshallowRepo(repoPath)
			if err != nil {
				ShowErrorDialog(err)
			}
			refreshRepo(projectStatus, repoPath)
		}
	case core.GIT_STATUS_REBASE_CONTINUABLE:
		projectStatus.RepoStatus.SetText("Rebase underway, ready to continue")
		projectStatus.RepoStatus.SetColor(theme.ColorNameForeground)
		projectStatus.RepoStatus.SetIcon(theme.WarningIcon())
		projectStatus.FixRepoStatusLink.Text = "continue"
		projectStatus.FixRepoStatusLink.Show()
		projectStatus.FixRepoStatusCallback = func() {
			err := core.FinishRebase(repoPath)
			if err != nil {
				ShowErrorDialog(err)
			}
			refreshRepo(projectStatus, repoPath)
		}
	case core.GIT_STATUS_REBASE_CONFLICTS:
		projectStatus.RepoStatus.SetText("Rebase underway, conflicts detected!")
		projectStatus.RepoStatus.SetColor(theme.ColorNameError)
		projectStatus.RepoStatus.SetIcon(theme.ErrorIcon())
		projectStatus.FixRepoStatusLink.Text = "continue"
		projectStatus.FixRepoStatusLink.Show()
		projectStatus.FixRepoStatusCallback = func() {
			err := core.FinishRebase(repoPath)
			if err != nil {
				ShowErrorDialog(err)
			}
			refreshRepo(projectStatus, repoPath)
		}
	case core.GIT_STATUS_LAST_COMMIT_MERGE:
		projectStatus.RepoStatus.SetText("Merge commit detected! This shouldn't have happened!")
		projectStatus.RepoStatus.SetColor(theme.ColorNameWarning)
		projectStatus.RepoStatus.SetIcon(theme.WarningIcon())
		projectStatus.FixRepoStatusLink.Hide()
	case core.GIT_STATUS_DEATACHED_HEAD:
		projectStatus.RepoStatus.SetText("Currently in a Flashback (Deatached HEAD)")
		projectStatus.RepoStatus.SetColor(theme.ColorNameWarning)
		projectStatus.RepoStatus.SetIcon(theme.WarningIcon())
		projectStatus.FixRepoStatusLink.Text = "end Flashback"
		projectStatus.FixRepoStatusCallback = func() {
			err := core.ReturnToLastBranch(repoPath)
			if err != nil {
				ShowErrorDialog(err)
			}
			refreshRepo(projectStatus, repoPath)
		}
		projectStatus.FixRepoStatusLink.Show()
	}

	if core.IsDeatachedHead(repoPath) {
		projectStatus.RepoBranch.SetText("in a Flashback")
		projectStatus.RepoBranch.SetIcon(theme.WarningIcon())
		projectStatus.RepoAhead.Hide()
		projectStatus.RepoBehind.Hide()
		projectStatus.RepoWorkingTree.Hide()
	} else {

		ahead, behind, _ := core.GetAheadBehind(repoPath)
		projectStatus.RepoAhead.SetText(strconv.Itoa(ahead))
		projectStatus.RepoBehind.SetText(strconv.Itoa(behind))
		projectStatus.RepoAhead.Show()
		projectStatus.RepoBehind.Show()

		changes := core.GetWorkingTreeChangeAmount(repoPath)
		if changes == 0 {
			projectStatus.RepoWorkingTree.Hide()
		} else {
			projectStatus.RepoWorkingTree.Show()
			projectStatus.RepoWorkingTree.SetText(strconv.Itoa(changes))
		}

		branch, _ := core.GetCurrentBranchFromRepository(repoPath)

		projectStatus.RepoBranch.SetText(branch)
		projectStatus.RepoBranch.SetIcon(assets.ResBranchSvg)
	}

}

func refreshRepoActions(projectStatus *view.ProjectStatus, repoPath string) {
	if core.GetGitStatus(repoPath) != core.GIT_STATUS_OK {
		projectStatus.PullButton.Disable()
		projectStatus.SyncButton.Disable()
		projectStatus.CommitButton.Disable()

		projectStatus.PullButton.Text = "Repo not ok. Can't pull"
		projectStatus.SyncButton.Text = "Repo not ok. Can't sync"
		projectStatus.CommitButton.Text = "Repo not ok. Can't commit"
		return
	}

	if core.GetWorkingTreeChangeAmount(repoPath) > 0 {
		projectStatus.CommitButton.Text = "Commit"
		projectStatus.CommitButton.Enable()
	} else {
		projectStatus.CommitButton.Text = "Nothing to commit"
		projectStatus.CommitButton.Disable()
	}

	ahead, behind, _ := core.GetAheadBehind(repoPath)
	if behind > 0 {
		projectStatus.PullButton.Text = "Pull"
		projectStatus.PullButton.Enable()
	} else {
		projectStatus.PullButton.Text = "Nothing to pull"
		projectStatus.PullButton.Disable()
	}

	if behind > 0 || ahead > 0 {
		projectStatus.SyncButton.Text = "Sync"
		projectStatus.SyncButton.Enable()
	} else {
		projectStatus.SyncButton.Text = "Nothing to sync"
		projectStatus.SyncButton.Disable()
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
