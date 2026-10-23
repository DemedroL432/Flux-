//go:build windows

package uiwin

import (
	. "github.com/lxn/walk/declarative"
	"pulseshield/app/models"
)

func (ui *AppUI) pageRouting() TabPage {
	return TabPage{
		Title:  "Routing",
		Layout: VBox{},
		Children: []Widget{
			GroupBox{
				Title:  "Profile",
				Layout: VBox{},
				Children: []Widget{
					ComboBox{
						AssignTo: &ui.routeProfileCB,
						OnCurrentIndexChanged: func() {
							if ui.isRefreshing || ui.routeProfileCB.CurrentIndex() < 0 {
								return
							}
							ui.save(func(cfg *models.AppConfig) { cfg.RoutingProfile = ui.routeProfileCB.Text() })
						},
					},
					PushButton{
						Text: "Apply proxy/routing now",
						OnClicked: func() {
							ui.applyProxyAndRouting()
						},
					},
				},
			},
			GroupBox{
				Title:    "Rules (read-only preview)",
				Layout:   VBox{},
				Children: []Widget{TextEdit{AssignTo: &ui.routeInfoTE, ReadOnly: true, VScroll: true}},
			},
		},
	}
}

func (ui *AppUI) viewRoutingWidgets() []Widget {
	return []Widget{
		card("Routing",
			Composite{
				Layout: HBox{Spacing: 10},
				Children: []Widget{
					Label{Text: "Profile:", TextColor: rgbTextMuted()},
					ComboBox{
						AssignTo: &ui.routeProfileCB,
						OnCurrentIndexChanged: func() {
							if ui.isRefreshing || ui.routeProfileCB.CurrentIndex() < 0 {
								return
							}
							ui.save(func(cfg *models.AppConfig) { cfg.RoutingProfile = ui.routeProfileCB.Text() })
						},
					},
					HSpacer{},
					PushButton{Text: "Apply now", OnClicked: func() { ui.applyProxyAndRouting() }},
					PushButton{Text: "Load JSON", OnClicked: func() { ui.loadRoutingJSONFromProfile() }},
					PushButton{Text: "Apply JSON", OnClicked: func() { ui.applyRoutingJSONToProfile() }},
					PushButton{Text: "Import file", OnClicked: func() { ui.importRoutingJSONFile() }},
					PushButton{Text: "Export file", OnClicked: func() { ui.exportRoutingJSONFile() }},
				},
			},
			TextEdit{AssignTo: &ui.routeInfoTE, ReadOnly: true, VScroll: true, MinSize: Size{Width: 0, Height: 320}},
			TextEdit{
				AssignTo: &ui.routeEditorTE,
				ReadOnly: false,
				VScroll:  true,
				MinSize:  Size{Width: 0, Height: 260},
				OnTextChanged: func() {
					if ui.isRefreshing {
						return
					}
					ui.routeEditorDirty = true
				},
			},
		),
	}
}

