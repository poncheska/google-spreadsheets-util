package application

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/poncheska/google-spreadsheets-util/pkg/utils"
	"log"
	"strings"
)

var (
	execChan = make(chan int)
	isExec   = false
)

type View struct {
	fieldBox        *fyne.Container
	urlEntry        *widget.Entry
	configPathEntry *widget.Entry
	sheetEntry      *widget.Entry
	fieldEntries    []*widget.Entry
	contentEntries  []*widget.Entry
}

func Run() {
	a := app.New()
	w := a.NewWindow("Google spreadsheets util")

	view := View{
		fieldBox:        container.New(layout.NewGridLayout(2)),
		urlEntry:        widget.NewEntry(),
		configPathEntry: widget.NewEntry(),
		sheetEntry:      widget.NewEntry(),
		fieldEntries:    []*widget.Entry{},
		contentEntries:  []*widget.Entry{},
	}

	view.configPathEntry.Text = "credential.json"

	prBar := widget.NewProgressBarInfinite()
	prBar.Stop()

	stopButton := widget.NewButton("STOP", func() {
		if isExec {
			execChan <- 0
			isExec = false
		}
	})
	stopButton.Hide()

	execButton := widget.NewButton("EXECUTE", func() {
		if isExec {
			return
		}
		var flds []utils.FieldData
		if len(view.contentEntries) != len(view.fieldEntries) {
			return
		}
		l := len(view.fieldEntries)
		for i := 0; i < l; i++ {
			flds = append(flds, utils.FieldData{
				Field:   strings.Trim(view.fieldEntries[i].Text, " "),
				Content: strings.Trim(view.contentEntries[i].Text, " "),
			})
		}
		data := &utils.Data{
			URL:        strings.Trim(view.urlEntry.Text, " "),
			ConfigPath: strings.Trim(view.configPathEntry.Text, " "),
			SheetName:  strings.Trim(view.sheetEntry.Text, " "),
			Fields:     flds,
		}
		log.Println(data)
		pdata, err := data.ValidateAndParse()
		if err != nil {
			log.Println(err.Error())
			return
		}
		log.Println(pdata)
		prBar.Start()
		isExec = true
		stopButton.Show()
		go func() {
			pdata.DoReq(execChan)
			stopButton.Hide()
			w.Resize(fyne.NewSize(300, 300))
			prBar.Stop()
			prBar.Refresh()
		}()
	})

	w.SetContent(container.NewVBox(
		widget.NewLabel("GOOGLE SPREADSHEETS UTIL"),
		container.New(layout.NewFormLayout(),
			widget.NewLabel("Spreadsheets URL:"),
			view.urlEntry,
		),
		container.New(layout.NewFormLayout(),
			widget.NewLabel("Config path:"),
			view.configPathEntry,
		),
		container.New(layout.NewFormLayout(),
			widget.NewLabel("Sheet name:"),
			view.sheetEntry,
		),
		container.New(layout.NewGridLayout(2),
			widget.NewLabel("Field:"),
			widget.NewLabel("Content:"),
		),
		view.fieldBox,
		container.New(layout.NewGridLayout(2),
			widget.NewButton("ADD", func() {
				view.fieldEntries = append(view.fieldEntries, widget.NewEntry())
				view.contentEntries = append(view.contentEntries, widget.NewEntry())
				view.fieldBox.Add(view.fieldEntries[len(view.fieldEntries)-1])
				view.fieldBox.Add(view.contentEntries[len(view.contentEntries)-1])
			}),
			widget.NewButton("DELETE", func() {
				l := len(view.fieldBox.Objects)
				if l != 0 {
					view.fieldBox.Objects = view.fieldBox.Objects[:l-2]
					view.fieldEntries = view.fieldEntries[:len(view.fieldEntries)-1]
					view.contentEntries = view.contentEntries[:len(view.contentEntries)-1]
					w.Resize(fyne.NewSize(300, 300))
				}
			}),
		),
		prBar,
		execButton,
		stopButton,
	))

	w.Resize(fyne.NewSize(300, 300))
	w.SetFixedSize(true)
	w.ShowAndRun()
}
