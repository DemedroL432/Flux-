//go:build windows

package uiwin

import (
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/walk"
	"pulseshield/app/models"
)

func card(title string, children ...Widget) Composite {
	return Composite{
		Background: SolidColorBrush{Color: rgbCard()},
		Layout:     VBox{Margins: Margins{Left: 16, Top: 14, Right: 16, Bottom: 16}, Spacing: 10},
		Children: append([]Widget{
			Label{
				Text:      title,
				TextColor: rgbTextMuted(),
				Font:      Font{Family: "Segoe UI", PointSize: 9, Bold: true},
			},
		}, children...),
	}
}

func (ui *AppUI) viewPanel() []Widget {
	return []Widget{
		card("Live overview",
			Composite{
				Layout: HBox{Spacing: 14},
				Children: []Widget{
					Composite{
						Background: SolidColorBrush{Color: rgbCardAlt()},
						Layout: VBox{Margins: Margins{Left: 12, Top: 10, Right: 12, Bottom: 10}, Spacing: 5},
						Children: []Widget{
							Label{Text: "CONNECTION", TextColor: rgbTextMuted(), Font: Font{Family: "Consolas", PointSize: 8}},
							Label{AssignTo: &ui.connStateLbl, Text: "Disconnected", TextColor: rgbText(), Font: Font{Family: "Segoe UI Semibold", PointSize: 11, Bold: true}},
						},
					},
					Composite{
						Background: SolidColorBrush{Color: rgbCardAlt()},
						Layout: VBox{Margins: Margins{Left: 12, Top: 10, Right: 12, Bottom: 10}, Spacing: 5},
						Children: []Widget{
							Label{Text: "PROXY", TextColor: rgbTextMuted(), Font: Font{Family: "Consolas", PointSize: 8}},
							Label{AssignTo: &ui.proxyStateLbl, Text: "Proxy: disabled", TextColor: rgbText(), Font: Font{Family: "Segoe UI Semibold", PointSize: 11, Bold: true}},
						},
					},
					Composite{
						Background: SolidColorBrush{Color: rgbCardAlt()},
						Layout: VBox{Margins: Margins{Left: 12, Top: 10, Right: 12, Bottom: 10}, Spacing: 5},
						Children: []Widget{
							Label{Text: "ROUTING", TextColor: rgbTextMuted(), Font: Font{Family: "Consolas", PointSize: 8}},
							Label{AssignTo: &ui.routeStateLbl, Text: "Routing: default", TextColor: rgbText(), Font: Font{Family: "Segoe UI Semibold", PointSize: 11, Bold: true}},
						},
					},
				},
			},
		),
		card("Latency (ms)",
			Label{AssignTo: &ui.latencyNowLbl, Text: "Latency: -- ms", TextColor: rgbAccent(), Font: Font{Family: "Consolas", PointSize: 10, Bold: true}},
			CustomWidget{
				MinSize: Size{Width: 0, Height: 140},
				PaintMode: PaintBuffered,
				Paint: func(canvas *walk.Canvas, bounds walk.Rectangle) error {
					if ui.latGraph == nil {
						return nil
					}
					return ui.latGraph.Paint(canvas, bounds)
				},
				AssignTo: &ui.latGraphWidget,
			},
		),
		Composite{
			Layout: HBox{Spacing: 12},
			Children: []Widget{
				Composite{
					Layout: VBox{},
					Children: []Widget{
						card("Подключение",
							Label{Text: "Main tunnel controls", TextColor: rgbTextMuted()},
							Composite{
								Layout: HBox{Spacing: 10},
								Children: []Widget{
									CheckBox{AssignTo: &ui.proxyEnabledCB, Text: "Сист. прокси", OnCheckedChanged: func() {
										if ui.isRefreshing {
											return
										}
										ui.save(func(cfg *models.AppConfig) { cfg.ProxyEnabled = ui.proxyEnabledCB.Checked() })
									}},
									ComboBox{AssignTo: &ui.proxyModeCB, MinSize: Size{Width: 160, Height: 0}},
								},
							},
							Composite{
								Layout: HBox{Spacing: 10},
								Children: []Widget{
									PushButton{
										Text: "Apply network state",
										Background: SolidColorBrush{Color: rgbAccent()},
										OnClicked: func() { ui.applyProxyAndRouting() },
									},
									PushButton{
										Text: "Stop All",
										Background: SolidColorBrush{Color: rgbCardAlt()},
										OnClicked: func() { ui.stopAllNow() },
									},
									CheckBox{
										AssignTo: &ui.autoApplyCB,
										Text: "Auto-apply",
										OnCheckedChanged: func() {
											if ui.isRefreshing {
												return
											}
											ui.autoApply = ui.autoApplyCB.Checked()
										},
									},
								},
							},
						),
					},
				},
				Composite{
					Layout: VBox{},
					Children: []Widget{
						card("Маршрутизация",
							Label{Text: "Профиль", TextColor: rgbTextMuted()},
							ComboBox{AssignTo: &ui.routeProfileCB, OnCurrentIndexChanged: func() {
								if ui.isRefreshing || ui.routeProfileCB.CurrentIndex() < 0 {
									return
								}
								ui.save(func(cfg *models.AppConfig) { cfg.RoutingProfile = ui.routeProfileCB.Text() })
							}},
							TextEdit{AssignTo: &ui.routeInfoTE, ReadOnly: true, VScroll: true, MinSize: Size{Width: 0, Height: 130}},
						),
					},
				},
			},
		},
	}
}

