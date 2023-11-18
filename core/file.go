package core

import "os"

func FileExists(file string) bool {
	_, err := os.Stat(file)
	return !os.IsNotExist(err)
}

func GetConfigString() string {
	return `[remote "origin"]
	tagOpt = --no-tags
	prune = true
	pruneTags = true
[core]
	whitespace = cr-at-eol,trailing-space,space-before-tab,indent-with-non-tab,tabwidth=4
	preloadIndex = true
	multiPackIndex = true
	splitIndex = true
	commitGraph = true
	autocrlf = false
	safecrlf = false
	fscache = true
	checkStat = minimal
	trustctime = false
	untrackedCache = true
	useBuiltinFSMonitor = false
	fsmonitor = false
	fsyncMethod = batch
[gc]
	auto = 0
	writeCommitGraph = false
	pruneExpire = 1.weeks.ago
[rerere]
	enabled = true
	autoupdate = false
[lfs]
	locksverify = true
	skipdownloaderrors = true
	batch = true
	setlockablereadonly = true
	tustransfers = true
	pruneverifyremotealways = true
	fetchrecentrefsdays = 1
	pruneoffsetdays = 0
	fetchrecentremoterefs = false
[filter "lfs"]
	smudge = git-lfs smudge -- %f
	process = git-lfs filter-process
	required = true
	clean = git-lfs clean -- %f
[index]
	threads = true
	version = 4
[fetch]
	prune = true
	pruneTags = true
	negotiationAlgorithm = skipping
	unpackLimit = 1
	writeCommitGraph = false
	showForcedUpdates = false
[help]
	autoCorrect = prompt
[pack]
	threads = 0
[push]
	autoSetupRemote = true
[pull]
	rebase = true
[receive]
	autogc = false
[rebase]
	autostash = true
[checkout]
	optimizenewbranch = true
[am]
	keepcr = true
[reset]
	quiet = true
[status]
	aheadbehind = false
[gui]
	GCWarning = false
[diff]
	algorithm = histogram
[maintenance]
	auto = false
[advice]
	fetchShowForcedUpdates = false
[log]
	excludedecoration = refs/prefetch/
[fastimport]
	unpackLimit = 1`
}
