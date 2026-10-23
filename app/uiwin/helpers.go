//go:build windows

package uiwin

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/lxn/walk"
	"pulseshield/app/models"
)

func (ui *AppUI) showErr(err error) {
	if err == nil {
		return
	}
	walk.MsgBox(ui.mw, "Error", err.Error(), walk.MsgBoxIconError)
}

func serverNames(cfg models.AppConfig) []string {
	out := make([]string, 0, len(cfg.Servers))
	for _, s := range cfg.Servers {
		name := s.Name
		if cfg.ActiveServerID == s.ID {
			name += " (active)"
		}
		out = append(out, name)
	}
	return out
}

func setComboByText(cb *walk.ComboBox, value string) {
	m := cb.Model()
	items, ok := m.([]string)
	if !ok {
		return
	}
	for i, it := range items {
		if it == value {
			cb.SetCurrentIndex(i)
			return
		}
	}
	if len(items) > 0 {
		cb.SetCurrentIndex(0)
	}
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
		b.WriteString("Rules (top-down):\r\n")
		for i, r := range p.Rules {
			b.WriteString(fmt.Sprintf("%d) %s => %s\r\n", i+1, r.Name, r.Action))
		}
		return b.String()
	}
	return "No routing profile selected"
}

func findRoutingProfile(cfg models.AppConfig, name string) (models.RoutingProfile, bool) {
	for _, p := range cfg.RoutingProfiles {
		if p.Name == name {
			return p, true
		}
	}
	return models.RoutingProfile{}, false
}

func mustPrettyJSON(v any) string {
	raw, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return ""
	}
	return string(raw)
}

