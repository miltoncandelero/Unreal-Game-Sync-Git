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

type ProjectController struct {
	ProjectStatus *view.ProjectStatus
	CommitList    *view.CommitList
	RepoPath      string
}

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

	project := &ProjectController{RepoPath: repoPath}

	project.ProjectStatus = view.MakeProjectStatus(uprojectPath)
	// stuff that won't change goes here
	project.ProjectStatus.ProjectTitle.Text = strings.ReplaceAll(filepath.Base(uprojectPath), ".uproject", "")
	project.ProjectStatus.Subtitle.Text = uprojectPath
	project.ProjectStatus.RefreshButtonCallback = project.refreshProject
	project.ProjectStatus.ExploreButtonCallback = project.openInExplorer
	project.ProjectStatus.TerminalButtonCallback = project.openInTerminal
	project.ProjectStatus.PullButtonCallback = project.pull
	project.ProjectStatus.SyncButtonCallback = project.sync

	project.ProjectStatus.CommitButtonCallback = project.commit

	project.ProjectStatus.RepoOrigin.SetText(core.GetRepoOrigin(repoPath))
	switch core.GetGitProviderName(repoPath) {
	case "GitHub":
		project.ProjectStatus.RepoOrigin.SetIcon(assets.ResGithubSvg)
	case "GitLab":
		project.ProjectStatus.RepoOrigin.SetIcon(assets.ResGitlabSvg)
	case "Gitea":
		project.ProjectStatus.RepoOrigin.SetIcon(assets.ResGiteaSvg)
	default:
		project.ProjectStatus.RepoOrigin.SetIcon(assets.ResGitSvg)
	}

	project.CommitList = view.MakeCommitList(
		project.checkoutCallback,
		project.resetCallback,
		GetApp().Window.Canvas(),
	)

	project.refreshProject()

	mainVertical := container.NewBorder(project.ProjectStatus, nil, nil, nil, project.CommitList.Container)

	appendProjectToMainWindow(mainVertical, uprojectPath)

}

func (project *ProjectController) checkoutCallback(hash string) {
	defer project.refreshProject()

	if core.GetWorkingTreeChangeAmount(project.RepoPath) > 0 {
		ShowWarningDialog("I'm afraid I can't do that", "You have uncommited changes.\nPlease commit (or discard) them before trying to flashback")
		return
	}

	d := ShowLoadingDialog("Flashing back...")

	err := core.Checkout(project.RepoPath, hash)
	if err != nil {
		d.Hide()
		ShowErrorDialog(err)
		return
	}
	d.Hide()
}

func (project *ProjectController) resetCallback(hash string) {
	defer project.refreshProject()

	if core.GetWorkingTreeChangeAmount(project.RepoPath) > 0 {
		ShowWarningDialog("I'm afraid I can't do that", "You have uncommited changes.\nPlease commit (or discard) them before trying to time travel")
		return
	}

	ahead, _, _ := core.GetAheadBehind(project.RepoPath)
	if ahead > 0 {
		ShowWarningDialog("I'm afraid I can't do that", "You have commits that haven't been pushed. Please push your changes before trying to time travel")
		return
	}

	d := ShowLoadingDialog("Time traveling...")
	// cant defer close if I have dialogs :(

	err := core.ResetHard(project.RepoPath, hash)
	if err != nil {
		d.Hide()
		ShowErrorDialog(err)
		return
	}

	d.Hide()
}

func (project *ProjectController) openInExplorer() {
	open.Run(project.RepoPath)
}

func (project *ProjectController) openInTerminal() {
	core.OpenCmd(project.RepoPath)
}

func (project *ProjectController) pull() {

	defer project.refreshProject()
	d := ShowLoadingDialog("Pulling...")
	if core.GetGitStatus(project.RepoPath) != core.GIT_STATUS_OK {
		d.Hide()
		ShowErrorDialog(fmt.Errorf("Repo not ok. Can't pull"))
		return
	}
	err := core.GitSmartPull(project.RepoPath)
	if err != nil {
		d.Hide()
		ShowErrorDialog(err)
		return
	}

	d.Hide()
}

func (project *ProjectController) sync() {
	defer project.refreshProject()
	d := ShowLoadingDialog("Syncing...")
	if core.GetGitStatus(project.RepoPath) != core.GIT_STATUS_OK {
		d.Hide()
		ShowErrorDialog(fmt.Errorf("Repo not ok. Can't sync"))
		return
	}
	err := core.GitSmartPull(project.RepoPath)
	if err != nil {
		d.Hide()
		ShowErrorDialog(err)
		return
	}
	err = core.GitPush(project.RepoPath)
	if err != nil {
		d.Hide()
		ShowErrorDialog(err)
		return
	}
	d.Hide()
}

func (project *ProjectController) commit() {

	defer project.refreshProject()
	if core.GetGitStatus(project.RepoPath) != core.GIT_STATUS_OK {
		ShowErrorDialog(fmt.Errorf("Repo not ok. Can't commit"))
		return
	}
	dialog.ShowInformation("Not implemented", "Not implemented yet :P", GetApp().Window)

}

func (project *ProjectController) refreshProject() {
	d := ShowLoadingDialog("Refreshing...")
	defer d.Hide()

	project.refreshRepo()
	// refresh build
	// reresh other stuff?
}

func (project *ProjectController) refreshRepo() {
	project.refreshRepoStatus()
	project.refreshRepoUserData()
	project.refreshRepoConfigStatus()
	project.refreshRepoActions()
	project.refreshCommits()
}

func (project *ProjectController) refreshRepoUserData() {
	if core.NeedsUsernameFix(project.RepoPath) {
		project.ProjectStatus.RepoUser.SetText("Username missing!")
		project.ProjectStatus.RepoUser.SetIcon(theme.ErrorIcon())
		project.ProjectStatus.RepoUser.SetColor(theme.ColorNameError)
		project.ProjectStatus.FixUserLink.SetText("Fix")
	} else {
		project.ProjectStatus.RepoUser.SetText(core.GetUsernameFromRepo(project.RepoPath) + " (" + core.GetUserEmailFromRepo(project.RepoPath) + ")")
		project.ProjectStatus.RepoUser.SetIcon(theme.AccountIcon())
		project.ProjectStatus.RepoUser.SetColor(theme.ColorNameForeground)
		project.ProjectStatus.FixUserLink.SetText("Change")
	}
	project.ProjectStatus.FixUserLinkCallback = func() {
		ShowUsernameEmailDialog(core.GetGitProviderName(project.RepoPath),
			func(username string, email string) error {
				err := core.SetUsernameAndEmail(project.RepoPath, username, email)
				if err != nil {
					return err
				}
				project.refreshRepoUserData()
				return nil
			})
	}
}

func (project *ProjectController) refreshRepoConfigStatus() {
	switch core.GetGitConfigStatus(project.RepoPath) {
	case core.CONFIG_STATUS_MISSING:
		project.ProjectStatus.ConfigStatus.SetText(".gitconfig missing")
		project.ProjectStatus.ConfigStatus.SetColor(theme.ColorNameWarning)
		project.ProjectStatus.ConfigStatus.SetIcon(theme.QuestionIcon())
		project.ProjectStatus.FixConfigLink.SetText("Create")
		project.ProjectStatus.FixConfigLink.Show()
		project.ProjectStatus.FixConfigLinkCallback = func() {
			err := core.CreateGitConfig(project.RepoPath)
			if err != nil {
				ShowErrorDialog(err)
			}
			project.refreshRepoConfigStatus()
		}
	case core.CONFIG_STATUS_NOT_LINKED:
		project.ProjectStatus.ConfigStatus.SetText(".gitconfig found but not installed!")
		project.ProjectStatus.ConfigStatus.SetColor(theme.ColorNameError)
		project.ProjectStatus.ConfigStatus.SetIcon(theme.ErrorIcon())
		project.ProjectStatus.FixConfigLink.SetText("Fix")
		project.ProjectStatus.FixConfigLink.Show()
		project.ProjectStatus.FixConfigLinkCallback = func() {
			err := core.LinkGitConfig(project.RepoPath)
			if err != nil {
				ShowErrorDialog(err)
			}
			project.refreshRepoConfigStatus()
		}
	case core.CONFIG_STATUS_LINKED:
		project.ProjectStatus.ConfigStatus.SetText(".gitconfig linked")
		project.ProjectStatus.ConfigStatus.SetColor(theme.ColorNameSuccess)
		project.ProjectStatus.ConfigStatus.SetIcon(theme.ConfirmIcon())
		project.ProjectStatus.FixConfigLink.Hide()
	}
}

func (project *ProjectController) refreshRepoStatus() {
	status := core.GetGitStatus(project.RepoPath)
	switch status {
	case core.GIT_STATUS_OK:
		project.ProjectStatus.RepoStatus.SetText("Repo ok")
		project.ProjectStatus.RepoStatus.SetColor(theme.ColorNameSuccess)
		project.ProjectStatus.RepoStatus.SetIcon(theme.ConfirmIcon())
		project.ProjectStatus.FixRepoStatusLink.Hide()
	case core.GIT_STATUS_SHALLOW:
		project.ProjectStatus.RepoStatus.SetText("Repo is shallow!")
		project.ProjectStatus.RepoStatus.SetColor(theme.ColorNameWarning)
		project.ProjectStatus.RepoStatus.SetIcon(theme.WarningIcon())
		project.ProjectStatus.FixRepoStatusLink.SetText("unshallow")
		project.ProjectStatus.FixRepoStatusLink.Show()
		project.ProjectStatus.FixRepoStatusCallback = func() {
			defer project.refreshRepo()
			d := ShowLoadingDialog("Unshallowing (This will take a while)...")
			err := core.UnshallowRepo(project.RepoPath)
			if err != nil {
				d.Hide()
				ShowErrorDialog(err)
				return
			}
			d.Hide()
		}
	case core.GIT_STATUS_REBASE_CONTINUABLE:
		project.ProjectStatus.RepoStatus.SetText("Rebase underway, ready to continue")
		project.ProjectStatus.RepoStatus.SetColor(theme.ColorNameForeground)
		project.ProjectStatus.RepoStatus.SetIcon(theme.WarningIcon())
		project.ProjectStatus.FixRepoStatusLink.SetText("continue")
		project.ProjectStatus.FixRepoStatusLink.Show()
		project.ProjectStatus.FixRepoStatusCallback = func() {
			defer project.refreshRepo()
			d := ShowLoadingDialog("Rebasing...")
			err := core.FinishRebase(project.RepoPath)
			if err != nil {
				d.Hide()
				ShowErrorDialog(err)
				return
			}
			d.Hide()
		}
	case core.GIT_STATUS_REBASE_CONFLICTS:
		project.ProjectStatus.RepoStatus.SetText("Rebase underway, conflicts detected!")
		project.ProjectStatus.RepoStatus.SetColor(theme.ColorNameError)
		project.ProjectStatus.RepoStatus.SetIcon(theme.ErrorIcon())
		project.ProjectStatus.FixRepoStatusLink.SetText("continue")
		project.ProjectStatus.FixRepoStatusLink.Show()
		project.ProjectStatus.FixRepoStatusCallback = func() {
			defer project.refreshRepo()
			d := ShowLoadingDialog("Rebasing...")
			err := core.FinishRebase(project.RepoPath)
			if err != nil {
				d.Hide()
				ShowErrorDialog(err)
				return
			}
			d.Hide()
		}
	case core.GIT_STATUS_LAST_COMMIT_MERGE:
		project.ProjectStatus.RepoStatus.SetText("Merge commit detected! This shouldn't have happened!")
		project.ProjectStatus.RepoStatus.SetColor(theme.ColorNameWarning)
		project.ProjectStatus.RepoStatus.SetIcon(theme.WarningIcon())
		project.ProjectStatus.FixRepoStatusLink.Hide()
	case core.GIT_STATUS_DEATACHED_HEAD:
		project.ProjectStatus.RepoStatus.SetText("Currently in a Flashback (Deatached HEAD)")
		project.ProjectStatus.RepoStatus.SetColor(theme.ColorNameWarning)
		project.ProjectStatus.RepoStatus.SetIcon(theme.WarningIcon())
		project.ProjectStatus.FixRepoStatusLink.SetText("end Flashback")
		project.ProjectStatus.FixRepoStatusCallback = func() {
			defer project.refreshRepo()
			d := ShowLoadingDialog("Returning...")
			err := core.ReturnToLastBranch(project.RepoPath)
			if err != nil {
				d.Hide()
				ShowErrorDialog(err)
				return
			}
			d.Hide()
		}
		project.ProjectStatus.FixRepoStatusLink.Show()
	}

	if core.IsDeatachedHead(project.RepoPath) {
		project.ProjectStatus.RepoBranch.SetText("in a Flashback")
		project.ProjectStatus.RepoBranch.SetIcon(theme.WarningIcon())
		project.ProjectStatus.RepoAhead.Hide()
		project.ProjectStatus.RepoBehind.Hide()
		project.ProjectStatus.RepoWorkingTree.Hide()
	} else {

		ahead, behind, _ := core.GetAheadBehind(project.RepoPath)
		project.ProjectStatus.RepoAhead.SetText(strconv.Itoa(ahead))
		project.ProjectStatus.RepoBehind.SetText(strconv.Itoa(behind))
		project.ProjectStatus.RepoAhead.Show()
		project.ProjectStatus.RepoBehind.Show()

		changes := core.GetWorkingTreeChangeAmount(project.RepoPath)
		if changes == 0 {
			project.ProjectStatus.RepoWorkingTree.Hide()
		} else {
			project.ProjectStatus.RepoWorkingTree.Show()
			project.ProjectStatus.RepoWorkingTree.SetText(strconv.Itoa(changes))
		}

		branch, _ := core.GetCurrentBranchFromRepository(project.RepoPath)

		project.ProjectStatus.RepoBranch.SetText(branch)
		project.ProjectStatus.RepoBranch.SetIcon(assets.ResBranchSvg)
	}

}

func (project *ProjectController) refreshRepoActions() {
	if core.GetGitStatus(project.RepoPath) != core.GIT_STATUS_OK {
		project.ProjectStatus.PullButton.Disable()
		project.ProjectStatus.SyncButton.Disable()
		project.ProjectStatus.CommitButton.Disable()

		project.ProjectStatus.PullButton.SetText("Repo not ok. Can't pull")
		project.ProjectStatus.SyncButton.SetText("Repo not ok. Can't sync")
		project.ProjectStatus.CommitButton.SetText("Repo not ok. Can't commit")
		return
	}

	if core.GetWorkingTreeChangeAmount(project.RepoPath) > 0 {
		project.ProjectStatus.CommitButton.SetText("Commit")
		project.ProjectStatus.CommitButton.Enable()
	} else {
		project.ProjectStatus.CommitButton.SetText("Nothing to commit")
		project.ProjectStatus.CommitButton.Disable()
	}

	ahead, behind, _ := core.GetAheadBehind(project.RepoPath)
	if behind > 0 {
		project.ProjectStatus.PullButton.SetText("Pull")
		project.ProjectStatus.PullButton.Enable()
	} else {
		project.ProjectStatus.PullButton.SetText("Nothing to pull")
		project.ProjectStatus.PullButton.Disable()
	}

	if behind > 0 || ahead > 0 {
		project.ProjectStatus.SyncButton.SetText("Sync")
		project.ProjectStatus.SyncButton.Enable()
	} else {
		project.ProjectStatus.SyncButton.SetText("Nothing to sync")
		project.ProjectStatus.SyncButton.Disable()
	}

}

func (project *ProjectController) refreshCommits() {
	commits, _ := core.GetRepoBranchInfo(project.RepoPath, "")
	project.CommitList.UpdateCommits(commits)
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
