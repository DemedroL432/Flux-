//go:build windows

package uiwin

import (
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"pulseshield/app/diagnostics"
)

func (ui *AppUI) pageDiagnostics() TabPage {
	return TabPage{
		Title:  "Diagnostics",
		Layout: VBox{},
		Children: []Widget{
			GroupBox{
				Title:  "Tools",
				Layout: HBox{},
				Children: []Widget{
					PushButton{
						Text: "Run self-check",
						OnClicked: func() {
							rep, err := diagnostics.Run()
							if err != nil {
								ui.showErr(err)
								return
							}
							snap := ui.st.Snapshot()
							ui.diagTE.SetText(fmt.Sprintf(
								"DataDir: %s\r\nConfig: %s\r\nLogs: %s\r\nBundled core: %v\r\nBundled dpi folder: %v\r\nCore status: %s\r\nProxy enabled: %v\r\nDPI enabled: %v\r\nDPI running: %v\r\nActive route: %s\r\nCore log: %s\r\nDPI log: %s\r\n",
								rep.DataDir, rep.ConfigPath, rep.LogsDir, rep.HasCore, rep.HasDpiFolder, ui.coreManager.Status(), snap.ProxyEnabled, snap.DPIEnabled, ui.dpiRunner.IsRunning(), snap.RoutingProfile, ui.coreManager.LastLogPath(), ui.dpiRunner.LastLogPath(),
							))
						},
					},
					PushButton{
						Text: "Open logs folder",
						OnClicked: func() {
							rep, err := diagnostics.Run()
							if err != nil {
								ui.showErr(err)
								return
							}
							_ = exec.Command("explorer.exe", rep.LogsDir).Start()
						},
					},
					PushButton{
						Text: "Test via local proxy",
						OnClicked: func() {
							client := &http.Client{
								Timeout: 6 * time.Second,
								Transport: &http.Transport{
									Proxy: http.ProxyURL(&url.URL{Scheme: "http", Host: "127.0.0.1:8080"}),
								},
							}
							resp, err := client.Get("https://example.com/")
							if err != nil {
								ui.showErr(err)
								return
							}
							_ = resp.Body.Close()
							walk.MsgBox(ui.mw, "OK", "HTTP request through 127.0.0.1:8080 succeeded", walk.MsgBoxIconInformation)
						},
					},
					PushButton{
						Text: "Export diagnostics bundle (zip)",
						OnClicked: func() {
							res, err := diagnostics.CreateBundle()
							if err != nil {
								ui.showErr(err)
								return
							}
							ui.diagTE.SetText(ui.diagTE.Text() + "\r\nBundle: " + res.ZipPath + "\r\n")
						},
					},
				},
			},
			TextEdit{AssignTo: &ui.diagTE, ReadOnly: true, VScroll: true},
		},
	}
}

func (ui *AppUI) viewDiagnosticsWidgets() []Widget {
	return []Widget{
		card("Diagnostics",
			Composite{
				Layout: HBox{Spacing: 10},
				Children: []Widget{
					PushButton{
						Text: "Self-check",
						OnClicked: func() {
							rep, err := diagnostics.Run()
							if err != nil {
								ui.showErr(err)
								return
							}
							ui.diagTE.SetText(fmt.Sprintf(
								"DataDir: %s\r\nConfig: %s\r\nLogs: %s\r\nBundled core: %v\r\nBundled dpi folder: %v\r\nCore log: %s\r\nDPI log: %s\r\n",
								rep.DataDir, rep.ConfigPath, rep.LogsDir, rep.HasCore, rep.HasDpiFolder, ui.coreManager.LastLogPath(), ui.dpiRunner.LastLogPath(),
							))
						},
					},
					PushButton{
						Text: "Open logs",
						OnClicked: func() {
							rep, err := diagnostics.Run()
							if err != nil {
								ui.showErr(err)
								return
							}
							_ = exec.Command("explorer.exe", rep.LogsDir).Start()
						},
					},
					PushButton{
						Text: "Routing guide",
						OnClicked: func() {
							if err := openBundledRoutingGuide(); err != nil {
								ui.showErr(err)
							}
						},
					},
					PushButton{
						Text: "Test proxy",
						OnClicked: func() {
							client := &http.Client{
								Timeout: 6 * time.Second,
								Transport: &http.Transport{
									Proxy: http.ProxyURL(&url.URL{Scheme: "http", Host: "127.0.0.1:8080"}),
								},
							}
							resp, err := client.Get("https://example.com/")
							if err != nil {
								ui.showErr(err)
								return
							}
							_ = resp.Body.Close()
							walk.MsgBox(ui.mw, "OK", "HTTP request through 127.0.0.1:8080 succeeded", walk.MsgBoxIconInformation)
						},
					},
					HSpacer{},
					PushButton{
						Text: "Export bundle",
						OnClicked: func() {
							res, err := diagnostics.CreateBundle()
							if err != nil {
								ui.showErr(err)
								return
							}
							ui.diagTE.SetText(ui.diagTE.Text() + "\r\nBundle: " + res.ZipPath + "\r\n")
						},
					},
				},
			},
			TextEdit{AssignTo: &ui.diagTE, ReadOnly: true, VScroll: true, MinSize: Size{Width: 0, Height: 360}},
		),
	}
}

