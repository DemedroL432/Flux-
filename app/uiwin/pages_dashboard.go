//go:build windows

package uiwin

import (
	. "github.com/lxn/walk/declarative"
)

func (ui *AppUI) pageDashboard() TabPage {
	return TabPage{
		Title:  "Dashboard",
		Layout: VBox{},
		Children: []Widget{
			GroupBox{
				Title:  "Quick actions",
				Layout: HBox{},
				Children: []Widget{
					PushButton{
						Text: "Apply proxy/routing",
						OnClicked: func() {
							ui.applyProxyAndRouting()
						},
					},
					PushButton{
						Text: "Start DPI",
						OnClicked: func() {
							ui.startDPI()
						},
					},
					PushButton{
						Text: "Stop DPI",
						OnClicked: func() {
							ui.stopDPI()
						},
					},
				},
			},
		},
	}
}

