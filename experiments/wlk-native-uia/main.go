//go:build windows

package main

import (
	"log"

	"github.com/xackery/wlk/walk"
	. "github.com/xackery/wlk/cpl"
)

type stringList struct {
	walk.ListModelBase
	items []string
}

func (m *stringList) ItemCount() int { return len(m.items) }
func (m *stringList) Value(index int) interface{} { return m.items[index] }

func main() {
	var mw *walk.MainWindow
	var search *walk.LineEdit
	var results *walk.ListBox
	var runSearch *walk.PushButton
	var status *walk.Label

	model := &stringList{items: []string{
		"Warranty confirmation — synthetic evidence",
		"Alex Rowan — synthetic person",
		"Northvale Devices — synthetic organisation",
	}}

	if err := (MainWindow{
		AssignTo: &mw,
		Title:    "ECO native accessibility donor probe",
		MinSize:  Size{Width: 520, Height: 360},
		Size:     Size{Width: 700, Height: 460},
		Layout:   VBox{},
		Children: []Widget{
			Label{
				Text:    "Synthetic ECO accessibility donor probe — no private case data",
				Visible: true,
			},
			Label{Text: "Search whole matter", Visible: true},
			LineEdit{
				AssignTo: &search,
				Text:     "warranty confirmation",
				Visible:  true,
			},
			PushButton{
				AssignTo: &runSearch,
				Text:     "Run search",
				Visible:  true,
				OnClicked: func() {
					status.SetText("Search complete: " + search.Text())
				},
			},
			ListBox{
				AssignTo: &results,
				Model:    model,
				Visible:  true,
			},
			Label{
				AssignTo: &status,
				Text:     "Ready",
				Visible:  true,
			},
		},
	}).Create(); err != nil {
		log.Fatal(err)
	}

	// Let Windows create/show the real native Edit, Button and ListBox first.
	// Only then add human-facing Name/Description metadata. We deliberately do
	// not override native roles/states/default actions, so Windows keeps its own
	// ValuePattern, InvokePattern, SelectionPattern and keyboard-focus behavior.
	mw.Show()
	if err := search.Accessibility().SetName("Search whole matter"); err != nil {
		log.Fatal(err)
	}
	if err := search.Accessibility().SetDescription("Synthetic search text entry"); err != nil {
		log.Fatal(err)
	}
	if err := runSearch.Accessibility().SetName("Run search"); err != nil {
		log.Fatal(err)
	}
	if err := runSearch.Accessibility().SetDescription("Activate the synthetic search"); err != nil {
		log.Fatal(err)
	}
	if err := results.Accessibility().SetName("Evidence results"); err != nil {
		log.Fatal(err)
	}
	if err := results.Accessibility().SetDescription("Synthetic evidence and casework result list"); err != nil {
		log.Fatal(err)
	}
	mw.Run()
}
