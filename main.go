package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type CommitDatum struct {
	Sha        string
	Msg        string
	User       string
	HasContent bool
	HasSource  bool
}

const REPO = "../../Unreal/ScaleGame/"

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("List Data")
	myWindow.Resize(fyne.NewSize(1000, 600))

	start := time.Now()

	//	cmd := exec.Command(`git.exe`, `-P`, `log`, `-10000`, `--pretty=%H||%s||%an`)

	r, err := git.PlainOpen(REPO)
	if err != nil {
		log.Fatal(err)
	}

	commitsIter, _ := r.Log(&git.LogOptions{})

	commitData := make([]*CommitDatum, 0)

	commitsIter.ForEach(func(c *object.Commit) error {
		fmt.Println(c.Hash, c.Message, c.Author.Name)

		newCommit := &CommitDatum{Sha: c.Hash.String(), Msg: strings.Split(c.Message, "\n")[0], User: c.Author.Name}
		if len(c.ParentHashes) > 0 {
			// diffCmd := exec.Command("git.exe", "diff-tree", "--no-commit-id", "--name-only", "-r", commit[0], "--")
			parentHash := c.ParentHashes[0]
			parent, _ := r.CommitObject(parentHash)
			parentTree, _ := parent.Tree()
			commitTree, _ := c.Tree()
			diffTree, _ := commitTree.Diff(parentTree)
			for _, change := range diffTree {
				fmt.Println(change.From.Name, change.To.Name)

				newCommit.HasContent = newCommit.HasContent || strings.Contains(change.To.Name, "Content/") || strings.Contains(change.From.Name, "Content/")
				newCommit.HasSource = newCommit.HasSource || strings.Contains(change.To.Name, "Source/") || strings.Contains(change.From.Name, "Source/")

				if newCommit.HasContent && newCommit.HasSource {
					break
				}
			}
		} else {

			fmt.Println("Initial Commit")
		}
		commitData = append(commitData, newCommit)

		return nil
	})

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
