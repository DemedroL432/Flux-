//go:build windows

package uiwin

import (
	"fmt"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"pulseshield/app/branding"
	"pulseshield/app/config"
	"pulseshield/app/core"
	"pulseshield/app/dpi"
	"pulseshield/app/models"
	"pulseshield/app/state"
)

type AppUI struct {
	version     string
	st          *state.AppState
	coreManager *core.Manager
	dpiRunner   *dpi.Runner

	mw        *walk.MainWindow
	statusLbl *walk.Label
	subtitleLbl *walk.Label
	heroHintLbl *walk.Label
	connStateLbl *walk.Label
	proxyStateLbl *walk.Label
	routeStateLbl *walk.Label
	latencyNowLbl *walk.Label

	serversLB *walk.ListBox
	navPanelBtn   *walk.PushButton
	navServersBtn *walk.PushButton
	navProxyBtn   *walk.PushButton
	navDPIBtn     *walk.PushButton
	navDiagBtn    *walk.PushButton

	proxyEnabledCB *walk.CheckBox
	proxyModeCB    *walk.ComboBox

	dpiEnabledCB  *walk.CheckBox
	dpiProfileCB  *walk.ComboBox
	dpiStatusLbl  *walk.Label

	routeProfileCB *walk.ComboBox
	routeInfoTE    *walk.TextEdit
	routeEditorTE  *walk.TextEdit
	routeEditorDirty bool

	diagTE *walk.TextEdit
	autoApplyCB *walk.CheckBox

	viewDashboard *walk.Composite
	viewServers   *walk.Composite
	viewProxy     *walk.Composite
	viewDPI       *walk.Composite
	viewDiag      *walk.Composite

	selectedServerIdx int
	isRefreshing      bool
	currentView       int
	autoApply         bool

	latGraph *LatencyGraph
	latGraphWidget *walk.CustomWidget
}

func Run(version string, st *state.AppState, coreManager *core.Manager, dpiRunner *dpi.Runner) error {
	ui := &AppUI{
		version:           version,
		st:                st,
		coreManager:       coreManager,
		dpiRunner:         dpiRunner,
		selectedServerIdx: -1,
		currentView:       0,
		autoApply:         true,
		latGraph:          NewLatencyGraph(48),
	}

	if err := ui.build(); err != nil {
		return err
	}
	if ui.latGraph != nil && ui.latGraphWidget != nil {
		ui.latGraph.Attach(ui.latGraphWidget)
	}
	ui.refreshAll()
	ui.mw.Run()
	return nil
}

func (ui *AppUI) build() error {
	return MainWindow{
		AssignTo: &ui.mw,
		Title:    fmt.Sprintf("%s (%s)", branding.AppNameUI, ui.version),
		MinSize:  Size{Width: 1040, Height: 720},
		Background: SolidColorBrush{Color: rgbBg()},
		Layout: HBox{Margins: Margins{Left: 0, Top: 0, Right: 0, Bottom: 0}, Spacing: 0},
		MenuItems: []MenuItem{
			Menu{
				Text: "&Help",
				Items: []MenuItem{
					Action{
						Text: "About",
						OnTriggered: func() {
							walk.MsgBox(
								ui.mw,
								branding.AppNameUI,
								fmt.Sprintf("%s\r\n\r\nDeveloper: %s\r\nTesters: %s\r\nVersion: %s", branding.AppNameUI, branding.Developer, branding.Testers, ui.version),
								walk.MsgBoxIconInformation,
							)
						},
					},
				},
			},
		},
		Children: []Widget{
			ui.sidebar(),
			ui.content(),
		},
	}.Create()
}

func (ui *AppUI) sidebar() Composite {
	return Composite{
		Background: SolidColorBrush{Color: rgbSidebar()},
		MinSize:    Size{Width: 220, Height: 0},
		MaxSize:    Size{Width: 260, Height: 0},
		Layout:     VBox{Margins: Margins{Left: 14, Top: 18, Right: 14, Bottom: 16}, Spacing: 10},
		Children: []Widget{
			Label{
				Text:      branding.AppNameUI + " // Control",
				TextColor: rgbText(),
				Font:      Font{Family: "Segoe UI Semibold", PointSize: 12, Bold: true},
			},
			Label{
				Text:      "PulseShield navigation",
				TextColor: rgbTextMuted(),
				Font:      Font{Family: "Segoe UI", PointSize: 9},
			},
			PushButton{AssignTo: &ui.navPanelBtn, Text: "Panel", OnClicked: func() { ui.showView(0) }, Background: SolidColorBrush{Color: rgbCard()}},
			PushButton{AssignTo: &ui.navServersBtn, Text: "Servers", OnClicked: func() { ui.showView(1) }, Background: SolidColorBrush{Color: rgbCard()}},
			PushButton{AssignTo: &ui.navProxyBtn, Text: "Network", OnClicked: func() { ui.showView(2) }, Background: SolidColorBrush{Color: rgbCard()}},
			PushButton{AssignTo: &ui.navDPIBtn, Text: "Zapret DPI", OnClicked: func() { ui.showView(3) }, Background: SolidColorBrush{Color: rgbCard()}},
			PushButton{AssignTo: &ui.navDiagBtn, Text: "Diagnostics", OnClicked: func() { ui.showView(4) }, Background: SolidColorBrush{Color: rgbCard()}},
			VSpacer{},
			Composite{
				Background: SolidColorBrush{Color: rgbCardAlt()},
				Layout:     VBox{Margins: Margins{Left: 10, Top: 8, Right: 10, Bottom: 10}, Spacing: 6},
				Children: []Widget{
					Label{Text: "Team", TextColor: rgbAccent(), Font: Font{Family: "Segoe UI", PointSize: 9, Bold: true}},
					Label{
						Text:      "Developer: " + branding.Developer,
						TextColor: rgbTextMuted(),
						Font:      Font{Family: "Segoe UI", PointSize: 8},
					},
					Label{
						Text:      "QA: " + branding.Testers,
						TextColor: rgbTextMuted(),
						Font:      Font{Family: "Segoe UI", PointSize: 8},
					},
				},
			},
			Label{
				Text:      "Build " + ui.version,
				TextColor: rgbTextMuted(),
				Font:      Font{Family: "Consolas", PointSize: 8},
			},
		},
	}
}

func (ui *AppUI) content() Composite {
	return Composite{
		Layout: VBox{Margins: Margins{Left: 18, Top: 16, Right: 18, Bottom: 16}, Spacing: 12},
		Children: []Widget{
			Composite{
				Background: SolidColorBrush{Color: rgbCardAlt()},
				Layout:     VBox{Margins: Margins{Left: 16, Top: 14, Right: 16, Bottom: 14}, Spacing: 6},
				Children: []Widget{
					Label{Text: "PulseShield Command Center", TextColor: rgbText(), Font: Font{Family: "Segoe UI Semibold", PointSize: 16, Bold: true}},
					Label{AssignTo: &ui.heroHintLbl, Text: "Adaptive route + DPI engine with single-click control", TextColor: rgbAccent(), Font: Font{Family: "Segoe UI", PointSize: 9}},
					Label{AssignTo: &ui.subtitleLbl, Text: "Active server: —", TextColor: rgbTextMuted(), Font: Font{Family: "Segoe UI", PointSize: 9}},
					Label{AssignTo: &ui.statusLbl, Text: "status: idle", TextColor: rgbTextMuted(), Font: Font{Family: "Consolas", PointSize: 9}},
				},
			},
			Composite{
				Layout: VBox{Margins: Margins{Left: 0, Top: 0, Right: 0, Bottom: 0}, Spacing: 12},
				Children: []Widget{
					Composite{AssignTo: &ui.viewDashboard, Visible: true, Layout: VBox{}, Children: ui.viewPanel()},
					Composite{AssignTo: &ui.viewServers, Visible: false, Layout: VBox{}, Children: ui.viewServersWidgets()},
					Composite{AssignTo: &ui.viewProxy, Visible: false, Layout: VBox{}, Children: ui.viewProxyWidgets()},
					Composite{AssignTo: &ui.viewDPI, Visible: false, Layout: VBox{}, Children: ui.viewDPIWidgets()},
					Composite{AssignTo: &ui.viewDiag, Visible: false, Layout: VBox{}, Children: ui.viewDiagnosticsWidgets()},
				},
			},
		},
	}
}

func (ui *AppUI) showView(idx int) {
	ui.currentView = idx
	views := []*walk.Composite{ui.viewDashboard, ui.viewServers, ui.viewProxy, ui.viewDPI, ui.viewDiag}
	for i, v := range views {
		if v == nil {
			continue
		}
		v.SetVisible(i == idx)
	}
	ui.refreshAll()
}

func (ui *AppUI) refreshNavButtons() {
	buttons := []*walk.PushButton{ui.navPanelBtn, ui.navServersBtn, ui.navProxyBtn, ui.navDPIBtn, ui.navDiagBtn}
	labels := []string{"[CTRL] Panel", "[NODES] Servers", "[NET] Network", "[SHIELD] Zapret DPI", "[LOG] Diagnostics"}
	for i, b := range buttons {
		if b == nil {
			continue
		}
		prefix := "  "
		if i == ui.currentView {
			prefix = "> "
		}
		b.SetText(prefix + labels[i])
	}
}

func (ui *AppUI) refreshAll() {
	ui.isRefreshing = true
	defer func() { ui.isRefreshing = false }()

	snap := ui.st.Snapshot()
	if ui.statusLbl != nil {
		ui.statusLbl.SetText("status: " + ui.coreManager.Status())
	}
	if ui.subtitleLbl != nil {
		active := "—"
		for _, s := range snap.Servers {
			if s.ID == snap.ActiveServerID {
				active = s.Name
				break
			}
		}
		ui.subtitleLbl.SetText("Active server: " + active)
	}
	if ui.connStateLbl != nil {
		if strings.Contains(strings.ToLower(ui.coreManager.Status()), "running") {
			ui.connStateLbl.SetText("Connected")
		} else {
			ui.connStateLbl.SetText("Disconnected")
		}
	}
	if ui.proxyStateLbl != nil {
		if snap.ProxyEnabled {
			ui.proxyStateLbl.SetText("Proxy: enabled")
		} else {
			ui.proxyStateLbl.SetText("Proxy: disabled")
		}
	}
	if ui.routeStateLbl != nil {
		ui.routeStateLbl.SetText("Routing: " + snap.RoutingProfile)
	}
	if ui.latencyNowLbl != nil && ui.latGraph != nil {
		v := ui.latGraph.Latest()
		if v <= 0 {
			ui.latencyNowLbl.SetText("Latency: -- ms")
		} else {
			ui.latencyNowLbl.SetText(fmt.Sprintf("Latency: %.0f ms", v))
		}
	}

	// Latency graph: if proxy enabled, measure through local HTTP proxy.
	if ui.latGraph != nil {
		if snap.ProxyEnabled {
			ui.latGraph.EnsureRunning(fmt.Sprintf("http://127.0.0.1:%d", snap.ListenPort()))
		} else {
			ui.latGraph.EnsureRunning("")
		}
	}

	if ui.proxyEnabledCB != nil {
		ui.proxyEnabledCB.SetChecked(snap.ProxyEnabled)
	}
	if ui.proxyModeCB != nil {
		_ = ui.proxyModeCB.SetModel([]string{
			string(models.ProxyModeSystem),
			string(models.ProxyModeTun),
		})
		setComboByText(ui.proxyModeCB, string(snap.ProxyMode))
	}

	if ui.dpiEnabledCB != nil {
		ui.dpiEnabledCB.SetChecked(snap.DPIEnabled)
	}
	if ui.dpiProfileCB != nil {
		_ = ui.dpiProfileCB.SetModel(profileNames(snap.DPIProfiles))
		setComboByText(ui.dpiProfileCB, snap.ActiveDPIProfile)
	}
	if ui.dpiStatusLbl != nil {
		if ui.dpiRunner.IsRunning() {
			ui.dpiStatusLbl.SetText("DPI status: running")
		} else {
			ui.dpiStatusLbl.SetText("DPI status: stopped")
		}
	}

	if ui.routeProfileCB != nil {
		_ = ui.routeProfileCB.SetModel(profileNamesR(snap.RoutingProfiles))
		setComboByText(ui.routeProfileCB, snap.RoutingProfile)
	}
	if ui.routeInfoTE != nil {
		ui.routeInfoTE.SetText(formatRouting(snap))
	}
	if ui.routeEditorTE != nil {
		if !ui.routeEditorDirty {
			if p, ok := findRoutingProfile(snap, snap.RoutingProfile); ok {
				ui.routeEditorTE.SetText(mustPrettyJSON(p))
			}
		}
	}
	if ui.autoApplyCB != nil {
		ui.autoApplyCB.SetChecked(ui.autoApply)
	}

	if ui.serversLB != nil {
		_ = ui.serversLB.SetModel(serverNames(snap))
	}
	ui.refreshNavButtons()
}

func (ui *AppUI) applyRuntimeState() {
	snap := ui.st.Snapshot()
	if snap.DPIEnabled {
		if prof, ok := findDPIProfile(snap, snap.ActiveDPIProfile); ok {
			if err := ui.dpiRunner.Start(prof, &snap); err != nil {
				ui.showErr(err)
			}
		}
	} else {
		_ = ui.dpiRunner.Stop()
	}
	if err := ui.coreManager.Apply(snap); err != nil {
		ui.showErr(err)
	}
	ui.refreshAll()
}

func (ui *AppUI) stopAllNow() {
	ui.st.Update(func(cfg *models.AppConfig) {
		cfg.ProxyEnabled = false
		cfg.DPIEnabled = false
	})
	_ = config.Save(ui.st.Snapshot())
	_ = ui.dpiRunner.Stop()
	if err := ui.coreManager.Apply(ui.st.Snapshot()); err != nil {
		ui.showErr(err)
	}
	ui.refreshAll()
}

func (ui *AppUI) applyProxyAndRouting() {
	ui.applyRuntimeState()
}

func (ui *AppUI) startDPI() {
	snap := ui.st.Snapshot()
	if prof, ok := findDPIProfile(snap, snap.ActiveDPIProfile); ok {
		if err := ui.dpiRunner.Start(prof, &snap); err != nil {
			ui.showErr(err)
		}
	}
	ui.refreshAll()
}

func (ui *AppUI) stopDPI() {
	_ = ui.dpiRunner.Stop()
	ui.refreshAll()
}

func (ui *AppUI) save(mutator func(cfg *models.AppConfig)) {
	ui.st.Update(mutator)
	_ = config.Save(ui.st.Snapshot())
	if ui.autoApply {
		ui.applyRuntimeState()
		return
	}
	ui.refreshAll()
}

