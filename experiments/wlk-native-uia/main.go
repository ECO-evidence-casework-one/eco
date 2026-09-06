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
		MinSize:  Size{520, 360},
		Size:     Size{700, 460},
		Layout:   VBox{},
		Children: []Widget{
			Label{
				Text: "Synthetic ECO accessibility donor probe — no private case data",
				Accessibility: Accessibility{
					Name:        "Synthetic ECO accessibility donor probe",
					Description: "Public FOSS donor qualification only",
					Role:        AccRoleStatictext,
				},
			},
			Label{Text: "Search whole matter"},
			LineEdit{
				AssignTo: &search,
				Text:     "warranty confirmation",
				Accessibility: Accessibility{
					Name:        "Search whole matter",
					Description: "Synthetic search text entry",
					Role:        AccRoleText,
					State:       AccStateFocusable,
				},
			},
			PushButton{
				Text: "Run search",
				Accessibility: Accessibility{
					Name:          "Run search",
					Description:   "Activate the synthetic search",
					DefaultAction: "Press",
					Role:          AccRolePushbutton,
					State:         AccStateFocusable,
				},
				OnClicked: func() {
					status.SetText("Search complete: " + search.Text())
				},
			},
			ListBox{
				AssignTo: &results,
				Model:    model,
				Accessibility: Accessibility{
					Name:        "Evidence results",
					Description: "Synthetic evidence and casework result list",
					Role:        AccRoleList,
					State:       AccStateFocusable | AccStateSelectable,
				},
			},
			Label{
				AssignTo: &status,
				Text:     "Ready",
				Accessibility: Accessibility{
					Name:        "Search status",
					Description: "Synthetic search status",
					Role:        AccRoleStatictext,
				},
			},
		},
	}.Run()); err != nil {
		log.Fatal(err)
	}
}
