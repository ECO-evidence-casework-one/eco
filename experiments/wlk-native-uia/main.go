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
	var status *walk.Label

	model := &stringList{items: []string{
		"Warranty confirmation — synthetic evidence",
		"Alex Rowan — synthetic person",
		"Northvale Devices — synthetic organisation",
	}}

	if _, err := (MainWindow{
		AssignTo: &mw,
		Title:    "ECO native accessibility donor probe",
		Visible:  true,
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
				// Preserve the native EDIT control's own UIA role and patterns.
				// Dynamic Annotation supplies only the human-facing name/detail.
				Accessibility: Accessibility{
					Name:        "Search whole matter",
					Description: "Synthetic search text entry",
				},
			},
			PushButton{
				Text:    "Run search",
				Visible: true,
				// Preserve the native BUTTON provider and InvokePattern.
				Accessibility: Accessibility{
					Name:        "Run search",
					Description: "Activate the synthetic search",
				},
				OnClicked: func() {
					status.SetText("Search complete: " + search.Text())
				},
			},
			ListBox{
				AssignTo: &results,
				Model:    model,
				Visible:  true,
				// Preserve the native LISTBOX provider and SelectionPattern.
				Accessibility: Accessibility{
					Name:        "Evidence results",
					Description: "Synthetic evidence and casework result list",
				},
			},
			Label{
				AssignTo: &status,
				Text:     "Ready",
				Visible:  true,
			},
		},
	}.Run()); err != nil {
		log.Fatal(err)
	}
}
