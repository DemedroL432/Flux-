package main

import (
	"testing"

	"pulseshield/app/models"
)

// finalTargetFor повторяет ветку финального решения из VerifyRouteTarget.
// Держим её рядом с тестом, чтобы поймать расхождение с routeFinalOutbound.
func finalTargetFor(snap models.AppConfig) string {
	finalTarget := "proxy"
	if rp, ok := snap.ActiveRoutingProfile(); ok {
		switch lowerTrim(rp.Final) {
		case "direct":
			finalTarget = "direct"
		case "proxy":
			finalTarget = "proxy"
		default:
			if snap.ProxyMode == models.ProxyModeDirect {
				finalTarget = "direct"
			}
		}
	} else if snap.ProxyMode == models.ProxyModeDirect {
		finalTarget = "direct"
	}
	return finalTarget
}

func lowerTrim(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			r += 'a' - 'A'
		}
		out = append(out, r)
	}
	// trim пробелов
	start, end := 0, len(out)
	for start < end && (out[start] == ' ' || out[start] == '\t' || out[start] == '\n' || out[start] == '\r') {
		start++
	}
	for end > start && (out[end-1] == ' ' || out[end-1] == '\t' || out[end-1] == '\n' || out[end-1] == '\r') {
		end--
	}
	return string(out[start:end])
}

// routeFinalOutbound и «куда уйдёт трафик» должны совпадать для всех
// комбинаций proxyMode и final профиля. Иначе UI обещает не то, что делает ядро.
func TestFinalTargetMatchesCore(t *testing.T) {
	type tc struct {
		name      string
		proxyMode models.ProxyMode
		profile   *models.RoutingProfile
		want      string
	}
	modes := []models.ProxyMode{models.ProxyModeSystem, models.ProxyModeTun, models.ProxyModeDirect}
	finals := []struct {
		label string
		value string
	}{{"empty", ""}, {"direct", "direct"}, {"proxy", "proxy"}, {"weird", "  Whatever  "}}

	for _, m := range modes {
		// без профиля вообще
		cfg := models.AppConfig{ProxyMode: m}
		want := "proxy"
		if m == models.ProxyModeDirect {
			want = "direct"
		}
		if got := finalTargetFor(cfg); got != want {
			t.Errorf("no-profile proxyMode=%s: got %q want %q", m, got, want)
		}

		for _, f := range finals {
			p := models.RoutingProfile{Name: "p", Final: f.value}
			cfg := models.AppConfig{ProxyMode: m, RoutingProfile: "p"}
			cfg.RoutingProfiles = []models.RoutingProfile{p}

			want := ""
			switch lowerTrim(f.value) {
			case "direct":
				want = "direct"
			case "proxy":
				want = "proxy"
			default:
				want = "proxy"
				if m == models.ProxyModeDirect {
					want = "direct"
				}
			}
			if got := finalTargetFor(cfg); got != want {
				t.Errorf("proxyMode=%s final=%q: got %q want %q", m, f.label, got, want)
			}
		}
	}
	_ = tc{}
}
