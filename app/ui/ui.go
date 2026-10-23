//go:build fyneui
// +build fyneui

package ui

import (
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"pulseshield/app/config"
	"pulseshield/app/core"
	"pulseshield/app/diagnostics"
	"pulseshield/app/dpi"
	"pulseshield/app/models"
	"pulseshield/app/state"
)

func DefaultWindowSize() fyne.Size {
	return fyne.NewSize(980, 680)
}

func Build(mainWindow fyne.Window, st *state.AppState, coreManager *core.Manager, dpiRunner *dpi.Runner) fyne.CanvasObject {
	var refreshAll func()
	selectedServerIdx := -1

	title := widget.NewLabelWithStyle("PulseShield", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	status := widget.NewLabel("status: idle")

	dashboard := container.NewVBox(
		title,
		widget.NewSeparator(),
		status,
	)

	serversList := widget.NewList(
		func() int { return len(st.Snapshot().Servers) },
		func() fyne.CanvasObject {
			return container.NewBorder(nil, nil, widget.NewIcon(theme.ComputerIcon()), nil, widget.NewLabel("server"))
		},
		func(i widget.ListItemID, obj fyne.CanvasObject) {
			cfg := st.Snapshot()
			if i < 0 || i >= len(cfg.Servers) {
				return
			}
			item := cfg.Servers[i]
			name := item.Name
			if cfg.ActiveServerID == item.ID {
				name += " (active)"
			}
			obj.(*fyne.Container).Objects[1].(*widget.Label).SetText(name)
		},
	)
	serversList.OnSelected = func(id widget.ListItemID) {
		selectedServerIdx = int(id)
	}

	importBtn := widget.NewButtonWithIcon("Add from clipboard", theme.ContentAddIcon(), func() {
		if mainWindow == nil {
			return
		}
		raw := strings.TrimSpace(mainWindow.Clipboard().Content())
		server, err := parseVLESS(raw)
		if err != nil {
			dialog.ShowError(err, mainWindow)
			return
		}
		st.Update(func(cfg *models.AppConfig) {
			cfg.Servers = append(cfg.Servers, server)
			if cfg.ActiveServerID == "" {
				cfg.ActiveServerID = server.ID
			}
		})
		_ = config.Save(st.Snapshot())
		refreshAll()
	})

	activateBtn := widget.NewButton("Set selected active", func() {
		idx := selectedServerIdx
		if idx < 0 {
			return
		}
		snap := st.Snapshot()
		if idx >= len(snap.Servers) {
			return
		}
		st.Update(func(cfg *models.AppConfig) {
			cfg.ActiveServerID = cfg.Servers[idx].ID
		})
		_ = config.Save(st.Snapshot())
		refreshAll()
	})

	serversTab := container.NewBorder(
		container.NewVBox(importBtn, activateBtn),
		nil, nil, nil,
		serversList,
	)

	proxyMode := widget.NewSelect([]string{
		string(models.ProxyModeSystem),
		string(models.ProxyModeTun),
		string(models.ProxyModeDirect),
	}, nil)
	proxyMode.SetSelected(string(st.Snapshot().ProxyMode))
	proxyMode.OnChanged = func(v string) {
		st.Update(func(cfg *models.AppConfig) {
			cfg.ProxyMode = models.ProxyMode(v)
		})
		_ = config.Save(st.Snapshot())
		refreshAll()
	}

	proxyEnabled := widget.NewCheck("Proxy enabled", func(v bool) {
		st.Update(func(cfg *models.AppConfig) { cfg.ProxyEnabled = v })
		_ = config.Save(st.Snapshot())
		refreshAll()
	})
	proxyEnabled.SetChecked(st.Snapshot().ProxyEnabled)

	proxyTab := container.NewVBox(proxyEnabled, widget.NewLabel("Proxy mode"), proxyMode)

	dpiEnabled := widget.NewCheck("DPI bypass enabled", func(v bool) {
		st.Update(func(cfg *models.AppConfig) { cfg.DPIEnabled = v })
		_ = config.Save(st.Snapshot())
		refreshAll()
	})
	dpiEnabled.SetChecked(st.Snapshot().DPIEnabled)

	dpiSelect := widget.NewSelect(profileNames(st.Snapshot().DPIProfiles), nil)
	dpiSelect.SetSelected(st.Snapshot().ActiveDPIProfile)
	dpiSelect.OnChanged = func(v string) {
		st.Update(func(cfg *models.AppConfig) { cfg.ActiveDPIProfile = v })
		_ = config.Save(st.Snapshot())
		refreshAll()
	}
	dpiTab := container.NewVBox(dpiEnabled, widget.NewLabel("DPI profile"), dpiSelect)

	routeSelect := widget.NewSelect(profileNamesR(st.Snapshot().RoutingProfiles), nil)
	routeSelect.SetSelected(st.Snapshot().RoutingProfile)
	routeSelect.OnChanged = func(v string) {
		st.Update(func(cfg *models.AppConfig) { cfg.RoutingProfile = v })
		_ = config.Save(st.Snapshot())
		refreshAll()
	}
	routeInfo := widget.NewMultiLineEntry()
	routeInfo.Disable()

	routingTab := container.NewVBox(widget.NewLabel("Routing profile"), routeSelect, routeInfo)

	diagInfo := widget.NewMultiLineEntry()
	diagInfo.Disable()
	selfCheckBtn := widget.NewButton("Run self-check", func() {
		rep, err := diagnostics.Run()
		if err != nil {
			dialog.ShowError(err, mainWindow)
			return
		}
		diagInfo.SetText(fmt.Sprintf(
			"DataDir: %s\nConfig: %s\nLogs: %s\nBundled core: %v\nBundled dpi folder: %v\nCore log: %s\nDPI log: %s\n",
			rep.DataDir,
			rep.ConfigPath,
			rep.LogsDir,
			rep.HasCore,
			rep.HasDpiFolder,
			coreManager.LastLogPath(),
			dpiRunner.LastLogPath(),
		))
	})
	openLogsBtn := widget.NewButton("Open logs folder", func() {
		rep, err := diagnostics.Run()
		if err != nil {
			dialog.ShowError(err, mainWindow)
			return
		}
		_ = exec.Command("explorer.exe", rep.LogsDir).Start()
	})
	testConnBtn := widget.NewButton("Test connection (via local proxy)", func() {
		client := &http.Client{
			Timeout: 6 * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyURL(&url.URL{Scheme: "http", Host: "127.0.0.1:8080"}),
			},
		}
		resp, err := client.Get("https://example.com/")
		if err != nil {
			dialog.ShowError(err, mainWindow)
			return
		}
		_ = resp.Body.Close()
		dialog.ShowInformation("OK", "HTTP request through 127.0.0.1:8080 succeeded", mainWindow)
	})
	exportBundleBtn := widget.NewButton("Export diagnostics bundle (zip)", func() {
		res, err := diagnostics.CreateBundle()
		if err != nil {
			dialog.ShowError(err, mainWindow)
			return
		}
		diagInfo.SetText(diagInfo.Text + "\nBundle: " + res.ZipPath + "\n")
	})
	diagnosticsTab := container.NewVBox(selfCheckBtn, openLogsBtn, testConnBtn, exportBundleBtn, diagInfo)

	applyBtn := widget.NewButton("Apply state", func() {
		snap := st.Snapshot()
		// DPI is independent from proxy.
		if snap.DPIEnabled {
			if prof, ok := findDPIProfile(snap, snap.ActiveDPIProfile); ok {
				if err := dpiRunner.Start(prof, &snap); err != nil {
					dialog.ShowError(err, mainWindow)
				}
			}
		} else {
			_ = dpiRunner.Stop()
		}

		if err := coreManager.Apply(st.Snapshot()); err != nil {
			dialog.ShowError(err, mainWindow)
		}
		refreshAll()
	})
	dashboard.Add(applyBtn)

	tabs := container.NewAppTabs(
		container.NewTabItem("Dashboard", dashboard),
		container.NewTabItem("Servers", serversTab),
		container.NewTabItem("Proxy Modes", proxyTab),
		container.NewTabItem("DPI", dpiTab),
		container.NewTabItem("Routing", routingTab),
		container.NewTabItem("Diagnostics", diagnosticsTab),
	)

	refreshAll = func() {
		snap := st.Snapshot()
		status.SetText("status: " + coreManager.Status())
		proxyEnabled.SetChecked(snap.ProxyEnabled)
		proxyMode.SetSelected(string(snap.ProxyMode))
		dpiEnabled.SetChecked(snap.DPIEnabled)
		dpiSelect.Options = profileNames(snap.DPIProfiles)
		dpiSelect.SetSelected(snap.ActiveDPIProfile)
		routeSelect.Options = profileNamesR(snap.RoutingProfiles)
		routeSelect.SetSelected(snap.RoutingProfile)
		routeInfo.SetText(formatRouting(snap))
		serversList.Refresh()
	}
	refreshAll()

	return tabs
}

func parseVLESS(raw string) (models.Server, error) {
	if !strings.HasPrefix(strings.ToLower(raw), "vless://") {
		return models.Server{}, fmt.Errorf("clipboard does not contain a vless link")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return models.Server{}, err
	}
	params := map[string]string{}
	for k, vs := range u.Query() {
		if len(vs) > 0 {
			params[k] = vs[0]
		}
	}
	parts := strings.Split(u.Host, ":")
	host := u.Host
	port := ""
	if len(parts) >= 2 {
		host = parts[0]
		port = parts[1]
	}
	name := u.Fragment
	if name == "" {
		name = host
	}
	return models.Server{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Name:      name,
		RawVLESS:  raw,
		Host:      host,
		Port:      port,
		UUID:      u.User.Username(),
		Params:    params,
		CreatedAt: time.Now(),
	}, nil
}

func profileNames(items []models.DPIProfile) []string {
	out := make([]string, 0, len(items))
	for _, p := range items {
		out = append(out, p.Name)
	}
	return out
}

func profileNamesR(items []models.RoutingProfile) []string {
	out := make([]string, 0, len(items))
	for _, p := range items {
		out = append(out, p.Name)
	}
	return out
}

func findDPIProfile(cfg models.AppConfig, name string) (models.DPIProfile, bool) {
	for _, p := range cfg.DPIProfiles {
		if p.Name == name {
			return p, true
		}
	}
	return models.DPIProfile{}, false
}

func formatRouting(cfg models.AppConfig) string {
	var b strings.Builder
	for _, p := range cfg.RoutingProfiles {
		if p.Name != cfg.RoutingProfile {
			continue
		}
		b.WriteString("Rules (top-down):\n")
		for i, r := range p.Rules {
			b.WriteString(fmt.Sprintf("%d) %s => %s\n", i+1, r.Name, r.Action))
		}
		return b.String()
	}
	return "No routing profile selected"
}
