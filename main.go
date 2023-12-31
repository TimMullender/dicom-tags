package main

import (
	"encoding/csv"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/TimMullender/dicom-tags/cmd"
	"github.com/suyashkumar/dicom/pkg/tag"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		cmd.Execute()
	} else {
		a := app.New()
		w := a.NewWindow("DICOM Tags")
		w.Resize(fyne.NewSize(480, 320))
		tags := make([]tag.Info, 3)
		tags[0] = tag.MustFind(tag.Tag{Group: 0x0020, Element: 0x000D})
		tags[1] = tag.MustFind(tag.Tag{Group: 0x0020, Element: 0x000E})
		tags[2] = tag.MustFind(tag.Tag{Group: 0x0008, Element: 0x0018})
		var records [][]string
		folder := widget.NewEntry()
		folderOpen := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if uri != nil {
				folder.SetText(uri.Path())
			}
		}, w)
		fileSave := dialog.NewFileSave(func(closer fyne.URIWriteCloser, err error) {
			if closer != nil {
				csvWriter := csv.NewWriter(closer)
				//_ = csvWriter.Write(headers)
				_ = csvWriter.WriteAll(records)
				csvWriter.Flush()
			}
		}, w)
		table := widget.NewTableWithHeaders(
			func() (rows int, cols int) {
				return len(records), len(tags) + 2
			},
			func() fyne.CanvasObject {
				return container.NewScroll(widget.NewLabel(folder.Text))
			},
			func(id widget.TableCellID, object fyne.CanvasObject) {
				if id.Col < len(records[id.Row]) {
					object.(*container.Scroll).Content.(*widget.Label).SetText(records[id.Row][id.Col])
				} else {
					object.(*container.Scroll).Content.(*widget.Label).SetText("")
				}
			},
		)
		selectTag := widget.NewSelectEntry(cmd.TagNames)
		addTagDialog := dialog.NewForm("Add Tag", "Add", "Cancel", []*widget.FormItem{widget.NewFormItem("Tag", selectTag)}, func(add bool) {
			if add {
				tagName := selectTag.Text
				info, err := tag.FindByName(tagName)
				_, _ = fmt.Fprintf(os.Stdout, "Adding tag (%s): %v\n", tagName, err)
				if err == nil {
					tags = append(tags, info)
					table.Refresh()
				}
			}
		}, w)
		addTag := widget.NewButton("Add Tag", func() {
			addTagDialog.Show()
		})
		search := widget.NewButton("Search", func() {
			result, err := cmd.WalkDirectory(folder.Text, tags, nil)
			if err == nil {
				records = result
			}
		})
		search.Disable()
		folder.OnChanged = func(s string) {
			if len(s) < 1 || len(tags) < 1 {
				search.Disable()
			} else {
				search.Enable()
			}
		}
		save := widget.NewButton("Save", func() {
			fileSave.Show()
		})
		table.CreateHeader = func() fyne.CanvasObject {
			return widget.NewLabel("")
		}
		table.UpdateHeader = func(id widget.TableCellID, template fyne.CanvasObject) {
			if id.Col == 0 {
				template.(*widget.Label).SetText("Filename")
			} else if id.Col > 0 && id.Col <= len(tags) {
				template.(*widget.Label).SetText(tags[id.Col-1].Name)
			} else {
				template.(*widget.Label).SetText("")
			}
		}
		for idx := 0; idx <= len(tags); idx++ {
			table.SetColumnWidth(idx, 150)
		}
		content := container.NewVBox(
			container.New(layout.NewFormLayout(),
				widget.NewButton("Select Folder", func() {
					folderOpen.Show()
				}),
				folder,
			),
			container.NewCenter(container.NewHBox(search, addTag, save)),
			table,
		)
		w.SetContent(content)
		w.ShowAndRun()
	}
}
