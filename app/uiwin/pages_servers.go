//go:build windows

package uiwin

import (
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"pulseshield/app/models"
)

func (ui *AppUI) pageServers() TabPage {
	return TabPage{
		Title:  "Servers",
		Layout: VBox{},
		Children: []Widget{
			GroupBox{
				Title:  "Manage",
				Layout: HBox{},
				Children: []Widget{
					PushButton{
						Text: "Add from clipboard (VLESS)",
						OnClicked: func() {
							raw, err := walk.Clipboard().Text()
							if err != nil {
								ui.showErr(err)
								return
							}
							server, err := parseVLESS(strings.TrimSpace(raw))
							if err != nil {
								ui.showErr(err)
								return
							}
							ui.save(func(cfg *models.AppConfig) {
								cfg.Servers = append(cfg.Servers, server)
								if cfg.ActiveServerID == "" {
									cfg.ActiveServerID = server.ID
								}
							})
						},
					},
					PushButton{
						Text: "Set selected active",
						OnClicked: func() {
							snap := ui.st.Snapshot()
							if ui.selectedServerIdx < 0 || ui.selectedServerIdx >= len(snap.Servers) {
								return
							}
							ui.save(func(cfg *models.AppConfig) {
								cfg.ActiveServerID = cfg.Servers[ui.selectedServerIdx].ID
							})
						},
					},
				},
			},
			ListBox{
				AssignTo: &ui.serversLB,
				OnCurrentIndexChanged: func() {
					ui.selectedServerIdx = ui.serversLB.CurrentIndex()
				},
			},
		},
	}
}

func (ui *AppUI) viewServersWidgets() []Widget {
	return []Widget{
		card("Servers",
			Composite{
				Layout: HBox{Spacing: 10},
				Children: []Widget{
					PushButton{
						Text: "Add from clipboard (VLESS)",
						OnClicked: func() {
							raw, err := walk.Clipboard().Text()
							if err != nil {
								ui.showErr(err)
								return
							}
							server, err := parseVLESS(strings.TrimSpace(raw))
							if err != nil {
								ui.showErr(err)
								return
							}
							ui.save(func(cfg *models.AppConfig) {
								cfg.Servers = append(cfg.Servers, server)
								if cfg.ActiveServerID == "" {
									cfg.ActiveServerID = server.ID
								}
							})
						},
					},
					PushButton{
						Text: "Set selected active",
						OnClicked: func() {
							snap := ui.st.Snapshot()
							if ui.selectedServerIdx < 0 || ui.selectedServerIdx >= len(snap.Servers) {
								return
							}
							ui.save(func(cfg *models.AppConfig) {
								cfg.ActiveServerID = cfg.Servers[ui.selectedServerIdx].ID
							})
						},
					},
				},
			},
			ListBox{
				AssignTo: &ui.serversLB,
				OnCurrentIndexChanged: func() {
					ui.selectedServerIdx = ui.serversLB.CurrentIndex()
				},
			},
		),
	}
}

