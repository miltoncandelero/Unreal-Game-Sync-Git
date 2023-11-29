# UGS: G - Unreal Game Sync for Git

## Overview

UGS: G, short for Unreal Game Sync: Git, is a tool inspired by [Unreal Game Sync](https://docs.unrealengine.com/5.3/en-US/unreal-game-sync-ugs-for-unreal-engine/) but tailored for Git instead of Perforce. Initially developed to aid artists and game designers at [Killabunnies](https://killabunnies.com.ar), this tool is designed to keep Unreal projects in sync without relying on Visual Studio for compilation. UGS: G is now released under the MIT license, potentially benefiting a broader audience.

This project does not intend to replace your entire Git client. It is recommended to use a Git client, with personal preferences leaning towards [Fork](https://git-fork.com/) or [GitFiend](https://gitfiend.com/).

## Acknowledgments

UGS: G owes its existence to the incredible individuals at [Project Borealis](https://github.com/ProjectBorealis) and their tool [PBSync](https://github.com/ProjectBorealis/PBSync).

## Features

UGS: G is a work in progress, and you can track its development through the [project roadmap board](https://github.com/miltoncandelero/ugsg/projects).

### Git

- [x] Detect and attempt to fix broken states
- [x] Configure username and Git settings
- [x] Push
- [x] Pull
- [x] Timetravel (`checkout` and `reset --hard`)
- [ ] Commit

### Build System

- [ ] Build / Upload / Download Binaries
- [x] Detect commits requiring Binaries

### Engine Tools

- [ ] Change engine version
- [ ] Generate Solution
- [ ] Build / Upload / Download Engine Build
- [ ] Detect and update Git plugin

### Base Tool

- [ ] Autoupdater
- [ ] Dependency checker/downloader (Git, LFS, credential manager, etc.)
- [x] Open project folder
- [x] Open project in console
- [ ] Per project settings

## How to Build

UGS: G is implemented in Go to facilitate compilation across multiple platforms and leverage go-git when possible. The UI is built with the fyne UI toolkit.

To compile UGS: G:

1. Set up the fyne UI toolkit.
2. Proceed with the compilation process.

## License

UGS: G is released under the MIT License.
