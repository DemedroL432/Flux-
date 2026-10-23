//go:build windows

package uiwin

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"pulseshield/app/models"
)

func (ui *AppUI) pageProxy() TabPage {
	return TabPage{
		Title:  "Proxy",
		Layout: VBox{},
		Children: []Widget{
			GroupBox{
				Title:  "Mode",
				Layout: VBox{},
				Children: []Widget{
					CheckBox{
						AssignTo: &ui.proxyEnabledCB,
						Text:     "Proxy enabled",
						OnCheckedChanged: func() {
							if ui.isRefreshing {
								return
							}
							ui.save(func(cfg *models.AppConfig) { cfg.ProxyEnabled = ui.proxyEnabledCB.Checked() })
						},
					},
					Composite{
						Layout: HBox{},
						Children: []Widget{
							Label{Text: "Proxy mode:"},
							ComboBox{
								AssignTo: &ui.proxyModeCB,
								OnCurrentIndexChanged: func() {
									if ui.isRefreshing || ui.proxyModeCB.CurrentIndex() < 0 {
										return
									}
									text := ui.proxyModeCB.Text()
									ui.save(func(cfg *models.AppConfig) { cfg.ProxyMode = models.ProxyMode(text) })
								},
							},
						},
					},
					Composite{
						Layout: HBox{},
						Children: []Widget{
							PushButton{
								Text: "Apply now",
								OnClicked: func() {
									ui.applyProxyAndRouting()
								},
							},
						},
					},
				},
			},
		},
	}
}

func (ui *AppUI) viewProxyWidgets() []Widget {
	return []Widget{
		card("Network (Proxy + Routing)",
			Composite{
				Layout: HBox{Spacing: 10},
				Children: []Widget{
					CheckBox{
						AssignTo: &ui.proxyEnabledCB,
						Text:     "Proxy enabled",
						OnCheckedChanged: func() {
							if ui.isRefreshing {
								return
							}
							ui.save(func(cfg *models.AppConfig) { cfg.ProxyEnabled = ui.proxyEnabledCB.Checked() })
						},
					},
					Label{Text: "Transport:", TextColor: rgbTextMuted()},
					ComboBox{AssignTo: &ui.proxyModeCB, OnCurrentIndexChanged: func() {
						if ui.isRefreshing || ui.proxyModeCB.CurrentIndex() < 0 {
							return
						}
						text := ui.proxyModeCB.Text()
						ui.save(func(cfg *models.AppConfig) { cfg.ProxyMode = models.ProxyMode(text) })
					}},
					HSpacer{},
					PushButton{
						Text: "Apply network",
						Background: SolidColorBrush{Color: rgbAccent()},
						OnClicked: func() { ui.applyProxyAndRouting() },
					},
				},
			},
			Composite{
				Layout: HBox{Spacing: 10},
				Children: []Widget{
					Label{Text: "Routing profile:", TextColor: rgbTextMuted()},
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
					PushButton{Text: "Load JSON", OnClicked: func() { ui.loadRoutingJSONFromProfile() }},
					PushButton{Text: "Apply JSON", OnClicked: func() { ui.applyRoutingJSONToProfile() }},
					PushButton{Text: "Import", OnClicked: func() { ui.importRoutingJSONFile() }},
					PushButton{Text: "Export", OnClicked: func() { ui.exportRoutingJSONFile() }},
				},
			},
			Label{Text: "Routing preview (what goes proxy/direct/block):", TextColor: rgbTextMuted()},
			TextEdit{AssignTo: &ui.routeInfoTE, ReadOnly: true, VScroll: true, MinSize: Size{Width: 0, Height: 180}},
			Label{Text: "Routing JSON editor:", TextColor: rgbTextMuted()},
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

func (ui *AppUI) loadRoutingJSONFromProfile() {
	snap := ui.st.Snapshot()
	p, ok := findRoutingProfile(snap, snap.RoutingProfile)
	if !ok {
		ui.showErr(fmt.Errorf("selected routing profile not found"))
		return
	}
	ui.routeEditorDirty = false
	ui.routeEditorTE.SetText(mustPrettyJSON(p))
}

func (ui *AppUI) applyRoutingJSONToProfile() {
	if ui.routeEditorTE == nil {
		return
	}
	raw := ui.routeEditorTE.Text()
	if raw == "" {
		ui.showErr(fmt.Errorf("routing JSON is empty"))
		return
	}

	snap := ui.st.Snapshot()
	name := snap.RoutingProfile
	if name == "" {
		ui.showErr(fmt.Errorf("select routing profile first"))
		return
	}

	var profile models.RoutingProfile
	if err := json.Unmarshal([]byte(raw), &profile); err != nil || len(profile.Rules) == 0 {
		var rules []models.RouteRule
		if err2 := json.Unmarshal([]byte(raw), &rules); err2 != nil {
			ui.showErr(fmt.Errorf("invalid JSON: expected RoutingProfile object or []RouteRule"))
			return
		}
		profile = models.RoutingProfile{Name: name, Rules: rules}
	}
	if profile.Name == "" {
		profile.Name = name
	}

	ui.save(func(cfg *models.AppConfig) {
		replaced := false
		for i := range cfg.RoutingProfiles {
			if cfg.RoutingProfiles[i].Name == name {
				cfg.RoutingProfiles[i] = profile
				replaced = true
				break
			}
		}
		if !replaced {
			cfg.RoutingProfiles = append(cfg.RoutingProfiles, profile)
		}
		cfg.RoutingProfile = profile.Name
	})
	ui.routeEditorDirty = false
}

func (ui *AppUI) importRoutingJSONFile() {
	dlg := walk.FileDialog{
		Title:  "Import routing JSON",
		Filter: "JSON files (*.json)|*.json|All files (*.*)|*.*",
	}
	ok, err := dlg.ShowOpen(ui.mw)
	if err != nil {
		ui.showErr(err)
		return
	}
	if !ok {
		return
	}
	raw, err := os.ReadFile(dlg.FilePath)
	if err != nil {
		ui.showErr(err)
		return
	}
	ui.routeEditorTE.SetText(string(raw))
	ui.routeEditorDirty = true
}

func (ui *AppUI) exportRoutingJSONFile() {
	if ui.routeEditorTE == nil {
		return
	}
	dlg := walk.FileDialog{
		Title:          "Export routing JSON",
		Filter:         "JSON files (*.json)|*.json|All files (*.*)|*.*",
		FilePath:       "routing-profile.json",
		InitialDirPath: "",
	}
	ok, err := dlg.ShowSave(ui.mw)
	if err != nil {
		ui.showErr(err)
		return
	}
	if !ok {
		return
	}
	if err := os.WriteFile(dlg.FilePath, []byte(ui.routeEditorTE.Text()), 0o644); err != nil {
		ui.showErr(err)
	}
}

