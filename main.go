package main

import (
	"log"
	"os/exec"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type CommitDatum struct {
	Sha        string
	Msg        string
	User       string
	HasContent bool
	HasSource  bool
}

const REPO = "../../Unreal/DemoRepo"

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("List Data")
	myWindow.Resize(fyne.NewSize(1000, 600))

	start := time.Now()

	cmd := exec.Command(`git.exe`, `-P`, `log`, `-10000`, `--pretty=%H||%s||%an`)
	cmd.Dir = REPO
	out, err := cmd.CombinedOutput()

	if err != nil {
		log.Fatal(err, string(out))
	}

	srtOut := string(out)
	// split the string into lines
	lines := strings.Split(srtOut, "\n")
	// split the lines by the delimiter
	// and create a slice of struct
	commitData := make([]*CommitDatum, 0, len(lines))
	for _, line := range lines {
		if line != "" {
			commit := strings.Split(line, "||")
			//`git diff-tree --no-commit-id --name-only -r <commit>`
			diffCmd := exec.Command("git.exe", "diff-tree", "--no-commit-id", "--name-only", "-r", commit[0], "--")
			diffCmd.Dir = REPO
			diffOut, _ := diffCmd.CombinedOutput()
			diffStr := string(diffOut)
			newCommit := &CommitDatum{Sha: commit[0], Msg: commit[1], User: commit[2]}
			newCommit.HasContent = strings.Contains(diffStr, "Content/")
			newCommit.HasSource = strings.Contains(diffStr, "Source/")
			commitData = append(commitData, newCommit)
		}
	}

	elapsed := time.Since(start)
	log.Printf("Git took %s", elapsed)

	list := widget.NewTableWithHeaders(
		func() (int, int) {
			return len(commitData), 4
		},
		func() fyne.CanvasObject {
			x := widget.NewLabel("wide content")
			x.Truncation = fyne.TextTruncateEllipsis
			return x
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			switch i.Col {
			case 0:
				o.(*widget.Label).SetText(commitData[i.Row].Sha)
			case 1:
				o.(*widget.Label).SetText(commitData[i.Row].Msg)
			case 2:
				o.(*widget.Label).SetText(commitData[i.Row].User)
			case 3:
				if commitData[i.Row].HasContent && commitData[i.Row].HasSource {
					o.(*widget.Label).SetText("Both")
				} else if commitData[i.Row].HasContent {
					o.(*widget.Label).SetText("Content")
				} else if commitData[i.Row].HasSource {
					o.(*widget.Label).SetText("Source")
				} else {
					o.(*widget.Label).SetText("None")
				}
			}
		})
	list.ShowHeaderColumn = false

	list.UpdateHeader = func(id widget.TableCellID, template fyne.CanvasObject) {
		switch id.Col {
		case 0:
			template.(*widget.Label).SetText("Sha")
		case 1:
			template.(*widget.Label).SetText("Message")
		case 2:
			template.(*widget.Label).SetText("User")
		case 3:
			template.(*widget.Label).SetText("Commit type")
		}
	}

	list.OnSelected = func(id widget.TableCellID) {
		log.Println("selected", id)
		//list.Select(widget.TableCellID{Row: -1, Col: -1})
	}

	list.SetColumnWidth(0, 500)
	list.SetColumnWidth(1, 500)
	list.SetColumnWidth(2, 500)
	list.SetColumnWidth(3, 500)

	add := widget.NewButton("Append", func() {
		val := &CommitDatum{Sha: "44444ddddd", Msg: "Yet another commit", User: "User 4"}
		commitData = append(commitData, val)
		//data.Append(val)
	})

	mangle := widget.NewButton("mangle", func() {
		// interfaceList, _ := data.Get()
		// interfaceList[0].(*CommitDatum).Sha = "mangled"
	})

	buttons := container.New(layout.NewVBoxLayout(), add, mangle)

	myWindow.SetContent(container.NewBorder(nil, buttons, nil, nil, list))
	myWindow.ShowAndRun()
}
