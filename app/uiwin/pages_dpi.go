//go:build windows

package uiwin

import (
	"os/exec"

	. "github.com/lxn/walk/declarative"
	"pulseshield/app/models"
)

func (ui *AppUI) pageDPI() TabPage {
	return TabPage{
		Title:  "DPI",
		Layout: VBox{},
		Children: []Widget{
			GroupBox{
				Title:  "Status",
				Layout: VBox{},
				Children: []Widget{
					Label{AssignTo: &ui.dpiStatusLbl, Text: "DPI status: unknown"},
					CheckBox{
						AssignTo: &ui.dpiEnabledCB,
						Text:     "DPI bypass enabled (auto-start on Apply)",
						OnCheckedChanged: func() {
							if ui.isRefreshing {
								return
							}
							ui.save(func(cfg *models.AppConfig) { cfg.DPIEnabled = ui.dpiEnabledCB.Checked() })
						},
					},
				},
			},
			GroupBox{
				Title:  "Profile",
				Layout: VBox{},
				Children: []Widget{
					ComboBox{
						AssignTo: &ui.dpiProfileCB,
						OnCurrentIndexChanged: func() {
							if ui.isRefreshing || ui.dpiProfileCB.CurrentIndex() < 0 {
								return
							}
							ui.save(func(cfg *models.AppConfig) { cfg.ActiveDPIProfile = ui.dpiProfileCB.Text() })
						},
					},
					Composite{
						Layout: HBox{},
						Children: []Widget{
							PushButton{
								Text: "Start",
								OnClicked: func() {
									ui.startDPI()
								},
							},
							PushButton{
								Text: "Stop",
								OnClicked: func() {
									ui.stopDPI()
								},
							},
							PushButton{
								Text: "Open DPI log",
								OnClicked: func() {
									p := ui.dpiRunner.LastLogPath()
									if p == "" {
										return
									}
									_ = exec.Command("notepad.exe", p).Start()
								},
							},
						},
					},
				},
			},
		},
	}
}

func (ui *AppUI) viewDPIWidgets() []Widget {
	return []Widget{
		card("Zapret (DPI)",
			Label{AssignTo: &ui.dpiStatusLbl, Text: "DPI status: unknown", TextColor: rgbTextMuted()},
			Composite{
				Layout: HBox{Spacing: 10},
				Children: []Widget{
					CheckBox{
						AssignTo: &ui.dpiEnabledCB,
						Text:     "Enabled (auto-start on Apply)",
						OnCheckedChanged: func() {
							if ui.isRefreshing {
								return
							}
							ui.save(func(cfg *models.AppConfig) { cfg.DPIEnabled = ui.dpiEnabledCB.Checked() })
						},
					},
					HSpacer{},
					PushButton{
						Text: "Start",
						Background: SolidColorBrush{Color: rgbAccent()},
						OnClicked: func() { ui.startDPI() },
					},
					PushButton{
						Text: "Stop",
						Background: SolidColorBrush{Color: rgbCardAlt()},
						OnClicked: func() { ui.stopDPI() },
					},
					PushButton{
						Text: "Open DPI log",
						Background: SolidColorBrush{Color: rgbCardAlt()},
						OnClicked: func() {
							p := ui.dpiRunner.LastLogPath()
							if p == "" {
								return
							}
							_ = exec.Command("notepad.exe", p).Start()
						},
					},
				},
			},
			ComboBox{
				AssignTo: &ui.dpiProfileCB,
				OnCurrentIndexChanged: func() {
					if ui.isRefreshing || ui.dpiProfileCB.CurrentIndex() < 0 {
						return
					}
					ui.save(func(cfg *models.AppConfig) { cfg.ActiveDPIProfile = ui.dpiProfileCB.Text() })
				},
			},
		),
	}
}

