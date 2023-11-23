package controller

import (
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"github.com/miltoncandelero/ugsg/core"
	"github.com/miltoncandelero/ugsg/gui/assets"
	"github.com/miltoncandelero/ugsg/gui/view"
	"github.com/ncruces/zenity"
	"github.com/skratchdot/open-golang/open"
)

func UProjectOpened(uprojectPath string) {
	d := ShowLoadingDialog("Opening...")
	defer d.Hide()

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
		d := ShowLoadingDialog("Refreshing...")
		defer d.Hide()
		refreshProject(projectStatus, repoPath)
	}
	projectStatus.ExploreButtonCallback = func() {
		open.Start(repoPath)
	}
	projectStatus.TerminalButtonCallback = func() {
		core.OpenCmd(repoPath)
	}

	projectStatus.PullButtonCallback = func() {
		d := ShowLoadingDialog("Pulling...")
		defer d.Hide()
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
		d := ShowLoadingDialog("Syncing...")
		defer d.Hide()
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

	refreshProject(projectStatus, repoPath)

	commits, _ := core.GetRepoBranchInfo(repoPath, "")
	commitList := view.MakeCommitList(commits)

	mainVertical := container.NewBorder(projectStatus, nil, nil, nil, commitList)

	appendProjectToMainWindow(mainVertical, uprojectPath)

}

func refreshProject(projectStatus *view.ProjectStatus, repoPath string) {
	refreshRepo(projectStatus, repoPath)
	// refresh build
	// reresh other stuff?
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
		projectStatus.FixUserLink.SetText("Fix")
	} else {
		projectStatus.RepoUser.SetText(core.GetUsernameFromRepo(repoPath) + " (" + core.GetUserEmailFromRepo(repoPath) + ")")
		projectStatus.RepoUser.SetIcon(theme.AccountIcon())
		projectStatus.RepoUser.SetColor(theme.ColorNameForeground)
		projectStatus.FixUserLink.SetText("Change")
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
		projectStatus.FixConfigLink.SetText("Create")
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
		projectStatus.FixConfigLink.SetText("Fix")
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
		projectStatus.FixRepoStatusLink.SetText("unshallow")
		projectStatus.FixRepoStatusLink.Show()
		projectStatus.FixRepoStatusCallback = func() {
			d := ShowLoadingDialog("Unshallowing (This will take a while)...")
			defer d.Hide()
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
		projectStatus.FixRepoStatusLink.SetText("continue")
		projectStatus.FixRepoStatusLink.Show()
		projectStatus.FixRepoStatusCallback = func() {
			d := ShowLoadingDialog("Rebasing...")
			defer d.Hide()
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
		projectStatus.FixRepoStatusLink.SetText("continue")
		projectStatus.FixRepoStatusLink.Show()
		projectStatus.FixRepoStatusCallback = func() {
			d := ShowLoadingDialog("Rebasing...")
			defer d.Hide()
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
		projectStatus.FixRepoStatusLink.SetText("end Flashback")
		projectStatus.FixRepoStatusCallback = func() {
			d := ShowLoadingDialog("Returning...")
			defer d.Hide()
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

		projectStatus.PullButton.SetText("Repo not ok. Can't pull")
		projectStatus.SyncButton.SetText("Repo not ok. Can't sync")
		projectStatus.CommitButton.SetText("Repo not ok. Can't commit")
		return
	}

	if core.GetWorkingTreeChangeAmount(repoPath) > 0 {
		projectStatus.CommitButton.SetText("Commit")
		projectStatus.CommitButton.Enable()
	} else {
		projectStatus.CommitButton.SetText("Nothing to commit")
		projectStatus.CommitButton.Disable()
	}

	ahead, behind, _ := core.GetAheadBehind(repoPath)
	if behind > 0 {
		projectStatus.PullButton.SetText("Pull")
		projectStatus.PullButton.Enable()
	} else {
		projectStatus.PullButton.SetText("Nothing to pull")
		projectStatus.PullButton.Disable()
	}

	if behind > 0 || ahead > 0 {
		projectStatus.SyncButton.SetText("Sync")
		projectStatus.SyncButton.Enable()
	} else {
		projectStatus.SyncButton.SetText("Nothing to sync")
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
