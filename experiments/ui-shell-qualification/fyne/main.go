package main

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.NewWithID("eco.ui.shell.fyne.qualification")
	w := a.NewWindow("ECO UI shell qualification — Fyne")
	w.Resize(fyne.NewSize(1000, 700))

	title := widget.NewLabel("Synthetic Matter Workspace")
	searchLabel := widget.NewLabel("Search this Matter")
	search := widget.NewEntry()
	search.SetPlaceHolder("Search this Matter")

	warranty := widget.NewButton("Warranty confirmation", func() {})
	timeline := widget.NewButton("Build the Matter timeline", func() {})
	addEvidence := widget.NewButton("Add evidence", func() {})
	reviewSource := widget.NewButton("Review source details", func() {})
	createTask := widget.NewButton("Create task", func() {})
	askEco := widget.NewButton("Ask ECO", func() {})

	transcriptTitle := widget.NewLabel("AI conversation transcript")
	transcript := widget.NewMultiLineEntry()
	transcript.SetText("Known: warranty confirmation appears in preserved source Email.eml. Source-backed synthetic qualification text only.")
	transcript.Wrapping = fyne.TextWrapWord

	search.OnChanged = func(value string) {
		q := strings.ToLower(strings.TrimSpace(value))
		if q == "" {
			warranty.Show()
			timeline.Show()
			return
		}
		if strings.Contains(strings.ToLower(warranty.Text), q) {
			warranty.Show()
		} else {
			warranty.Hide()
		}
		if strings.Contains(strings.ToLower(timeline.Text), q) {
			timeline.Show()
		} else {
			timeline.Hide()
		}
	}

	actions := container.NewGridWithColumns(3, addEvidence, reviewSource, createTask, askEco, warranty, timeline)
	content := container.NewVBox(
		title,
		searchLabel,
		search,
		actions,
		transcriptTitle,
		transcript,
	)
	w.SetContent(container.NewPadded(content))
	w.ShowAndRun()
}
